# API Documentation

## Overview

The News Article Recommendation System provides RESTful APIs for article management and a real-time WebSocket interface for recommendation streaming. The system consists of two main services:

- **Go API Server** (Port 8080): Primary REST API and WebSocket endpoints
- **Python ML Service** (Port 8000): Machine learning operations and vector database management

## Authentication

All API endpoints require authentication via JWT tokens:

```http
Authorization: Bearer <jwt_token>
```

## Go API Server Endpoints

### Health Endpoints

#### Get System Health
```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "database": {
    "status": "healthy", 
    "response_time": "5ms",
    "total_conns": 25,
    "idle_conns": 20
  },
  "uptime": "1h45m30s",
  "version": "1.0.0"
}
```

#### Get Detailed Health
```http
GET /health/detail
```

**Response (200 OK):**
```json
{
  "service": {
    "status": "healthy",
    "timestamp": "2025-07-27T10:00:00Z",
    "version": "1.0.0"
  },
  "dependencies": [
    {
      "name": "postgresql",
      "status": "healthy",
      "response_time": 5.2,
      "error": null
    },
    {
      "name": "redis", 
      "status": "healthy",
      "response_time": 1.8,
      "error": null
    },
    {
      "name": "rabbitmq",
      "status": "healthy", 
      "response_time": 3.1,
      "error": null
    }
  ],
  "uptime": 6330.0
}
```

### Article Management

#### Create Article
```http
POST /api/upload
Content-Type: application/json

{
  "title": "AI Breakthrough in Machine Learning",
  "content": "Detailed article content about recent AI developments...",
  "url": "https://example.com/ai-breakthrough-2025", 
  "category": "Technology",
  "published_at": "2025-07-27T10:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "AI Breakthrough in Machine Learning",
  "content": "Detailed article content...",
  "url": "https://example.com/ai-breakthrough-2025",
  "category": "Technology", 
  "published_at": "2025-07-27T10:00:00Z",
  "created_at": "2025-07-27T10:00:00Z",
  "updated_at": "2025-07-27T10:00:00Z"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "title",
      "message": "Title is required and must be less than 500 characters"
    },
    {
      "field": "url", 
      "message": "URL must be valid and unique"
    }
  ]
}
```

#### List Articles
```http
GET /api/v1/articles?category=Technology&limit=10&offset=0&search=AI
```

**Query Parameters:**
- `category` (string, optional): Filter by article category
- `limit` (integer, optional): Number of articles to return (default: 20, max: 100)
- `offset` (integer, optional): Number of articles to skip (default: 0)
- `search` (string, optional): Search term for title/content
- `start_date` (string, optional): Filter articles after this date (ISO 8601)
- `end_date` (string, optional): Filter articles before this date (ISO 8601)

**Response (200 OK):**
```json
{
  "articles": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "AI Breakthrough in Machine Learning",
      "content": "Detailed article content...",
      "url": "https://example.com/ai-breakthrough-2025",
      "category": "Technology",
      "published_at": "2025-07-27T10:00:00Z",
      "created_at": "2025-07-27T10:00:00Z",
      "updated_at": "2025-07-27T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

#### Get Article by ID
```http
GET /api/v1/articles/{id}
```

**Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "AI Breakthrough in Machine Learning",
  "content": "Detailed article content...",
  "url": "https://example.com/ai-breakthrough-2025",
  "category": "Technology",
  "published_at": "2025-07-27T10:00:00Z",
  "created_at": "2025-07-27T10:00:00Z", 
  "updated_at": "2025-07-27T10:00:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Article not found",
  "error_code": "ARTICLE_NOT_FOUND"
}
```

#### Update Article
```http
PUT /api/v1/articles/{id}
Content-Type: application/json

{
  "title": "Updated AI Breakthrough in Machine Learning",
  "category": "Technology"
}
```

**Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Updated AI Breakthrough in Machine Learning",
  "content": "Detailed article content...",
  "url": "https://example.com/ai-breakthrough-2025",
  "category": "Technology",
  "published_at": "2025-07-27T10:00:00Z",
  "created_at": "2025-07-27T10:00:00Z",
  "updated_at": "2025-07-27T10:01:00Z"
}
```

#### Delete Article
```http
DELETE /api/v1/articles/{id}
```

**Response (200 OK):**
```json
{
  "message": "Article deleted successfully"
}
```

### WebSocket Connection

#### Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function(event) {
    console.log('Connected to WebSocket');
    
    // Send authentication
    ws.send(JSON.stringify({
        type: 'auth',
        token: 'your-jwt-token'
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};
```

#### WebSocket Message Types

**Authentication Message:**
```json
{
  "type": "auth",
  "token": "your-jwt-token"
}
```

**Recommendation Request:**
```json
{
  "type": "recommend",
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "top_k": 5,
  "score_threshold": 0.7
}
```

**Recommendation Response:**
```json
{
  "type": "recommendation",
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "recommendations": [
    {
      "id": "987fcdeb-51a3-43d2-9876-543210987654",
      "title": "Related AI Article",
      "similarity_score": 0.92,
      "category": "Technology",
      "published_at": "2025-07-26T15:30:00Z"
    }
  ],
  "processing_time": "45ms",
  "timestamp": "2025-07-27T10:00:00Z"
}
```

**Error Message:**
```json
{
  "type": "error",
  "error": "Authentication failed",
  "error_code": "AUTH_FAILED",
  "timestamp": "2025-07-27T10:00:00Z"
}
```

## Python ML Service Endpoints

### Health Endpoints

#### Basic Health Check
```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-07-27T10:00:00Z",
  "version": "1.0.0"
}
```

#### Detailed Health Check
```http
GET /health/detail
```

**Response (200 OK):**
```json
{
  "service": {
    "status": "healthy",
    "timestamp": "2025-07-27T10:00:00Z",
    "version": "1.0.0"
  },
  "dependencies": [
    {
      "name": "openai",
      "status": "healthy",
      "response_time": 245.6,
      "error": null
    },
    {
      "name": "pinecone",
      "status": "healthy", 
      "response_time": 12.3,
      "error": null
    }
  ],
  "uptime": 3600.0
}
```

### Embedding Operations

#### Generate Single Embedding
```http
POST /embeddings/generate
Content-Type: application/json

{
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "text": "This is the article content to generate embeddings for..."
}
```

**Response (200 OK):**
```json
{
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "embeddings": [0.1, 0.2, -0.1, ...],
  "dimensions": 1536,
  "model": "text-embedding-3-small",
  "tokens_used": 45
}
```

#### Generate Batch Embeddings
```http
POST /embeddings/batch
Content-Type: application/json

{
  "items": [
    {
      "article_id": "123e4567-e89b-12d3-a456-426614174000",
      "text": "First article content..."
    },
    {
      "article_id": "987fcdeb-51a3-43d2-9876-543210987654", 
      "text": "Second article content..."
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "results": [
    {
      "article_id": "123e4567-e89b-12d3-a456-426614174000",
      "embeddings": [0.1, 0.2, -0.1, ...],
      "dimensions": 1536,
      "model": "text-embedding-3-small",
      "tokens_used": 45
    },
    {
      "article_id": "987fcdeb-51a3-43d2-9876-543210987654",
      "embeddings": [0.3, -0.1, 0.2, ...],
      "dimensions": 1536,
      "model": "text-embedding-3-small", 
      "tokens_used": 52
    }
  ],
  "total_tokens": 97,
  "processing_time": 1.234
}
```

#### Upload Embedding to Pinecone
```http
POST /embeddings/upload
Content-Type: application/json

{
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "embeddings": [0.1, 0.2, -0.1, ...],
  "metadata": {
    "title": "AI Breakthrough in Machine Learning",
    "category": "Technology",
    "published_at": "2025-07-27T10:00:00Z"
  }
}
```

**Response (200 OK):**
```json
{
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "uploaded",
  "upload_time": 0.156,
  "upserted_count": 1
}
```

### Vector Search Operations

#### Search Similar Articles
```http
POST /search/similar
Content-Type: application/json

{
  "article_id": "123e4567-e89b-12d3-a456-426614174000",
  "top_k": 5,
  "score_threshold": 0.7,
  "include_metadata": true
}
```

**Response (200 OK):**
```json
{
  "query_article_id": "123e4567-e89b-12d3-a456-426614174000",
  "results": [
    {
      "id": "987fcdeb-51a3-43d2-9876-543210987654",
      "score": 0.92,
      "metadata": {
        "title": "Related AI Article",
        "category": "Technology",
        "published_at": "2025-07-26T15:30:00Z"
      }
    },
    {
      "id": "456789ab-cdef-1234-5678-90abcdef1234",
      "score": 0.85,
      "metadata": {
        "title": "Machine Learning Advances",
        "category": "Technology",
        "published_at": "2025-07-25T12:00:00Z"
      }
    }
  ],
  "total_found": 2,
  "processing_time": 0.045
}
```

#### Get Index Statistics
```http
GET /index/stats
```

**Response (200 OK):**
```json
{
  "total_vectors": 15420,
  "dimension": 1536,
  "index_fullness": 0.15,
  "namespaces": {
    "": {
      "vector_count": 15420
    }
  }
}
```

## Error Handling

### Standard Error Response Format
```json
{
  "error": "Human readable error message",
  "error_code": "MACHINE_READABLE_CODE",
  "timestamp": "2025-07-27T10:00:00Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | External service unavailable |

### Rate Limiting

All endpoints are rate limited:
- **Authenticated users**: 1000 requests/hour
- **Anonymous users**: 100 requests/hour  
- **ML operations**: 60 requests/minute

Rate limit headers included in responses:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1643723400
```

## SDK Examples

### JavaScript/TypeScript
```typescript
class RecMindClient {
  constructor(private baseUrl: string, private token: string) {}
  
  async createArticle(article: CreateArticleRequest): Promise<Article> {
    const response = await fetch(`${this.baseUrl}/api/upload`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify(article)
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    return response.json();
  }
  
  connectWebSocket(): WebSocket {
    const ws = new WebSocket(`${this.baseUrl.replace('http', 'ws')}/ws`);
    
    ws.onopen = () => {
      ws.send(JSON.stringify({ type: 'auth', token: this.token }));
    };
    
    return ws;
  }
}
```

### Python
```python
import httpx
import asyncio

class RecMindClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.headers = {"Authorization": f"Bearer {token}"}
    
    async def create_article(self, article: dict) -> dict:
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{self.base_url}/api/upload",
                json=article,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
    
    async def generate_embedding(self, article_id: str, text: str) -> dict:
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{self.base_url}:8000/embeddings/generate",
                json={"article_id": article_id, "text": text}
            )
            response.raise_for_status()
            return response.json()
```

### Go
```go
type Client struct {
    baseURL string
    token   string
    http    *http.Client
}

func (c *Client) CreateArticle(ctx context.Context, article CreateArticleRequest) (*Article, error) {
    body, err := json.Marshal(article)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/upload", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.token)
    
    resp, err := c.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }
    
    var result Article
    return &result, json.NewDecoder(resp.Body).Decode(&result)
}
```

This API documentation provides comprehensive coverage of all endpoints with practical examples for integration.