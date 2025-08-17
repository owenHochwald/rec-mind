package models

import (
	"time"

	"github.com/google/uuid"
)

type RecommendationJob struct {
	JobID         string    `json:"job_id"`
	ArticleID     uuid.UUID `json:"article_id"`
	SessionID     string    `json:"session_id"`
	CreatedAt     time.Time `json:"created_at"`
	CorrelationID string    `json:"correlation_id"`
}

type ChunkSearchMessage struct {
	SearchID        string    `json:"search_id"`
	JobID           string    `json:"job_id"`
	ChunkID         uuid.UUID `json:"chunk_id"`
	SourceArticleID uuid.UUID `json:"source_article_id"`
	TopK            int       `json:"top_k"`
	ScoreThreshold  float64   `json:"score_threshold"`
}

type ChunkSearchResult struct {
	ChunkID         string    `json:"chunk_id"`
	SimilarityScore float64   `json:"similarity_score"`
	ArticleID       uuid.UUID `json:"article_id"`
	ChunkIndex      int       `json:"chunk_index"`
	ArticleTitle    string    `json:"article_title"`
	Category        string    `json:"category"`
	ContentPreview  string    `json:"content_preview"`
}

type ChunkSearchResponse struct {
	SearchID          string              `json:"search_id"`
	SourceChunkID     uuid.UUID           `json:"source_chunk_id"`
	Results           []ChunkSearchResult `json:"results"`
	TotalFound        int                 `json:"total_found"`
	ProcessingTime    float64             `json:"processing_time"`
	ServiceInstanceID string              `json:"service_instance_id"`
}

type ChunkSearchError struct {
	SearchID          string `json:"search_id"`
	ErrorMessage      string `json:"error_message"`
	ErrorCode         string `json:"error_code"`
	ServiceInstanceID string `json:"service_instance_id"`
}

type ChunkMatch struct {
	ChunkID        uuid.UUID `json:"chunk_id"`
	Score          float64   `json:"score"`
	ChunkIndex     int       `json:"chunk_index"`
	ContentPreview string    `json:"content_preview"`
}

type ArticleRecommendation struct {
	ArticleID     uuid.UUID    `json:"article_id"`
	Title         string       `json:"title"`
	Category      string       `json:"category"`
	URL           string       `json:"url"`
	HybridScore   float64      `json:"hybrid_score"`
	MaxSimilarity float64      `json:"max_similarity"`
	AvgSimilarity float64      `json:"avg_similarity"`
	ChunkMatches  []ChunkMatch `json:"chunk_matches"`
	MatchedChunks int          `json:"matched_chunks"`
}

type RecommendationResult struct {
	JobID           string                  `json:"job_id"`
	SourceArticleID uuid.UUID               `json:"source_article_id"`
	Recommendations []ArticleRecommendation `json:"recommendations"`
	TotalFound      int                     `json:"total_found"`
	ProcessingTime  string                  `json:"processing_time"`
	Status          string                  `json:"status"`
	Error           string                  `json:"error,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
}