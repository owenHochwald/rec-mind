package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rec-mind/internal/database"
	"rec-mind/internal/redis"
	"rec-mind/mq"
)

type RecommendationController struct{}

func NewRecommendationController() *RecommendationController {
	return &RecommendationController{}
}

type CreateRecommendationJobRequest struct {
	ArticleID     uuid.UUID `json:"article_id" binding:"required"`
	SessionID     string    `json:"session_id"`
	CorrelationID string    `json:"correlation_id"`
}

type RecommendationJobResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// CreateRecommendationJob creates a new recommendation job
// @Summary Create recommendation job
// @Description Create an async recommendation job for article similarity search
// @Tags recommendations
// @Accept json
// @Produce json
// @Param job body CreateRecommendationJobRequest true "Recommendation job data"
// @Success 202 {object} RecommendationJobResponse
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/recommendations [post]
func (rc *RecommendationController) CreateRecommendationJob(c *gin.Context) {
	var req CreateRecommendationJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create recommendation job
	job := database.RecommendationJob{
		JobID:         jobID,
		ArticleID:     req.ArticleID,
		SessionID:     req.SessionID,
		CreatedAt:     time.Now(),
		CorrelationID: req.CorrelationID,
	}

	// Publish job to queue
	if err := mq.PublishRecommendationJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue recommendation job",
			"details": err.Error(),
		})
		return
	}

	response := RecommendationJobResponse{
		JobID:     jobID,
		Status:    "queued",
		Message:   "Recommendation job has been queued for processing",
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusAccepted, response)
}

// GetRecommendationJobStatus gets the status and results of a recommendation job
// @Summary Get recommendation job status
// @Description Get the status and results of a recommendation job by job ID
// @Tags recommendations
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} database.RecommendationResult
// @Success 404 {object} object{job_id=string,status=string,message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/jobs/{job_id} [get]
func (rc *RecommendationController) GetRecommendationJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job_id parameter is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check Redis for result
	key := fmt.Sprintf("recommendation_result:%s", jobID)
	resultJSON, err := redis.RedisClient.Get(ctx, key).Result()
	
	if err != nil {
		// Job not found or still processing
		c.JSON(http.StatusNotFound, gin.H{
			"job_id": jobID,
			"status": "processing",
			"message": "Job is still being processed or does not exist",
		})
		return
	}

	// Parse result
	var result database.RecommendationResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse job result",
			"job_id": jobID,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRecommendations gets recommendations for an article (convenience endpoint)
// @Summary Get article recommendations
// @Description Get recommendations for an article using RAG-based similarity search
// @Tags recommendations
// @Produce json
// @Param id path string true "Article ID"
// @Success 200 {object} database.RecommendationResult
// @Success 202 {object} object{job_id=string,status=string,message=string,poll_url=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles/{id}/recommend [get]
func (rc *RecommendationController) GetRecommendations(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid article_id format",
		})
		return
	}

	// Create and submit job
	jobID := uuid.New().String()
	job := database.RecommendationJob{
		JobID:         jobID,
		ArticleID:     articleID,
		SessionID:     c.GetString("session_id"),
		CreatedAt:     time.Now(),
		CorrelationID: c.GetHeader("X-Correlation-ID"),
	}

	if err := mq.PublishRecommendationJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue recommendation job",
			"details": err.Error(),
		})
		return
	}

	// Wait for a short time to see if we get quick results
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := fmt.Sprintf("recommendation_result:%s", jobID)
	
	// Poll for results
	for i := 0; i < 20; i++ {
		resultJSON, err := redis.RedisClient.Get(ctx, key).Result()
		if err == nil {
			// Found result
			var result database.RecommendationResult
			if err := json.Unmarshal([]byte(resultJSON), &result); err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	// Return job ID for async polling
	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"status": "processing",
		"message": "Recommendation job is being processed. Use /api/v1/jobs/{job_id} to check status",
		"poll_url": fmt.Sprintf("/api/v1/jobs/%s", jobID),
	})
}

// HealthCheck for recommendation service
// @Summary Recommendation service health
// @Description Check the health of recommendation service dependencies
// @Tags recommendations
// @Produce json
// @Success 200 {object} object{service=string,status=string,redis_status=string,rabbitmq_status=string}
// @Success 503 {object} object{service=string,status=string,redis_status=string,rabbitmq_status=string}
// @Router /api/v1/recommendations/health [get]
func (rc *RecommendationController) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	health := gin.H{
		"service": "recommendation",
		"status":  "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Check Redis connection
	if err := redis.HealthCheck(ctx); err != nil {
		health["redis_status"] = "unhealthy"
		health["redis_error"] = err.Error()
		health["status"] = "degraded"
	} else {
		health["redis_status"] = "healthy"
	}

	// Check RabbitMQ connection
	if mq.MQChannel == nil || mq.MQChannel.IsClosed() {
		health["rabbitmq_status"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["rabbitmq_status"] = "healthy"
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}