# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a distributed news article recommendation system built in Go that demonstrates RAG (Retrieval-Augmented Generation), distributed systems architecture, and modern backend technologies. The system scrapes news articles, generates embeddings, and provides real-time semantic similarity recommendations via WebSocket streaming.

Tech Stack: Go (for handling the backend), PostgreSQL (for storing articles), Redis (for caching requests), Pinecone (for RAG app), RabbitMQ (for distributed system), WebSocket (for communication with pinecone / postgresql), Prometheus (for analysis), docker, AWS EC2
## Architecture

The system follows a distributed microservices architecture with async processing:

User Request → Go API → Redis Cache Check → RabbitMQ Queue → RAG Workers → Pinecone Vector Search → WebSocket Response

- API Server: Gin-based REST API with WebSocket support for real-time recommendations
- RAG Workers: Go workers that consume recommendation jobs and perform vector similarity search
- Data Pipeline: Article scraper and embedding generation service
- Caching Layer: Redis for API responses and job status
- Vector Database: Pinecone for semantic similarity search
- Message Queue: RabbitMQ for async job processing

### Key Components

- `api/server/main.go`: Application entry point with route definitions
- `api/controllers/`: HTTP handlers (currently only `UploadArticle` is implemented)  
- `api/models/`: Data models (Article struct)
- `api/database/`: PostgreSQL connection and initialization
- `api/mq/`: RabbitMQ publisher and consumer setup
- `api/config/`: Environment configuration loading

## Current Implementation Status

- [X] Project foundation and database schema
- [ ] Article scraping and data ingestion
- [ ] OpenAI embeddings generation
- [X] Pinecone vector database integration
- [ ] HTTP API server with caching
- [ ] RabbitMQ async job processing
- [ ] RAG worker implementation
- [ ] WebSocket real-time streaming
- [ ] Frontend client
- [ ] Monitoring and production deployment

## Development Commands

The project is organized into sequential development tickets (see project documentation). Each ticket builds incrementally toward a complete system.

### Building and Running
- `make dev` - Run without building binary (recommended for development)
- `make run` - Build and run the application
- `make build` - Build binary only
- `make clean` - Remove build artifacts

### Testing
- `make test` - Run all tests with verbose output

### Dependencies
- `make deps` - Download and tidy Go modules

### Docker
- `make docker-build` - Build Docker image
- `make docker-run` - Run containerized application
- `make docker-compose-up` - Start with dependencies
- `make docker-compose-down` - Stop all services

## Infrastructure Setup

Start development dependencies from project root:
```bash
cd infra/
docker-compose up -d  # Starts Redis and RabbitMQ
```

## Environment Variables

Required for database connection:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

The application will load from `.env` file if present, otherwise uses OS environment variables.

## API Endpoints

Current implementation:
- `GET /health` - Health check
- `POST /api/upload` - Upload article (publishes to RabbitMQ queue)
- `POST /api/interact` - Placeholder (not implemented)
- `GET /api/recommend` - Placeholder (not implemented)

### Article Management

- GET /health - Comprehensive health check with dependency status
- GET /health/detail - Detailed system health and metrics
- GET /api/v1/articles - List articles with pagination and filtering
- GET /api/v1/articles/:id - Get specific article details

## Recommendation System
- POST /api/v1/articles/:id/recommend - Request recommendations (async)
- GET /api/v1/jobs/:job_id - Check recommendation job status
- WS /ws - WebSocket endpoint for real-time recommendation streaming
 
## Database Schema
### Articles Table
```sql
    CREATE TABLE articles (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        title VARCHAR(500) NOT NULL,
        content TEXT NOT NULL,
        url VARCHAR(1000) UNIQUE NOT NULL,
        category VARCHAR(100) NOT NULL,
        published_at TIMESTAMP NOT NULL,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW()
    );

    CREATE INDEX idx_articles_category ON articles(category);
    CREATE INDEX idx_articles_published_at ON articles(published_at);
```

## Message Queue Patterns
### Recommendation Job Flow
1. User clicks article → API publishes job to recommendation_jobs queue
2. RAG worker consumes job → performs Pinecone vector search
3. Worker enriches results with PostgreSQL data → stores in Redis
4. Results streamed to user via WebSocket connection

### Message Structure 
```json
    {
    "job_id": "uuid",
    "article_id": "uuid",
    "session_id": "string",
    "created_at": "timestamp",
    "correlation_id": "string"
    }
```

## Vector Search Implementation
### Pinecone Integration

- Index Dimension: 1536 (OpenAI text-embedding-3-small)
- Similarity Metric: Cosine similarity
- Top-K Results: 5 recommendations per query
- Score Threshold: 0.7 minimum similarity

### Embedding Strategy

- Combine article title + first 500 words of content
- Generate embeddings via OpenAI API with batching
- Store in Pinecone with metadata: article_id, title, category, published_at

## WebSocket Implementation
### Real-time Features

- Persistent connections for instant recommendation delivery
- Connection pooling with session management
- Heartbeat/ping-pong for connection health
- Graceful reconnection handling

## Monitoring & Observability
### Prometheus Metrics
- HTTP request rates and latency
- Recommendation processing time
- Cache hit/miss rates
- Queue depth and worker status
- Database connection pool stats

### Health Checks

- Dependency connectivity (PostgreSQL, Redis, RabbitMQ, Pinecone)
- Service-specific health indicators
- Performance metrics and thresholds

## Development Guidelines
### Code Organization
- Use dependency injection for testability
- Implement interfaces for external services
- Follow clean architecture patterns
- Add comprehensive error handling and logging

### Testing Strategy
- Unit tests for business logic
- Load tests for WebSocket connections
- End-to-end tests for recommendation flow

### Performance Considerations
- Cache frequently accessed articles
- Batch vector operations when possible
- Use connection pooling for all external services
- Implement circuit breakers for external APIs

Articles uploaded via `/api/upload` are published to the `article_events` queue for processing by ML components.