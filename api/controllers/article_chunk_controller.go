package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
)

// CreateArticleChunk creates a new chunk for an article
func CreateArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req database.CreateArticleChunkRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		chunk, err := repo.Create(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article chunk"})
			return
		}

		c.JSON(http.StatusCreated, chunk.ToResponse())
	}
}

// CreateArticleChunksBatch creates multiple chunks for an article in a single request
func CreateArticleChunksBatch(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Chunks []*database.CreateArticleChunkRequest `json:"chunks" binding:"required,dive"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		chunks, err := repo.CreateBatch(c.Request.Context(), req.Chunks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article chunks"})
			return
		}

		response := make([]map[string]interface{}, len(chunks))
		for i, chunk := range chunks {
			response[i] = chunk.ToResponse()
		}

		c.JSON(http.StatusCreated, gin.H{
			"chunks": response,
			"count":  len(chunks),
		})
	}
}

// GetArticleChunk retrieves a single chunk by ID
func GetArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk ID"})
			return
		}

		chunk, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article chunk not found"})
			return
		}

		c.JSON(http.StatusOK, chunk.ToResponse())
	}
}

// GetArticleChunks retrieves all chunks for a specific article
func GetArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleIDStr := c.Param("id")
		articleID, err := uuid.Parse(articleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		chunks, err := repo.GetByArticleID(c.Request.Context(), articleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve article chunks"})
			return
		}

		response := make([]map[string]interface{}, len(chunks))
		for i, chunk := range chunks {
			response[i] = chunk.ToResponse()
		}

		c.JSON(http.StatusOK, gin.H{
			"chunks": response,
			"count":  len(chunks),
		})
	}
}

// GetArticleChunkByIndex retrieves a specific chunk by article ID and chunk index
func GetArticleChunkByIndex(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleIDStr := c.Param("id")
		articleID, err := uuid.Parse(articleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		indexStr := c.Param("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk index"})
			return
		}

		chunk, err := repo.GetByArticleIDAndIndex(c.Request.Context(), articleID, index)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article chunk not found"})
			return
		}

		c.JSON(http.StatusOK, chunk.ToResponse())
	}
}

// ListArticleChunks retrieves chunks with filtering and pagination
func ListArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter database.ArticleChunkFilter

		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter.SetDefaults()
		chunks, err := repo.List(c.Request.Context(), &filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve article chunks"})
			return
		}

		count, err := repo.Count(c.Request.Context(), &filter)
		if err != nil {
			count = 0
		}

		response := make([]map[string]interface{}, len(chunks))
		for i, chunk := range chunks {
			response[i] = chunk.ToResponse()
		}

		c.JSON(http.StatusOK, gin.H{
			"chunks": response,
			"total":  count,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		})
	}
}

// UpdateArticleChunk updates an existing chunk
func UpdateArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk ID"})
			return
		}

		var req database.UpdateArticleChunkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		chunk, err := repo.Update(c.Request.Context(), id, &req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article chunk not found"})
			return
		}

		c.JSON(http.StatusOK, chunk.ToResponse())
	}
}

// DeleteArticleChunk removes a specific chunk
func DeleteArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk ID"})
			return
		}

		if err := repo.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article chunk not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article chunk deleted successfully"})
	}
}

// DeleteArticleChunks removes all chunks for a specific article
func DeleteArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleIDStr := c.Param("id")
		articleID, err := uuid.Parse(articleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		if err := repo.DeleteByArticleID(c.Request.Context(), articleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article chunks"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "All article chunks deleted successfully"})
	}
}