# RecMind ML Service

FastAPI-based Python service for machine learning operations including OpenAI embeddings generation and Pinecone vector database operations.

## Features

- 🧠 **OpenAI Embeddings**: Generate text embeddings with retry logic and rate limiting
- 🔍 **Vector Search**: Pinecone integration for similarity search operations
- 📊 **Health Monitoring**: Comprehensive health checks for all dependencies
- 🚀 **Async Operations**: Full async/await support for optimal performance
- 📋 **Auto Documentation**: FastAPI auto-generated OpenAPI docs
- 🛡️ **Error Handling**: Robust error handling with structured logging
- 🐳 **Containerized**: Docker support for easy deployment

## Quick Start

### 1. Environment Setup

```bash
# Create .env file with your API keys
OPENAI_API_KEY=your-openai-key-here
PINECONE_API_KEY=your-pinecone-key-here
PINECONE_INDEX_NAME=news-articles
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=recmind
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
```

### 2. Local Development

```bash
# Install dependencies
pip install -r requirements.txt

# Option 1: Quick start (recommended)
python run.py

# Option 2: Check infrastructure first
python dev-setup.py

# Option 3: Start with uvicorn directly
python -m uvicorn app.main:app --reload --port 8000
```

### 3. Infrastructure Setup (Optional)

The service works with graceful degradation - core features work without infrastructure:

```bash
# Start full infrastructure (PostgreSQL + RabbitMQ + Redis)
cd ../infra && docker-compose up -d

# Check what's running
python dev-setup.py
```

**Features available without infrastructure:**
- ✅ Text chunking with LangChain
- ✅ OpenAI embeddings generation  
- ✅ Pinecone vector operations
- ✅ API documentation

**Features requiring infrastructure:**
- 🔶 PostgreSQL: Chunk storage in database
- 🔶 RabbitMQ: Async article processing queue

### 3. Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build manually
docker build -t recmind-ml-service .
docker run -p 8000:8000 --env-file .env recmind-ml-service
```

## API Endpoints

### Health Checks
- `GET /health` - Basic service health
- `GET /health/detail` - Detailed health with dependency status

### Article Processing
- `POST /articles/process` - Full article processing with chunking
- `POST /text/chunk` - Split text into chunks
- `GET /chunks/parameters` - View chunking configuration

### Embeddings  
- `POST /embeddings/batch` - Batch embedding generation
- `POST /embeddings/batch-and-upload` - Generate and upload to Pinecone

### Vector Search
- `POST /search/similar` - Find similar articles
- `GET /index/stats` - Pinecone index statistics

### Documentation
- `GET /docs` - Interactive API documentation
- `GET /redoc` - Alternative documentation

## Usage Examples

### Generate Embedding
```python
import httpx

async def generate_embedding():
    async with httpx.AsyncClient() as client:
        response = await client.post("http://localhost:8000/embeddings/generate", json={
            "article_id": "123e4567-e89b-12d3-a456-426614174000",
            "text": "Your article content here"
        })
        return response.json()
```

### Search Similar Articles
```python
async def search_similar():
    async with httpx.AsyncClient() as client:
        response = await client.post("http://localhost:8000/search/similar", json={
            "article_id": "123e4567-e89b-12d3-a456-426614174000",
            "top_k": 5,
            "score_threshold": 0.7
        })
        return response.json()
```

## Testing

```bash
# Start the service first
python run.py

# Then run tests (in another terminal)
python test.py
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `OPENAI_API_KEY` | Required | OpenAI API key |
| `PINECONE_API_KEY` | Required | Pinecone API key |
| `PINECONE_ENVIRONMENT` | us-west1-gcp | Pinecone environment |
| `PINECONE_INDEX_NAME` | news-articles | Pinecone index name |
| `ML_SERVICE_HOST` | 0.0.0.0 | Service host |
| `ML_SERVICE_PORT` | 8000 | Service port |
| `LOG_LEVEL` | info | Logging level |
| `OPENAI_MODEL` | text-embedding-3-small | OpenAI model |
| `EMBEDDING_DIMENSIONS` | 1536 | Embedding dimensions |
| `BATCH_SIZE` | 100 | Batch processing size |

## Architecture

The service is built with:

- **FastAPI**: Modern Python web framework
- **Pydantic**: Data validation and settings
- **Structlog**: Structured logging
- **OpenAI**: Embeddings generation
- **Pinecone**: Vector database operations
- **Uvicorn**: ASGI server
- **Tenacity**: Retry logic for external APIs

## Performance Features

- Async/await throughout the application
- Connection pooling for external APIs
- Retry logic with exponential backoff
- Batch processing for multiple operations
- Request timeout configuration
- CORS support for web applications