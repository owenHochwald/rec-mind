package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rec-mind/internal/repository"
	"rec-mind/models"
	"rec-mind/pkg/response"
)

func CreateArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateArticleChunkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format")
			return
		}

		chunk, err := repo.Create(c.Request.Context(), &req)
		if err != nil {
			response.InternalServerError(c, "Failed to create article chunk")
			return
		}

		response.Created(c, chunk.ToResponse())
	}
}

func CreateArticleChunksBatch(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Chunks []*models.CreateArticleChunkRequest `json:"chunks" binding:"required,dive"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format")
			return
		}

		chunks, err := repo.CreateBatch(c.Request.Context(), req.Chunks)
		if err != nil {
			response.InternalServerError(c, "Failed to create article chunks")
			return
		}

		chunkResponses := make([]map[string]interface{}, len(chunks))
		for i, chunk := range chunks {
			chunkResponses[i] = chunk.ToResponse()
		}

		data := gin.H{
			"chunks": chunkResponses,
			"count":  len(chunks),
		}
		response.Created(c, data)
	}
}

func GetArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid chunk ID")
			return
		}

		chunk, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			response.NotFound(c, "Article chunk not found")
			return
		}

		response.Success(c, chunk.ToResponse())
	}
}

func GetArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid article ID")
			return
		}

		chunks, err := repo.GetByArticleID(c.Request.Context(), articleID)
		if err != nil {
			response.InternalServerError(c, "Failed to retrieve article chunks")
			return
		}

		chunkResponses := make([]map[string]interface{}, len(chunks))
		for i, chunk := range chunks {
			chunkResponses[i] = chunk.ToResponse()
		}

		data := gin.H{
			"chunks": chunkResponses,
			"count":  len(chunks),
		}
		response.Success(c, data)
	}
}

func GetArticleChunkByIndex(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid article ID")
			return
		}

		index, err := strconv.Atoi(c.Param("index"))
		if err != nil {
			response.BadRequest(c, "Invalid chunk index")
			return
		}

		chunk, err := repo.GetByArticleIDAndIndex(c.Request.Context(), articleID, index)
		if err != nil {
			response.NotFound(c, "Article chunk not found")
			return
		}

		response.Success(c, chunk.ToResponse())
	}
}

func ListArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter models.ArticleChunkFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			response.BadRequest(c, "Invalid query parameters")
			return
		}

		filter.SetDefaults()
		chunks, err := repo.List(c.Request.Context(), &filter)
		if err != nil {
			response.InternalServerError(c, "Failed to retrieve article chunks")
			return
		}

		count, err := repo.Count(c.Request.Context(), &filter)
		if err != nil {
			count = 0
		}

		chunkResponses := make([]map[string]interface{}, len(chunks))
		for i, chunk := range chunks {
			chunkResponses[i] = chunk.ToResponse()
		}

		data := gin.H{
			"chunks": chunkResponses,
			"total":  count,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		}
		response.Success(c, data)
	}
}

func UpdateArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid chunk ID")
			return
		}

		var req models.UpdateArticleChunkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format")
			return
		}

		chunk, err := repo.Update(c.Request.Context(), id, &req)
		if err != nil {
			response.NotFound(c, "Article chunk not found")
			return
		}

		response.Success(c, chunk.ToResponse())
	}
}

func DeleteArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid chunk ID")
			return
		}

		if err := repo.Delete(c.Request.Context(), id); err != nil {
			response.NotFound(c, "Article chunk not found")
			return
		}

		response.SuccessWithMessage(c, nil, "Article chunk deleted successfully")
	}
}

func DeleteArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid article ID")
			return
		}

		if err := repo.DeleteByArticleID(c.Request.Context(), articleID); err != nil {
			response.InternalServerError(c, "Failed to delete article chunks")
			return
		}

		response.SuccessWithMessage(c, nil, "All article chunks deleted successfully")
	}
}