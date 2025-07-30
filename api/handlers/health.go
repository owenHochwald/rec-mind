package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/health"
)

// PythonHealthResponse represents the health check response from Python service
type PythonHealthResponse struct {
	PythonServiceReachable bool        `json:"python_service_reachable"`
	PythonResponse         interface{} `json:"python_response"`
	ResponseTime           string      `json:"response_time,omitempty"`
	Error                  string      `json:"error,omitempty"`
}

// SystemHealthResponse represents the overall system health including Python service
type SystemHealthResponse struct {
	Status      string                `json:"status"`
	Timestamp   string                `json:"timestamp"`
	Uptime      string                `json:"uptime"`
	Version     string                `json:"version"`
	Database    interface{}           `json:"database"`
	Python      PythonHealthResponse  `json:"python"`
}

// DetailedHealthResponse represents detailed health information
type DetailedHealthResponse struct {
	Database interface{}          `json:"database"`
	Python   PythonHealthResponse `json:"python"`
	Uptime   string               `json:"uptime"`
	Version  string               `json:"version"`
}

// CheckPythonHealth checks connectivity to the Python FastAPI service
// @Summary Check Python service health
// @Description Checks if the Python FastAPI service is reachable and healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} PythonHealthResponse
// @Router /health/python [get]
func CheckPythonHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		pythonServiceURL := os.Getenv("PYTHON_SERVICE_URL")
		if pythonServiceURL == "" {
			pythonServiceURL = "http://localhost:8000" // Default Python service URL
		}

		start := time.Now()
		response := PythonHealthResponse{
			PythonServiceReachable: false,
		}

		// Make HTTP request to Python service health endpoint
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get(pythonServiceURL + "/health")
		if err != nil {
			response.Error = fmt.Sprintf("Failed to reach Python service: %v", err)
			c.JSON(200, response)
			return
		}
		defer resp.Body.Close()

		response.ResponseTime = time.Since(start).String()

		if resp.StatusCode == http.StatusOK {
			response.PythonServiceReachable = true
			
			// Parse the response body
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

		c.JSON(200, response)
	}
}

// SystemHealth handles the main health check endpoint with Python service info
// @Summary System health check
// @Description Returns overall system health including database and Python service status
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} SystemHealthResponse
// @Success 503 {object} SystemHealthResponse
// @Router /health [get]
func SystemHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		systemHealth := health.CheckSystemHealth(db, startTime)
		
		// Check Python service health
		pythonHealth := checkPythonHealthInternal()

		// Create comprehensive response
		response := SystemHealthResponse{
			Status:    systemHealth.Status,
			Timestamp: time.Now().Format(time.RFC3339),
			Uptime:    time.Since(startTime).String(),
			Version:   "1.0.0",
			Database:  systemHealth,
			Python:    pythonHealth,
		}

		// If Python service is down, consider system degraded but not completely unhealthy
		if !pythonHealth.PythonServiceReachable && response.Status == "healthy" {
			response.Status = "degraded"
		}

		statusCode := 200
		if response.Status == "unhealthy" {
			statusCode = 503
		}

		c.JSON(statusCode, response)
	}
}

// DetailedHealth handles the detailed health check endpoint
// @Summary Detailed health check
// @Description Returns detailed health information for all system components
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} DetailedHealthResponse
// @Router /health/detail [get]
func DetailedHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbHealth := health.CheckDatabase(db)
		pythonHealth := checkPythonHealthInternal()

		response := DetailedHealthResponse{
			Database: dbHealth,
			Python:   pythonHealth,
			Uptime:   time.Since(startTime).String(),
			Version:  "1.0.0",
		}

		c.JSON(200, response)
	}
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
		Timeout: 5 * time.Second, // Shorter timeout for internal checks
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