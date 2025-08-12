package database

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	URL       string    `json:"url" db:"url"`
	Category  string    `json:"category" db:"category"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ArticleChunk struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ArticleID      uuid.UUID `json:"article_id" db:"article_id"`
	ChunkIndex     int       `json:"chunk_index" db:"chunk_index"`
	Content        string    `json:"content" db:"content"`
	TokenCount     *int      `json:"token_count" db:"token_count"`
	CharacterCount *int      `json:"character_count" db:"character_count"`
	PineconeID     *string   `json:"pinecone_id" db:"pinecone_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type CreateArticleRequest struct {
	Title    string `json:"title" binding:"required,max=500"`
	Content  string `json:"content" binding:"required"`
	URL      string `json:"url" binding:"required,url,max=1000"`
	Category string `json:"category" binding:"required,max=100"`
}

type UpdateArticleRequest struct {
	Title    *string `json:"title,omitempty" binding:"omitempty,max=500"`
	Content  *string `json:"content,omitempty"`
	URL      *string `json:"url,omitempty" binding:"omitempty,url,max=1000"`
	Category *string `json:"category,omitempty" binding:"omitempty,max=100"`
}

type CreateArticleChunkRequest struct {
	ArticleID      uuid.UUID `json:"article_id" binding:"required"`
	ChunkIndex     int       `json:"chunk_index" binding:"required,min=0"`
	Content        string    `json:"content" binding:"required"`
	TokenCount     *int      `json:"token_count,omitempty" binding:"omitempty,min=0"`
	CharacterCount *int      `json:"character_count,omitempty" binding:"omitempty,min=0"`
	PineconeID     *string   `json:"pinecone_id,omitempty"`
}

type UpdateArticleChunkRequest struct {
	Content        *string `json:"content,omitempty"`
	TokenCount     *int    `json:"token_count,omitempty" binding:"omitempty,min=0"`
	CharacterCount *int    `json:"character_count,omitempty" binding:"omitempty,min=0"`
	PineconeID     *string `json:"pinecone_id,omitempty"`
}

type ArticleChunkFilter struct {
	ArticleID  *uuid.UUID `form:"article_id"`
	Limit      int        `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset     int        `form:"offset" binding:"omitempty,min=0"`
}

type ArticleFilter struct {
	Category   *string `form:"category"`
	Limit      int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset     int     `form:"offset" binding:"omitempty,min=0"`
	SearchTerm *string `form:"search"`
}

func (a *Article) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":         a.ID,
		"title":      a.Title,
		"content":    a.Content,
		"url":        a.URL,
		"category":   a.Category,
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

func (a *ArticleChunk) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":              a.ID,
		"article_id":      a.ArticleID,
		"chunk_index":     a.ChunkIndex,
		"content":         a.Content,
		"token_count":     a.TokenCount,
		"character_count": a.CharacterCount,
		"pinecone_id":     a.PineconeID,
		"created_at":      a.CreatedAt,
	}
}

func (f *ArticleFilter) SetDefaults() {
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}

func (f *ArticleChunkFilter) SetDefaults() {
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}

// RAG Worker Models

type RecommendationJob struct {
	JobID           string    `json:"job_id"`
	ArticleID       uuid.UUID `json:"article_id"`
	SessionID       string    `json:"session_id"`
	CreatedAt       time.Time `json:"created_at"`
	CorrelationID   string    `json:"correlation_id"`
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
	SearchID           string              `json:"search_id"`
	SourceChunkID      uuid.UUID           `json:"source_chunk_id"`
	Results            []ChunkSearchResult `json:"results"`
	TotalFound         int                 `json:"total_found"`
	ProcessingTime     float64             `json:"processing_time"`
	ServiceInstanceID  string              `json:"service_instance_id"`
}

type ChunkSearchError struct {
	SearchID          string `json:"search_id"`
	ErrorMessage      string `json:"error_message"`
	ErrorCode         string `json:"error_code"`
	ServiceInstanceID string `json:"service_instance_id"`
}

type ChunkMatch struct {
	ChunkID   uuid.UUID `json:"chunk_id"`
	Score     float64   `json:"score"`
	ChunkIndex int      `json:"chunk_index"`
	ContentPreview string `json:"content_preview"`
}

type ArticleRecommendation struct {
	ArticleID       uuid.UUID    `json:"article_id"`
	Title           string       `json:"title"`
	Category        string       `json:"category"`
	URL             string       `json:"url"`
	HybridScore     float64      `json:"hybrid_score"`
	MaxSimilarity   float64      `json:"max_similarity"`
	AvgSimilarity   float64      `json:"avg_similarity"`
	ChunkMatches    []ChunkMatch `json:"chunk_matches"`
	MatchedChunks   int          `json:"matched_chunks"`
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

// Query-Based Recommendation Models

type QuerySearchJob struct {
	JobID           string    `json:"job_id"`
	Query           string    `json:"query"`
	SessionID       string    `json:"session_id"`
	MaxResults      int       `json:"max_results"`
	ScoreThreshold  float64   `json:"score_threshold"`
	CreatedAt       time.Time `json:"created_at"`
	CorrelationID   string    `json:"correlation_id"`
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
	SearchID           string              `json:"search_id"`
	Query              string              `json:"query"`
	Results            []QuerySearchResult `json:"results"`
	TotalFound         int                 `json:"total_found"`
	ProcessingTime     float64             `json:"processing_time"`
	ServiceInstanceID  string              `json:"service_instance_id"`
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