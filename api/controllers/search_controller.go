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

type SearchController struct{}

func NewSearchController() *SearchController {
	return &SearchController{}
}

type QuerySearchRequest struct {
	Query          string  `json:"query" binding:"required,min=1,max=1000"`
	SessionID      string  `json:"session_id"`
	MaxResults     int     `json:"max_results,omitempty"`
	ScoreThreshold float64 `json:"score_threshold,omitempty"`
	CorrelationID  string  `json:"correlation_id,omitempty"`
}

type QuerySearchJobResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	PollURL   string `json:"poll_url"`
	CreatedAt string `json:"created_at"`
}

// SearchByQuery creates a query-based recommendation job
// @Summary Search articles by text query
// @Description Create an async search job to find articles matching a text query using semantic similarity
// @Tags search
// @Accept json
// @Produce json
// @Param query body QuerySearchRequest true "Search query data"
// @Success 202 {object} QuerySearchJobResponse
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/search/recommendations [post]
func (sc *SearchController) SearchByQuery(c *gin.Context) {
	var req QuerySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if req.MaxResults == 0 {
		req.MaxResults = 10
	}
	if req.MaxResults > 50 {
		req.MaxResults = 50 // Cap at 50 results
	}
	if req.ScoreThreshold == 0 {
		req.ScoreThreshold = 0.7
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create query search job
	job := database.QuerySearchJob{
		JobID:          jobID,
		Query:          req.Query,
		SessionID:      req.SessionID,
		MaxResults:     req.MaxResults,
		ScoreThreshold: req.ScoreThreshold,
		CreatedAt:      time.Now(),
		CorrelationID:  req.CorrelationID,
	}

	// Publish job to queue
	if err := mq.PublishQuerySearchJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue search job",
			"details": err.Error(),
		})
		return
	}

	response := QuerySearchJobResponse{
		JobID:     jobID,
		Status:    "queued",
		Message:   "Search job has been queued for processing",
		PollURL:   fmt.Sprintf("/api/v1/search/jobs/%s", jobID),
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusAccepted, response)
}

// GetQuerySearchJobStatus gets the status and results of a query search job
// @Summary Get search job status
// @Description Get the status and results of a query search job by job ID
// @Tags search
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} database.QueryRecommendationResult
// @Success 404 {object} object{job_id=string,status=string,message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/search/jobs/{job_id} [get]
func (sc *SearchController) GetQuerySearchJobStatus(c *gin.Context) {
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
	key := fmt.Sprintf("query_search_result:%s", jobID)
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
	var result database.QueryRecommendationResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse job result",
			"job_id": jobID,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchWithImmediateResponse searches for articles and waits for quick results
// @Summary Search articles with immediate response
// @Description Search for articles matching a text query, with optional polling for faster response
// @Tags search
// @Accept json
// @Produce json
// @Param query body QuerySearchRequest true "Search query data"
// @Success 200 {object} database.QueryRecommendationResult
// @Success 202 {object} QuerySearchJobResponse
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/search/immediate [post]
func (sc *SearchController) SearchWithImmediateResponse(c *gin.Context) {
	var req QuerySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if req.MaxResults == 0 {
		req.MaxResults = 10
	}
	if req.MaxResults > 50 {
		req.MaxResults = 50
	}
	if req.ScoreThreshold == 0 {
		req.ScoreThreshold = 0.7
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create query search job
	job := database.QuerySearchJob{
		JobID:          jobID,
		Query:          req.Query,
		SessionID:      req.SessionID,
		MaxResults:     req.MaxResults,
		ScoreThreshold: req.ScoreThreshold,
		CreatedAt:      time.Now(),
		CorrelationID:  req.CorrelationID,
	}

	// Publish job to queue
	if err := mq.PublishQuerySearchJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue search job",
			"details": err.Error(),
		})
		return
	}

	// Wait for a short time to see if we get quick results
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := fmt.Sprintf("query_search_result:%s", jobID)
	
	// Poll for results
	for i := 0; i < 20; i++ {
		resultJSON, err := redis.RedisClient.Get(ctx, key).Result()
		if err == nil {
			// Found result
			var result database.QueryRecommendationResult
			if err := json.Unmarshal([]byte(resultJSON), &result); err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	// Return job ID for async polling
	response := QuerySearchJobResponse{
		JobID:     jobID,
		Status:    "processing",
		Message:   "Search job is being processed. Use the poll_url to check status",
		PollURL:   fmt.Sprintf("/api/v1/search/jobs/%s", jobID),
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusAccepted, response)
}

// HealthCheck for search service
// @Summary Search service health
// @Description Check the health of search service dependencies
// @Tags search
// @Produce json
// @Success 200 {object} object{service=string,status=string,redis_status=string,rabbitmq_status=string}
// @Success 503 {object} object{service=string,status=string,redis_status=string,rabbitmq_status=string}
// @Router /api/v1/search/health [get]
func (sc *SearchController) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	health := gin.H{
		"service": "search",
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