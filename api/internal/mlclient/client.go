package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// MLClient handles communication with the Python ML service
type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

// EmbeddingRequest represents a single embedding request
type EmbeddingRequest struct {
	ArticleID uuid.UUID `json:"article_id"`
	Text      string    `json:"text"`
}

// BatchEmbeddingRequest represents a batch embedding request
type BatchEmbeddingRequest struct {
	Items []EmbeddingRequest `json:"items"`
}

// EmbeddingResponse represents a single embedding response
type EmbeddingResponse struct {
	ArticleID   uuid.UUID `json:"article_id"`
	Embeddings  []float64 `json:"embeddings"`
	Dimensions  int       `json:"dimensions"`
	Model       string    `json:"model"`
	TokensUsed  int       `json:"tokens_used"`
}

// BatchEmbeddingResponse represents a batch embedding response
type BatchEmbeddingResponse struct {
	Results         []EmbeddingResponse `json:"results"`
	TotalTokens     int                 `json:"total_tokens"`
	ProcessingTime  float64             `json:"processing_time"`
}

// BatchAndUploadResponse represents the response from batch-and-upload endpoint
type BatchAndUploadResponse struct {
	Embeddings BatchEmbeddingResponse `json:"embeddings"`
	Uploads    []UploadResult         `json:"uploads"`
	Summary    ProcessingSummary      `json:"summary"`
}

// UploadResult represents a single upload result
type UploadResult struct {
	ArticleID     string  `json:"article_id"`
	Status        string  `json:"status"`
	UploadTime    float64 `json:"upload_time"`
	UpsertedCount int     `json:"upserted_count"`
}

// ProcessingSummary represents processing summary
type ProcessingSummary struct {
	TotalProcessed int     `json:"total_processed"`
	TotalTokens    int     `json:"total_tokens"`
	ProcessingTime float64 `json:"processing_time"`
}

// MLServiceError represents an error from the ML service
type MLServiceError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e *MLServiceError) Error() string {
	return fmt.Sprintf("ML service error (status %d): %s", e.StatusCode, e.Message)
}

// NewMLClient creates a new ML service client
func NewMLClient() *MLClient {
	baseURL := os.Getenv("PYTHON_ML_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for ML operations
		},
	}
}

// Health checks the health of the ML service
func (c *MLClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ML service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// GenerateBatchEmbeddingsAndUpload sends articles to the Python service for embedding generation and Pinecone upload
func (c *MLClient) GenerateBatchEmbeddingsAndUpload(ctx context.Context, articles []EmbeddingRequest) (*BatchAndUploadResponse, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles provided for embedding generation")
	}

	// Prepare the batch request
	batchRequest := BatchEmbeddingRequest{
		Items: articles,
	}

	// Marshal the request
	requestBody, err := json.Marshal(batchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings/batch-and-upload", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("batch embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if json.Unmarshal(responseBody, &errorResponse) == nil {
			if detail, ok := errorResponse["detail"].(string); ok {
				return nil, &MLServiceError{
					Message:    detail,
					StatusCode: resp.StatusCode,
				}
			}
		}
		return nil, &MLServiceError{
			Message:    fmt.Sprintf("Request failed with status %d: %s", resp.StatusCode, string(responseBody)),
			StatusCode: resp.StatusCode,
		}
	}

	// Parse successful response
	var result BatchAndUploadResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}

	return &result, nil
}

// GenerateSingleEmbeddingAndUpload sends a single article for embedding generation and upload
func (c *MLClient) GenerateSingleEmbeddingAndUpload(ctx context.Context, articleID uuid.UUID, text string) (*BatchAndUploadResponse, error) {
	return c.GenerateBatchEmbeddingsAndUpload(ctx, []EmbeddingRequest{
		{
			ArticleID: articleID,
			Text:      text,
		},
	})
}

// CreateEmbeddingText combines article title and content for embedding generation
func CreateEmbeddingText(title, content string) string {
	// Combine title and content with clear separation
	// This follows common practices for article embedding
	return fmt.Sprintf("%s\n\n%s", title, content)
}