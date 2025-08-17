package models

import (
	"time"

	"github.com/google/uuid"
)

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
	ArticleID *uuid.UUID `form:"article_id"`
	Limit     int        `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset    int        `form:"offset" binding:"omitempty,min=0"`
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

func (f *ArticleChunkFilter) SetDefaults() {
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}