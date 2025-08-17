package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rec-mind/internal/redis"
	"rec-mind/models"
	"rec-mind/mq"
	"rec-mind/pkg/response"
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

func (sc *SearchController) SearchByQuery(c *gin.Context) {
	var req QuerySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
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
	job := models.QuerySearchJob{
		JobID:          jobID,
		Query:          req.Query,
		SessionID:      req.SessionID,
		MaxResults:     req.MaxResults,
		ScoreThreshold: req.ScoreThreshold,
		CreatedAt:      time.Now(),
		CorrelationID:  req.CorrelationID,
	}

	if err := mq.PublishQuerySearchJob(job); err != nil {
		response.InternalServerError(c, "Failed to queue search job")
		return
	}

	data := QuerySearchJobResponse{
		JobID:     jobID,
		Status:    "queued",
		Message:   "Search job has been queued for processing",
		PollURL:   fmt.Sprintf("/api/v1/search/jobs/%s", jobID),
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	}

	response.Accepted(c, data)
}

func (sc *SearchController) GetQuerySearchJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		response.BadRequest(c, "job_id parameter is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check Redis for result
	key := fmt.Sprintf("query_search_result:%s", jobID)
	resultJSON, err := redis.RedisClient.Get(ctx, key).Result()
	
	if err != nil {
		data := gin.H{
			"job_id": jobID,
			"status": "processing",
			"message": "Job is still being processed or does not exist",
		}
		c.JSON(http.StatusNotFound, data)
		return
	}

	var result models.QueryRecommendationResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		response.InternalServerError(c, "Failed to parse job result")
		return
	}

	response.Success(c, result)
}

func (sc *SearchController) SearchWithImmediateResponse(c *gin.Context) {
	var req QuerySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
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
	job := models.QuerySearchJob{
		JobID:          jobID,
		Query:          req.Query,
		SessionID:      req.SessionID,
		MaxResults:     req.MaxResults,
		ScoreThreshold: req.ScoreThreshold,
		CreatedAt:      time.Now(),
		CorrelationID:  req.CorrelationID,
	}

	if err := mq.PublishQuerySearchJob(job); err != nil {
		response.InternalServerError(c, "Failed to queue search job")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := fmt.Sprintf("query_search_result:%s", jobID)
	
	for i := 0; i < 20; i++ {
		resultJSON, err := redis.RedisClient.Get(ctx, key).Result()
		if err == nil {
			var result models.QueryRecommendationResult
			if err := json.Unmarshal([]byte(resultJSON), &result); err == nil {
				response.Success(c, result)
				return
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	data := QuerySearchJobResponse{
		JobID:     jobID,
		Status:    "processing",
		Message:   "Search job is being processed. Use the poll_url to check status",
		PollURL:   fmt.Sprintf("/api/v1/search/jobs/%s", jobID),
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	}

	response.Accepted(c, data)
}

func (sc *SearchController) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	health := gin.H{
		"service": "search",
		"status":  "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := redis.HealthCheck(ctx); err != nil {
		health["redis_status"] = "unhealthy"
		health["redis_error"] = err.Error()
		health["status"] = "degraded"
	} else {
		health["redis_status"] = "healthy"
	}

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