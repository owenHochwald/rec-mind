package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rec-mind/internal/repository"
	"rec-mind/internal/services"
	"rec-mind/models"
	"rec-mind/pkg/response"
)

func UploadArticle(articleService *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateArticleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format")
			return
		}

		processingMode := c.DefaultQuery("processing", "async")
		ctx := c.Request.Context()

		switch processingMode {
		case "sync":
			result, err := articleService.CreateArticleWithEmbedding(ctx, &req)
			if err != nil {
				response.InternalServerError(c, "Failed to create article")
				return
			}

			data := gin.H{
				"article":         result.Article.ToResponse(),
				"processing_time": result.ProcessingTime.String(),
				"processing_mode": "sync_embedding",
			}

			if result.EmbeddingResult != nil {
				data["embedding_summary"] = gin.H{
					"tokens_used":     result.EmbeddingResult.Summary.TotalTokens,
					"vectors_uploaded": len(result.EmbeddingResult.Uploads),
				}
			}

			response.CreatedWithMessage(c, data, "Article created with embeddings")

		default:
			article, err := articleService.CreateArticleWithAsyncEmbedding(ctx, &req)
			if err != nil {
				response.InternalServerError(c, "Failed to create article")
				return
			}

			data := gin.H{
				"article":         article.ToResponse(),
				"processing_mode": "async_chunking",
			}

			response.CreatedWithMessage(c, data, "Article created successfully. Chunking and embedding generation are processing in the background.")
		}
	}
}

func ListArticles(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter models.ArticleFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			response.BadRequest(c, "Invalid query parameters")
			return
		}

		filter.SetDefaults()
		articles, err := repo.List(c.Request.Context(), &filter)
		if err != nil {
			response.InternalServerError(c, "Failed to fetch articles")
			return
		}

		var articleResponses []map[string]interface{}
		for _, article := range articles {
			articleResponses = append(articleResponses, article.ToResponse())
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		response.Paginated(c, articleResponses, len(articleResponses), page, filter.Limit)
	}
}

func GetArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid article ID")
			return
		}

		article, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			response.NotFound(c, "Article not found")
			return
		}

		response.Success(c, article.ToResponse())
	}
}

func DeleteArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid article ID")
			return
		}

		if err := repo.Delete(c.Request.Context(), id); err != nil {
			response.InternalServerError(c, "Failed to delete article")
			return
		}

		response.SuccessWithMessage(c, nil, "Article deleted successfully")
	}
}