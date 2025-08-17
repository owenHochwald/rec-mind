package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"rec-mind/internal/database"
	"rec-mind/internal/redis"
	"rec-mind/mq"
)

type HealthResponse struct {
	Service      string                 `json:"service"`
	Status       string                 `json:"status"`
	Timestamp    string                 `json:"timestamp"`
	Uptime       string                 `json:"uptime"`
	Version      string                 `json:"version"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

type DependencyStatus struct {
	Status       string  `json:"status"`
	ResponseTime *string `json:"response_time,omitempty"`
	Error        *string `json:"error,omitempty"`
}

// SystemHealth provides comprehensive health check for all system components
// @Summary Comprehensive system health check
// @Description Returns health status for all system components including database, Redis, RabbitMQ, and Python ML service
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Success 503 {object} HealthResponse
// @Router /health [get]
func SystemHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dependencies := make(map[string]interface{})
		overallHealthy := true

		// Check Database
		dbStart := time.Now()
		if err := db.Pool.Ping(ctx); err != nil {
			dependencies["database"] = DependencyStatus{
				Status: "unhealthy",
				Error:  stringPtr(err.Error()),
			}
			overallHealthy = false
		} else {
			dbTime := time.Since(dbStart)
			dependencies["database"] = DependencyStatus{
				Status:       "healthy",
				ResponseTime: stringPtr(dbTime.String()),
			}
		}

		// Check Redis
		redisStart := time.Now()
		if err := redis.HealthCheck(ctx); err != nil {
			dependencies["redis"] = DependencyStatus{
				Status: "unhealthy",
				Error:  stringPtr(err.Error()),
			}
			overallHealthy = false
		} else {
			redisTime := time.Since(redisStart)
			dependencies["redis"] = DependencyStatus{
				Status:       "healthy",
				ResponseTime: stringPtr(redisTime.String()),
			}
		}

		// Check RabbitMQ
		if mq.MQChannel == nil || mq.MQChannel.IsClosed() {
			dependencies["rabbitmq"] = DependencyStatus{
				Status: "unhealthy",
				Error:  stringPtr("Connection closed or not initialized"),
			}
			overallHealthy = false
		} else {
			dependencies["rabbitmq"] = DependencyStatus{
				Status: "healthy",
			}
		}

		// Check Python ML Service
		mlStart := time.Now()
		pythonHealth := checkPythonHealthInternal()
		if !pythonHealth.PythonServiceReachable {
			dependencies["python_ml_service"] = DependencyStatus{
				Status: "unhealthy",
				Error:  stringPtr(pythonHealth.Error),
			}
			overallHealthy = false
		} else {
			mlTime := time.Since(mlStart)
			dependencies["python_ml_service"] = map[string]interface{}{
				"status":        "healthy",
				"response_time": mlTime.String(),
				"response":      pythonHealth.PythonResponse,
			}
		}

		// Check Query RAG Worker (check if query_search_jobs queue exists and is accessible)
		ragWorkerStart := time.Now()
		ragWorkerHealthy := checkQueryRAGWorkerHealth()
		if !ragWorkerHealthy.IsHealthy {
			dependencies["query_rag_worker"] = DependencyStatus{
				Status: "unhealthy",
				Error:  stringPtr(ragWorkerHealthy.Error),
			}
			overallHealthy = false
		} else {
			ragWorkerTime := time.Since(ragWorkerStart)
			dependencies["query_rag_worker"] = DependencyStatus{
				Status:       "healthy",
				ResponseTime: stringPtr(ragWorkerTime.String()),
			}
		}

		// Determine overall status
		status := "healthy"
		statusCode := http.StatusOK
		if !overallHealthy {
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		response := HealthResponse{
			Service:      "RecMind API",
			Status:       status,
			Timestamp:    time.Now().Format(time.RFC3339),
			Uptime:       time.Since(startTime).String(),
			Version:      "1.0.0",
			Dependencies: dependencies,
		}

		c.JSON(statusCode, response)
	}
}

// Internal types for Python health check
type PythonHealthResponse struct {
	PythonServiceReachable bool        `json:"python_service_reachable"`
	PythonResponse         interface{} `json:"python_response"`
	ResponseTime           string      `json:"response_time,omitempty"`
	Error                  string      `json:"error,omitempty"`
}

// checkPythonHealthInternal is an internal helper function to check Python service health
func checkPythonHealthInternal() PythonHealthResponse {
	pythonServiceURL := os.Getenv("PYTHON_SERVICE_URL")
	if pythonServiceURL == "" {
		pythonServiceURL = "http://localhost:8000"
	}

	start := time.Now()
	response := PythonHealthResponse{
		PythonServiceReachable: false,
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(pythonServiceURL + "/health")
	if err != nil {
		response.Error = fmt.Sprintf("Failed to reach Python service: %v", err)
		return response
	}
	defer resp.Body.Close()

	response.ResponseTime = time.Since(start).String()

	if resp.StatusCode == http.StatusOK {
		response.PythonServiceReachable = true
		
		var pythonHealth interface{}
		if err := json.NewDecoder(resp.Body).Decode(&pythonHealth); err != nil {
			response.PythonResponse = fmt.Sprintf("Status: %d, but failed to parse response: %v", resp.StatusCode, err)
		} else {
			response.PythonResponse = pythonHealth
		}
	} else {
		response.Error = fmt.Sprintf("Python service returned status code: %d", resp.StatusCode)
		response.PythonResponse = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return response
}

// RAGWorkerHealthResponse represents the health status of the Query RAG Worker
type RAGWorkerHealthResponse struct {
	IsHealthy bool   `json:"is_healthy"`
	Error     string `json:"error,omitempty"`
}

// checkQueryRAGWorkerHealth checks if the Query RAG Worker is healthy by verifying queue accessibility
func checkQueryRAGWorkerHealth() RAGWorkerHealthResponse {
	response := RAGWorkerHealthResponse{
		IsHealthy: false,
	}

	// Check if RabbitMQ connection is available
	if mq.MQChannel == nil || mq.MQChannel.IsClosed() {
		response.Error = "RabbitMQ connection not available"
		return response
	}

	// Try to inspect the query_search_jobs queue to see if it's accessible
	_, err := mq.MQChannel.QueueInspect("query_search_jobs")
	if err != nil {
		response.Error = fmt.Sprintf("Cannot access query_search_jobs queue: %v", err)
		return response
	}

	response.IsHealthy = true
	return response
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}