# RecMind - Semantic News Search System

A distributed semantic search system for news articles using Go microservices, Python ML pipeline, and vector databases. Uses modern backend architecture with RAG (Retrieval-Augmented Generation), distributed processing, and real-time recommendations.

![System Diagram](public/architecture.svg)
## Architecture

**System Flow:**
User Query → Go API → RabbitMQ → Query RAG Worker → Python ML Service → Pinecone Vector Search → Aggregated Results

**Tech Stack:**
- **Go**: API server (Gin), RAG workers, distributed processing
- **Python**: ML service (FastAPI), OpenAI embeddings, Pinecone integration
- **PostgreSQL**: Article and chunk storage with optimized indexing
- **Redis**: Search result caching and job status tracking
- **RabbitMQ**: Async job processing and service communication
- **Pinecone**: Vector database for semantic similarity search

## Key Features

### Semantic Search
- **Query-Based Recommendations**: Users input text queries (e.g., "trump politics") to find relevant articles
- **Vector Similarity Search**: OpenAI embeddings with cosine similarity in Pinecone
- **Hybrid Scoring**: Combines relevance and coverage metrics for optimal results
- **Article Chunking**: Smart text splitting for granular search and better context

### Distributed Architecture
- **Microservices Design**: Independent Go and Python services
- **Async Processing**: Non-blocking job queues with RabbitMQ
- **Fault Tolerance**: Circuit breakers, retry logic, and graceful degradation
- **Horizontal Scaling**: Stateless services supporting multiple instances

### Performance
- **Dual Search Modes**: Async (with job polling) and immediate (sync) responses
- **Redis Caching**: Sub-100ms responses for cached searches
- **Connection Pooling**: Optimized database and service connections
- **Processing Speed**: 100-200 articles/minute embedding generation

## API Endpoints

### Core Functionality
- `GET /health` - System health with dependency status
- `POST /api/upload` - Upload articles with ML processing
- `POST /api/v1/search/recommendations` - Query-based search (async)
- `POST /api/v1/search/immediate` - Query-based search (sync)
- `GET /api/v1/search/jobs/:job_id` - Check search job status

### Article Management
- `GET /api/v1/articles` - List articles with pagination
- `GET /api/v1/articles/:id` - Get article details
- `DELETE /api/v1/articles/:id` - Delete article

### Chunks Management
- `POST /api/v1/chunks` - Create article chunks
- `GET /api/v1/chunks` - List chunks with filtering
- `GET /api/v1/articles/:id/chunks` - Get article's chunks

## Quick Start

### Prerequisites
- Go 1.21+, Python 3.9+, Docker
- OpenAI API Key, Pinecone API Key

### Local Development
```bash
# 1. Start infrastructure
cd infra && docker-compose up -d

# 2. Start Python ML service
cd llm && python -m uvicorn app.main:app --reload

# 3. Start Go API server
cd api && go run server/main.go

# 4. Start Query RAG Worker
cd api && go run cmd/query_rag_worker/main.go

# 5. Test the system
curl http://localhost:8080/health
open http://localhost:8080/swagger/index.html
```

### Environment Variables
```bash
# Go API (.env)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=rec_mind
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
PYTHON_ML_SERVICE_URL=http://localhost:8000

# Python ML (.env)
OPENAI_API_KEY=your_openai_api_key
PINECONE_API_KEY=your_pinecone_api_key
PINECONE_INDEX_NAME=your_index_name
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

## System Workflow

### Article Processing
1. Article uploaded via API → Stored in PostgreSQL
2. Published to RabbitMQ → Python ML service consumes
3. Text chunked → OpenAI embeddings generated → Uploaded to Pinecone

### Search Processing
1. User submits query → Job created and queued
2. Query RAG Worker consumes → Sends query to Python ML service
3. Query embedded → Pinecone vector search → Results returned
4. Go worker aggregates chunks by article → Applies hybrid scoring
5. Results enriched with PostgreSQL data → Cached in Redis → Returned to user

## Documentation

- **API Docs**: http://localhost:8080/swagger/index.html
- **Python ML Docs**: http://localhost:8000/docs
- **System Architecture**: See `CLAUDE.md` for detailed technical documentation

## Development Commands

```bash
make dev      # Run Go server without building
make build    # Build binary
make test     # Run tests
make deps     # Update dependencies
```

---

*Built with Go, Python, PostgreSQL, Redis, RabbitMQ, Pinecone, and OpenAI*