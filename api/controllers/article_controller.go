package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
	"github.com/owenHochwald/rec-mind-api/internal/services"
)


// UploadArticle creates a new article with ML embedding generation
func UploadArticle(articleService *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req database.CreateArticleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check processing mode from query parameter
		processingMode := c.DefaultQuery("processing", "async")

		switch processingMode {
		case "sync":
			// Synchronous processing - wait for embedding generation
			result, err := articleService.CreateArticleWithEmbedding(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
				return
			}

			// Return comprehensive result
			response := gin.H{
				"article": result.Article.ToResponse(),
				"processing_time": result.ProcessingTime.String(),
				"embedding_generated": result.EmbeddingResult != nil,
			}

			if result.EmbeddingResult != nil {
				response["embedding_summary"] = gin.H{
					"tokens_used": result.EmbeddingResult.Summary.TotalTokens,
					"processing_time": result.EmbeddingResult.Summary.ProcessingTime,
					"vectors_uploaded": len(result.EmbeddingResult.Uploads),
				}
			}

			if result.Error != "" {
				response["warning"] = result.Error
			}

			c.JSON(http.StatusCreated, response)

		case "async":
			// Asynchronous processing with chunking - return immediately, process in background
			article, err := articleService.CreateArticleWithChunking(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"article": article.ToResponse(),
				"message": "Article created successfully. Chunking and embedding generation are processing in the background.",
				"processing_mode": "async_chunking",
			})

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid processing mode. Use 'sync' or 'async'"})
		}
	}
}

// ListArticles retrieves articles with filtering and pagination
func ListArticles(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter database.ArticleFilter

		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter.SetDefaults()
		articles, err := repo.List(c.Request.Context(), &filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve articles"})
			return
		}

		count, err := repo.Count(c.Request.Context(), &filter)
		if err != nil {
			log.Printf("Failed to get article count: %v", err)
			count = 0
		}

		response := gin.H{
			"articles": articles,
			"total":    count,
			"limit":    filter.Limit,
			"offset":   filter.Offset,
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetArticle retrieves a single article by ID
func GetArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		article, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, article.ToResponse())
	}
}

// UpdateArticle updates an existing article
func UpdateArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		var req database.UpdateArticleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		article, err := repo.Update(c.Request.Context(), id, &req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, article.ToResponse())
	}
}

// DeleteArticle removes an article
func DeleteArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		if err := repo.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
	}
}

// CheckMLHealth checks the health of the ML service
func CheckMLHealth(articleService *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		err := articleService.CheckMLServiceHealth(ctx)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ml_service_healthy": false,
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ml_service_healthy": true,
			"message": "ML service is healthy and ready for embedding generation",
		})
	}
}

