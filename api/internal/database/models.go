package database

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	URL         string    `json:"url" db:"url"`
	Category    string    `json:"category" db:"category"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateArticleRequest struct {
	Title       string    `json:"title" binding:"required,max=500"`
	Content     string    `json:"content" binding:"required"`
	URL         string    `json:"url" binding:"required,url,max=1000"`
	Category    string    `json:"category" binding:"required,max=100"`
	PublishedAt time.Time `json:"published_at" binding:"required"`
}

type UpdateArticleRequest struct {
	Title       *string    `json:"title,omitempty" binding:"omitempty,max=500"`
	Content     *string    `json:"content,omitempty"`
	URL         *string    `json:"url,omitempty" binding:"omitempty,url,max=1000"`
	Category    *string    `json:"category,omitempty" binding:"omitempty,max=100"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type ArticleFilter struct {
	Category    *string    `form:"category"`
	StartDate   *time.Time `form:"start_date"`
	EndDate     *time.Time `form:"end_date"`
	Limit       int        `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset      int        `form:"offset" binding:"omitempty,min=0"`
	SearchTerm  *string    `form:"search"`
}

func (a *Article) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           a.ID,
		"title":        a.Title,
		"content":      a.Content,
		"url":          a.URL,
		"category":     a.Category,
		"published_at": a.PublishedAt,
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
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