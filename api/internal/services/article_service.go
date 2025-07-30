package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/mlclient"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
)

// ArticleService handles article processing with ML integration
type ArticleService struct {
	repo     repository.ArticleRepository
	mlClient *mlclient.MLClient
}

// ArticleProcessingResult represents the result of processing an article
type ArticleProcessingResult struct {
	Article         *database.Article                   `json:"article"`
	EmbeddingResult *mlclient.BatchAndUploadResponse    `json:"embedding_result,omitempty"`
	Error           string                              `json:"error,omitempty"`
	ProcessingTime  time.Duration                       `json:"processing_time"`
}

// NewArticleService creates a new article service
func NewArticleService(repo repository.ArticleRepository, mlClient *mlclient.MLClient) *ArticleService {
	return &ArticleService{
		repo:     repo,
		mlClient: mlClient,
	}
}

// CreateArticleWithEmbedding creates an article and generates its embedding
func (s *ArticleService) CreateArticleWithEmbedding(ctx context.Context, req *database.CreateArticleRequest) (*ArticleProcessingResult, error) {
	startTime := time.Now()
	
	result := &ArticleProcessingResult{
		ProcessingTime: 0,
	}

	// Step 1: Create article in database
	log.Printf("Creating article in database: %s", req.Title)
	article, err := s.repo.Create(ctx, req)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create article: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, fmt.Errorf("failed to create article: %w", err)
	}
	
	result.Article = article
	log.Printf("Article created successfully with ID: %s", article.ID)

	// Step 2: Generate embedding and upload to Pinecone
	log.Printf("Generating embedding for article: %s", article.ID)
	embeddingText := mlclient.CreateEmbeddingText(article.Title, article.Content)
	
	embeddingResult, err := s.mlClient.GenerateSingleEmbeddingAndUpload(ctx, article.ID, embeddingText)
	if err != nil {
		// Log the error but don't fail the entire operation
		// The article is already created, we just couldn't generate embeddings
		log.Printf("Failed to generate embedding for article %s: %v", article.ID, err)
		result.Error = fmt.Sprintf("Article created but embedding generation failed: %v", err)
		result.ProcessingTime = time.Since(startTime)
		return result, nil // Return success but with error info
	}

	result.EmbeddingResult = embeddingResult
	result.ProcessingTime = time.Since(startTime)

	log.Printf("Article processing completed successfully for %s in %v", article.ID, result.ProcessingTime)
	return result, nil
}

// CreateArticleWithAsyncEmbedding creates an article and schedules embedding generation asynchronously
func (s *ArticleService) CreateArticleWithAsyncEmbedding(ctx context.Context, req *database.CreateArticleRequest) (*database.Article, error) {
	// Step 1: Create article in database
	article, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	// Step 2: Generate embedding asynchronously (fire and forget)
	go func() {
		// Use a background context with timeout for the async operation
		asyncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		log.Printf("Starting async embedding generation for article: %s", article.ID)
		embeddingText := mlclient.CreateEmbeddingText(article.Title, article.Content)
		
		_, err := s.mlClient.GenerateSingleEmbeddingAndUpload(asyncCtx, article.ID, embeddingText)
		if err != nil {
			log.Printf("Async embedding generation failed for article %s: %v", article.ID, err)
		} else {
			log.Printf("Async embedding generation completed for article: %s", article.ID)
		}
	}()

	return article, nil
}

// ProcessBatchArticles processes multiple articles for embedding generation
func (s *ArticleService) ProcessBatchArticles(ctx context.Context, articles []*database.Article) (*mlclient.BatchAndUploadResponse, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles provided for batch processing")
	}

	// Prepare embedding requests
	embeddingRequests := make([]mlclient.EmbeddingRequest, len(articles))
	for i, article := range articles {
		embeddingRequests[i] = mlclient.EmbeddingRequest{
			ArticleID: article.ID,
			Text:      mlclient.CreateEmbeddingText(article.Title, article.Content),
		}
	}

	log.Printf("Processing batch of %d articles for embedding generation", len(articles))
	
	// Send batch request to ML service
	result, err := s.mlClient.GenerateBatchEmbeddingsAndUpload(ctx, embeddingRequests)
	if err != nil {
		log.Printf("Batch processing failed: %v", err)
		return nil, fmt.Errorf("batch embedding generation failed: %w", err)
	}

	log.Printf("Batch processing completed successfully: %d articles processed", result.Summary.TotalProcessed)
	return result, nil
}

// CheckMLServiceHealth checks if the ML service is available
func (s *ArticleService) CheckMLServiceHealth(ctx context.Context) error {
	return s.mlClient.Health(ctx)
}