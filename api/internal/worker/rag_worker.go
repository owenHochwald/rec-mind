package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"rec-mind/internal/database"
	"rec-mind/internal/repository"
	"rec-mind/mq"
)

type RAGWorker struct {
	chunkRepo     repository.ArticleChunkRepository
	articleRepo   repository.ArticleRepository
	redisClient   *redis.Client
	channel       *amqp.Channel
	resultChannel chan SearchResultMessage
	timeoutMap    map[string]*SearchTimeout
	timeoutMutex  sync.RWMutex
}

type SearchResultMessage struct {
	Response *database.ChunkSearchResponse
	Error    *database.ChunkSearchError
}

type SearchTimeout struct {
	SearchIDs []string
	Timer     *time.Timer
	JobID     string
}

func NewRAGWorker(chunkRepo repository.ArticleChunkRepository, articleRepo repository.ArticleRepository, redisClient *redis.Client) (*RAGWorker, error) {
	if mq.MQChannel == nil {
		return nil, fmt.Errorf("RabbitMQ channel not initialized")
	}

	worker := &RAGWorker{
		chunkRepo:     chunkRepo,
		articleRepo:   articleRepo,
		redisClient:   redisClient,
		channel:       mq.MQChannel,
		resultChannel: make(chan SearchResultMessage, 100),
		timeoutMap:    make(map[string]*SearchTimeout),
	}

	// Start search results consumer
	go worker.startSearchResultsConsumer()

	return worker, nil
}

func (w *RAGWorker) ProcessRecommendationJob(job database.RecommendationJob) error {
	startTime := time.Now()
	ctx := context.Background()

	log.Printf("🚀 Processing recommendation job %s for article %s", job.JobID, job.ArticleID)

	// 1. Get source article chunks
	chunks, err := w.chunkRepo.GetByArticleID(ctx, job.ArticleID)
	if err != nil {
		return fmt.Errorf("failed to get chunks for article %s: %w", job.ArticleID, err)
	}

	if len(chunks) == 0 {
		log.Printf("⚠️ No chunks found for article %s", job.ArticleID)
		return w.storeErrorResult(job.JobID, job.ArticleID, "No chunks found for source article")
	}

	log.Printf("📝 Found %d chunks for article %s", len(chunks), job.ArticleID)

	// 2. Create and publish chunk search jobs
	searchIDs := make([]string, len(chunks))
	for i, chunk := range chunks {
		searchID := uuid.New().String()
		searchMsg := database.ChunkSearchMessage{
			SearchID:        searchID,
			JobID:           job.JobID,
			ChunkID:         chunk.ID,
			SourceArticleID: job.ArticleID,
			TopK:            5,
			ScoreThreshold:  0.7,
		}

		err = mq.PublishChunkSearch(searchMsg)
		if err != nil {
			log.Printf("❌ Failed to publish chunk search %s: %v", searchID, err)
			return fmt.Errorf("failed to publish search for chunk %s: %w", chunk.ID, err)
		}
		searchIDs[i] = searchID
	}

	log.Printf("📤 Published %d chunk searches for job %s", len(searchIDs), job.JobID)

	// 3. Collect search results with timeout
	timeout := 30 * time.Second
	results := w.collectSearchResults(searchIDs, timeout, job.JobID)

	log.Printf("📥 Collected %d search results for job %s", len(results), job.JobID)

	// 4. Aggregate and rank by article
	recommendations := w.aggregateAndRank(results)

	log.Printf("🏆 Generated %d recommendations for job %s", len(recommendations), job.JobID)

	// 5. Enrich with full article data
	finalResults, err := w.enrichWithArticleData(recommendations)
	if err != nil {
		log.Printf("❌ Failed to enrich results for job %s: %v", job.JobID, err)
		return w.storeErrorResult(job.JobID, job.ArticleID, fmt.Sprintf("Failed to enrich results: %v", err))
	}

	// 6. Store results and notify completion
	processingTime := time.Since(startTime)
	result := database.RecommendationResult{
		JobID:           job.JobID,
		SourceArticleID: job.ArticleID,
		Recommendations: finalResults,
		TotalFound:      len(finalResults),
		ProcessingTime:  processingTime.String(),
		Status:          "completed",
		CreatedAt:       time.Now(),
	}

	err = w.storeResult(ctx, result)
	if err != nil {
		log.Printf("❌ Failed to store results for job %s: %v", job.JobID, err)
		return fmt.Errorf("failed to store results: %w", err)
	}

	log.Printf("✅ Completed recommendation job %s in %v", job.JobID, processingTime)
	return nil
}

func (w *RAGWorker) collectSearchResults(searchIDs []string, timeout time.Duration, jobID string) []SearchResultMessage {
	results := make([]SearchResultMessage, 0, len(searchIDs))
	resultMap := make(map[string]bool)
	
	// Initialize result map
	for _, id := range searchIDs {
		resultMap[id] = false
	}

	// Set up timeout
	w.timeoutMutex.Lock()
	timer := time.NewTimer(timeout)
	w.timeoutMap[jobID] = &SearchTimeout{
		SearchIDs: searchIDs,
		Timer:     timer,
		JobID:     jobID,
	}
	w.timeoutMutex.Unlock()

	defer func() {
		w.timeoutMutex.Lock()
		delete(w.timeoutMap, jobID)
		w.timeoutMutex.Unlock()
		timer.Stop()
	}()

	// Collect results
	for {
		select {
		case result := <-w.resultChannel:
			var searchID string
			if result.Response != nil {
				searchID = result.Response.SearchID
			} else if result.Error != nil {
				searchID = result.Error.SearchID
			}

			// Check if this result belongs to our job
			if found, exists := resultMap[searchID]; exists && !found {
				results = append(results, result)
				resultMap[searchID] = true

				// Check if we have all results
				allReceived := true
				for _, received := range resultMap {
					if !received {
						allReceived = false
						break
					}
				}
				if allReceived {
					return results
				}
			}

		case <-timer.C:
			log.Printf("⏰ Timeout collecting search results for job %s. Got %d/%d results", jobID, len(results), len(searchIDs))
			return results
		}
	}
}

func (w *RAGWorker) aggregateAndRank(results []SearchResultMessage) []database.ArticleRecommendation {
	articleMatches := make(map[uuid.UUID][]database.ChunkMatch)

	// Group results by article
	for _, result := range results {
		if result.Response != nil {
			for _, searchResult := range result.Response.Results {
				chunkMatch := database.ChunkMatch{
					ChunkID:        uuid.MustParse(searchResult.ChunkID),
					Score:          searchResult.SimilarityScore,
					ChunkIndex:     searchResult.ChunkIndex,
					ContentPreview: searchResult.ContentPreview,
				}
				articleMatches[searchResult.ArticleID] = append(articleMatches[searchResult.ArticleID], chunkMatch)
			}
		}
	}

	// Calculate hybrid scores for each article
	recommendations := make([]database.ArticleRecommendation, 0, len(articleMatches))
	for articleID, matches := range articleMatches {
		hybridScore := w.calculateHybridScore(matches)
		maxSim, avgSim := w.calculateSimilarityStats(matches)

		recommendation := database.ArticleRecommendation{
			ArticleID:     articleID,
			HybridScore:   hybridScore,
			MaxSimilarity: maxSim,
			AvgSimilarity: avgSim,
			ChunkMatches:  matches,
			MatchedChunks: len(matches),
		}
		recommendations = append(recommendations, recommendation)
	}

	// Sort by hybrid score (descending)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].HybridScore > recommendations[j].HybridScore
	})

	return recommendations
}

func (w *RAGWorker) calculateHybridScore(articleMatches []database.ChunkMatch) float64 {
	if len(articleMatches) == 0 {
		return 0
	}

	// Extract similarity scores
	scores := make([]float64, len(articleMatches))
	for i, match := range articleMatches {
		scores[i] = match.Score
	}

	// Calculate components
	maxSimilarity := slices.Max(scores)
	avgSimilarity := w.calculateMean(scores)
	chunkCount := float64(len(scores))

	// Hybrid scoring formula
	relevanceScore := (maxSimilarity * 0.6) + (avgSimilarity * 0.4)
	coverageBonus := math.Min(chunkCount/3.0, 0.2)

	return relevanceScore + coverageBonus
}

func (w *RAGWorker) calculateSimilarityStats(matches []database.ChunkMatch) (float64, float64) {
	if len(matches) == 0 {
		return 0, 0
	}

	scores := make([]float64, len(matches))
	for i, match := range matches {
		scores[i] = match.Score
	}

	maxSim := slices.Max(scores)
	avgSim := w.calculateMean(scores)

	return maxSim, avgSim
}

func (w *RAGWorker) calculateMean(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

func (w *RAGWorker) enrichWithArticleData(recommendations []database.ArticleRecommendation) ([]database.ArticleRecommendation, error) {
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

func (w *RAGWorker) storeResult(ctx context.Context, result database.RecommendationResult) error {
	// Store in Redis with TTL
	key := fmt.Sprintf("recommendation_result:%s", result.JobID)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	err = w.redisClient.Set(ctx, key, resultJSON, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store result in Redis: %w", err)
	}

	log.Printf("💾 Stored recommendation result for job %s in Redis", result.JobID)
	return nil
}

func (w *RAGWorker) storeErrorResult(jobID string, articleID uuid.UUID, errorMsg string) error {
	ctx := context.Background()
	result := database.RecommendationResult{
		JobID:           jobID,
		SourceArticleID: articleID,
		Recommendations: []database.ArticleRecommendation{},
		TotalFound:      0,
		ProcessingTime:  "0s",
		Status:          "error",
		Error:           errorMsg,
		CreatedAt:       time.Now(),
	}

	return w.storeResult(ctx, result)
}

func (w *RAGWorker) startSearchResultsConsumer() {
	queue, err := w.channel.QueueDeclare(
		"search_results", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare search_results queue: %v", err)
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
		log.Fatalf("Failed to register search results consumer: %v", err)
	}

	log.Println("📡 Started search results consumer")

	for d := range msgs {
		var resultMsg SearchResultMessage

		// Try to parse as response first
		var response database.ChunkSearchResponse
		if err := json.Unmarshal(d.Body, &response); err == nil && response.SearchID != "" {
			resultMsg.Response = &response
		} else {
			// Try to parse as error
			var errorResp database.ChunkSearchError
			if err := json.Unmarshal(d.Body, &errorResp); err == nil && errorResp.SearchID != "" {
				resultMsg.Error = &errorResp
			} else {
				log.Printf("❌ Failed to parse search result message: %v", err)
				d.Nack(false, false)
				continue
			}
		}

		// Send to result channel (non-blocking)
		select {
		case w.resultChannel <- resultMsg:
		default:
			log.Printf("⚠️ Result channel is full, dropping message")
		}

		d.Ack(false)
	}
}