package models

import (
	"time"

	"github.com/google/uuid"
)

type QuerySearchJob struct {
	JobID          string    `json:"job_id"`
	Query          string    `json:"query"`
	SessionID      string    `json:"session_id"`
	MaxResults     int       `json:"max_results"`
	ScoreThreshold float64   `json:"score_threshold"`
	CreatedAt      time.Time `json:"created_at"`
	CorrelationID  string    `json:"correlation_id"`
}

type QuerySearchMessage struct {
	SearchID       string  `json:"search_id"`
	JobID          string  `json:"job_id"`
	Query          string  `json:"query"`
	MaxResults     int     `json:"max_results"`
	ScoreThreshold float64 `json:"score_threshold"`
}

type QuerySearchResult struct {
	ChunkID         string    `json:"chunk_id"`
	SimilarityScore float64   `json:"similarity_score"`
	ArticleID       uuid.UUID `json:"article_id"`
	ChunkIndex      int       `json:"chunk_index"`
	ArticleTitle    string    `json:"article_title"`
	Category        string    `json:"category"`
	ContentPreview  string    `json:"content_preview"`
	URL             string    `json:"url"`
}

type QuerySearchResponse struct {
	SearchID          string              `json:"search_id"`
	Query             string              `json:"query"`
	Results           []QuerySearchResult `json:"results"`
	TotalFound        int                 `json:"total_found"`
	ProcessingTime    float64             `json:"processing_time"`
	ServiceInstanceID string              `json:"service_instance_id"`
}

type QuerySearchError struct {
	SearchID          string `json:"search_id"`
	Query             string `json:"query"`
	ErrorMessage      string `json:"error_message"`
	ErrorCode         string `json:"error_code"`
	ServiceInstanceID string `json:"service_instance_id"`
}

type QueryRecommendationResult struct {
	JobID           string                  `json:"job_id"`
	Query           string                  `json:"query"`
	Recommendations []ArticleRecommendation `json:"recommendations"`
	TotalFound      int                     `json:"total_found"`
	ProcessingTime  string                  `json:"processing_time"`
	Status          string                  `json:"status"`
	Error           string                  `json:"error,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
}