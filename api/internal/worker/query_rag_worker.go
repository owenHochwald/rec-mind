package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"rec-mind/models"
	"rec-mind/internal/repository"
	"rec-mind/mq"
)

type QueryRAGWorker struct {
	articleRepo   repository.ArticleRepository
	redisClient   *redis.Client
	channel       *amqp.Channel
	resultChannel chan QuerySearchResultMessage
	timeoutMap    map[string]*QuerySearchTimeout
	timeoutMutex  sync.RWMutex
}

type QuerySearchResultMessage struct {
	Response *models.QuerySearchResponse
	Error    *models.QuerySearchError
}

type QuerySearchTimeout struct {
	SearchID string
	Timer    *time.Timer
	JobID    string
}

func NewQueryRAGWorker(articleRepo repository.ArticleRepository, redisClient *redis.Client) (*QueryRAGWorker, error) {
	if mq.MQChannel == nil {
		return nil, fmt.Errorf("RabbitMQ channel not initialized")
	}

	worker := &QueryRAGWorker{
		articleRepo:   articleRepo,
		redisClient:   redisClient,
		channel:       mq.MQChannel,
		resultChannel: make(chan QuerySearchResultMessage, 100),
		timeoutMap:    make(map[string]*QuerySearchTimeout),
	}

	// Start search results consumer
	go worker.startQuerySearchResultsConsumer()

	return worker, nil
}

func (w *QueryRAGWorker) ProcessQuerySearchJob(job models.QuerySearchJob) error {
	startTime := time.Now()
	ctx := context.Background()

	log.Printf("🔍 Processing query search job %s for query: \"%s\"", job.JobID, job.Query)

	// Create single query search message (no chunking needed)
	searchID := uuid.New().String()
	searchMsg := models.QuerySearchMessage{
		SearchID:       searchID,
		JobID:          job.JobID,
		Query:          job.Query,
		MaxResults:     job.MaxResults,
		ScoreThreshold: job.ScoreThreshold,
	}

	// Publish query search to ML service
	err := mq.PublishQuerySearch(searchMsg)
	if err != nil {
		log.Printf("❌ Failed to publish query search %s: %v", searchID, err)
		return w.storeQueryErrorResult(job.JobID, job.Query, fmt.Sprintf("Failed to publish search: %v", err))
	}

	log.Printf("📤 Published query search %s for job %s", searchID, job.JobID)

	// Wait for search result with timeout
	timeout := 30 * time.Second
	result := w.collectQuerySearchResult(searchID, timeout, job.JobID)

	if result == nil {
		log.Printf("⏰ Timeout waiting for query search result for job %s", job.JobID)
		return w.storeQueryErrorResult(job.JobID, job.Query, "Search timeout - no response from ML service")
	}

	log.Printf("📥 Received query search result for job %s", job.JobID)

	// Process and enrich results
	var recommendations []models.ArticleRecommendation
	if result.Response != nil && len(result.Response.Results) > 0 {
		recommendations = w.processQueryResults(result.Response.Results)

		// Enrich with full article data
		enrichedResults, err := w.enrichWithArticleData(recommendations)
		if err != nil {
			log.Printf("❌ Failed to enrich results for job %s: %v", job.JobID, err)
			return w.storeQueryErrorResult(job.JobID, job.Query, fmt.Sprintf("Failed to enrich results: %v", err))
		}
		recommendations = enrichedResults
	}

	// Store final results
	processingTime := time.Since(startTime)
	queryResult := models.QueryRecommendationResult{
		JobID:           job.JobID,
		Query:           job.Query,
		Recommendations: recommendations,
		TotalFound:      len(recommendations),
		ProcessingTime:  processingTime.String(),
		Status:          "completed",
		CreatedAt:       time.Now(),
	}

	err = w.storeQueryResult(ctx, queryResult)
	if err != nil {
		log.Printf("❌ Failed to store query results for job %s: %v", job.JobID, err)
		return fmt.Errorf("failed to store results: %w", err)
	}

	log.Printf("✅ Completed query search job %s in %v - found %d recommendations", 
		job.JobID, processingTime, len(recommendations))
	return nil
}

func (w *QueryRAGWorker) collectQuerySearchResult(searchID string, timeout time.Duration, jobID string) *QuerySearchResultMessage {
	// Set up timeout
	w.timeoutMutex.Lock()
	timer := time.NewTimer(timeout)
	w.timeoutMap[jobID] = &QuerySearchTimeout{
		SearchID: searchID,
		Timer:    timer,
		JobID:    jobID,
	}
	w.timeoutMutex.Unlock()

	defer func() {
		w.timeoutMutex.Lock()
		delete(w.timeoutMap, jobID)
		w.timeoutMutex.Unlock()
		timer.Stop()
	}()

	// Wait for result
	for {
		select {
		case result := <-w.resultChannel:
			var resultSearchID string
			if result.Response != nil {
				resultSearchID = result.Response.SearchID
			} else if result.Error != nil {
				resultSearchID = result.Error.SearchID
			}

			// Check if this result belongs to our search
			if resultSearchID == searchID {
				return &result
			}

		case <-timer.C:
			log.Printf("⏰ Timeout collecting query search result for job %s", jobID)
			return nil
		}
	}
}

func (w *QueryRAGWorker) processQueryResults(results []models.QuerySearchResult) []models.ArticleRecommendation {
	// Group results by article ID
	articleGroups := make(map[uuid.UUID][]models.QuerySearchResult)
	for _, result := range results {
		articleGroups[result.ArticleID] = append(articleGroups[result.ArticleID], result)
	}

	// Convert to recommendations
	var recommendations []models.ArticleRecommendation
	for articleID, articleResults := range articleGroups {
		// Calculate scores for this article
		scores := make([]float64, len(articleResults))
		chunkMatches := make([]models.ChunkMatch, len(articleResults))

		for i, result := range articleResults {
			scores[i] = result.SimilarityScore
			chunkMatches[i] = models.ChunkMatch{
				ChunkID:        uuid.MustParse(result.ChunkID),
				Score:          result.SimilarityScore,
				ChunkIndex:     result.ChunkIndex,
				ContentPreview: result.ContentPreview,
			}
		}

		// Calculate aggregate scores
		maxSim := w.maxFloat64(scores)
		avgSim := w.meanFloat64(scores)
		
		// For query-based search, use simple average of similarities as hybrid score
		hybridScore := (maxSim*0.7 + avgSim*0.3)

		recommendation := models.ArticleRecommendation{
			ArticleID:     articleID,
			HybridScore:   hybridScore,
			MaxSimilarity: maxSim,
			AvgSimilarity: avgSim,
			ChunkMatches:  chunkMatches,
			MatchedChunks: len(chunkMatches),
			// Title, Category, URL will be filled by enrichWithArticleData
		}
		recommendations = append(recommendations, recommendation)
	}

	// Sort by hybrid score (descending)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].HybridScore > recommendations[j].HybridScore
	})

	return recommendations
}

func (w *QueryRAGWorker) enrichWithArticleData(recommendations []models.ArticleRecommendation) ([]models.ArticleRecommendation, error) {
	for i := range recommendations {
		article, err := w.articleRepo.GetByID(context.Background(), recommendations[i].ArticleID)
		if err != nil {
			log.Printf("⚠️ Failed to get article %s: %v", recommendations[i].ArticleID, err)
			continue
		}

		recommendations[i].Title = article.Title
		recommendations[i].Category = article.Category
		recommendations[i].URL = article.URL
	}

	return recommendations, nil
}

func (w *QueryRAGWorker) storeQueryResult(ctx context.Context, result models.QueryRecommendationResult) error {
	// Store in Redis with TTL
	key := fmt.Sprintf("query_search_result:%s", result.JobID)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	err = w.redisClient.Set(ctx, key, resultJSON, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store result in Redis: %w", err)
	}

	log.Printf("💾 Stored query search result for job %s in Redis", result.JobID)
	return nil
}

func (w *QueryRAGWorker) storeQueryErrorResult(jobID string, query string, errorMsg string) error {
	ctx := context.Background()
	result := models.QueryRecommendationResult{
		JobID:           jobID,
		Query:           query,
		Recommendations: []models.ArticleRecommendation{},
		TotalFound:      0,
		ProcessingTime:  "0s",
		Status:          "error",
		Error:           errorMsg,
		CreatedAt:       time.Now(),
	}

	return w.storeQueryResult(ctx, result)
}

func (w *QueryRAGWorker) startQuerySearchResultsConsumer() {
	queue, err := w.channel.QueueDeclare(
		"query_search_results", // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare query_search_results queue: %v", err)
	}

	msgs, err := w.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Fatalf("Failed to register query search results consumer: %v", err)
	}

	log.Println("📡 Started query search results consumer")

	for d := range msgs {
		var resultMsg QuerySearchResultMessage

		// Try to parse as response first
		var response models.QuerySearchResponse
		if err := json.Unmarshal(d.Body, &response); err == nil && response.SearchID != "" {
			resultMsg.Response = &response
		} else {
			// Try to parse as error
			var errorResp models.QuerySearchError
			if err := json.Unmarshal(d.Body, &errorResp); err == nil && errorResp.SearchID != "" {
				resultMsg.Error = &errorResp
			} else {
				log.Printf("❌ Failed to parse query search result message: %v", err)
				d.Nack(false, false)
				continue
			}
		}

		// Send to result channel (non-blocking)
		select {
		case w.resultChannel <- resultMsg:
		default:
			log.Printf("⚠️ Query result channel is full, dropping message")
		}

		d.Ack(false)
	}
}

// Helper functions
func (w *QueryRAGWorker) maxFloat64(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	max := slice[0]
	for _, v := range slice {
		if v > max {
			max = v
		}
	}
	return max
}

func (w *QueryRAGWorker) meanFloat64(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum / float64(len(slice))
}