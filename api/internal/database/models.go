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
}

type UpdateArticleChunkRequest struct {
	Content        *string `json:"content,omitempty"`
	TokenCount     *int    `json:"token_count,omitempty" binding:"omitempty,min=0"`
	CharacterCount *int    `json:"character_count,omitempty" binding:"omitempty,min=0"`
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