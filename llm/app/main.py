"""FastAPI ML Service for embeddings and vector operations."""

import time
import uuid
from contextlib import asynccontextmanager
from typing import Dict, Any

import structlog
import uvicorn
from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse

from .config import get_settings
from .models import (
    EmbeddingRequest, EmbeddingResponse,
    BatchEmbeddingRequest, BatchEmbeddingResponse,
    PineconeUploadRequest, PineconeSearchRequest, PineconeSearchResponse,
    HealthStatus, DependencyHealth, DetailedHealthResponse,
    ErrorResponse
)
from .embeddings import get_embeddings_service, EmbeddingsService
from .vectordb import get_vectordb_service, VectorDBService
from .utils import endpoint_error_handler


# Configure structured logging
structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
        structlog.processors.JSONRenderer()
    ],
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger(__name__)
settings = get_settings()

# Service startup time for uptime calculation
SERVICE_START_TIME = time.time()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan events."""
    logger.info("Starting ML service", version="1.0.0")
    
    # Startup
    try:
        # Initialize services
        embeddings_svc = await get_embeddings_service()
        vectordb_svc = await get_vectordb_service()
        
        logger.info("Services initialized successfully")
        yield
        
    except Exception as e:
        logger.error("Failed to initialize services", error=str(e))
        raise
    finally:
        # Shutdown
        logger.info("Shutting down ML service")


# Create FastAPI application
app = FastAPI(
    title="RecMind ML Service",
    description="Machine Learning service for embeddings generation and vector operations",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# Add trusted host middleware
app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=["*"]  # Configure appropriately for production
)


# Request ID middleware
@app.middleware("http")
async def add_request_id(request: Request, call_next):
    """Add unique request ID to each request."""
    request_id = str(uuid.uuid4())
    request.state.request_id = request_id
    
    # Add to logger context
    structlog.contextvars.clear_contextvars()
    structlog.contextvars.bind_contextvars(request_id=request_id)
    
    response = await call_next(request)
    response.headers["X-Request-ID"] = request_id
    return response


# Global exception handler
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Handle unexpected exceptions."""
    request_id = getattr(request.state, "request_id", "unknown")
    
    logger.error(
        "Unhandled exception",
        request_id=request_id,
        path=request.url.path,
        method=request.method,
        error=str(exc)
    )
    
    return JSONResponse(
        status_code=500,
        content=ErrorResponse(
            error="Internal server error",
            error_code="INTERNAL_ERROR",
            request_id=request_id
        ).dict()
    )


@app.get("/health", response_model=HealthStatus)
async def health_check():
    """Basic health check endpoint."""
    return HealthStatus(status="healthy")


@app.get("/health/detail", response_model=DetailedHealthResponse)
async def detailed_health_check(
    embeddings_svc: EmbeddingsService = Depends(get_embeddings_service),
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Detailed health check with dependency status."""
    dependencies = []
    
    # Check OpenAI connectivity
    openai_health = await embeddings_svc.health_check()
    dependencies.append(DependencyHealth(
        name="openai",
        status=openai_health["status"],
        response_time=openai_health.get("response_time"),
        error=openai_health.get("error")
    ))
    
    # Check Pinecone connectivity
    pinecone_health = await vectordb_svc.health_check()
    dependencies.append(DependencyHealth(
        name="pinecone",
        status=pinecone_health["status"],
        response_time=pinecone_health.get("response_time"),
        error=pinecone_health.get("error")
    ))
    
    # Determine overall service status
    service_status = "healthy" if all(dep.status == "healthy" for dep in dependencies) else "unhealthy"
    
    return DetailedHealthResponse(
        service=HealthStatus(status=service_status),
        dependencies=dependencies,
        uptime=time.time() - SERVICE_START_TIME
    )


@app.post("/embeddings/generate", response_model=EmbeddingResponse)
@endpoint_error_handler("embedding generation")
async def generate_embedding(
    request: EmbeddingRequest,
    embeddings_svc: EmbeddingsService = Depends(get_embeddings_service)
):
    """Generate embedding for a single text."""
    return await embeddings_svc.generate_embedding(request)


@app.post("/embeddings/batch", response_model=BatchEmbeddingResponse)
@endpoint_error_handler("batch embedding generation")
async def generate_batch_embeddings(
    request: BatchEmbeddingRequest,
    embeddings_svc: EmbeddingsService = Depends(get_embeddings_service)
):
    """Generate embeddings for multiple texts."""
    logger.info("Processing batch embedding request", batch_size=len(request.items))
    result = await embeddings_svc.generate_batch_embeddings(request)
    logger.info("Batch embeddings generated successfully", batch_size=len(result.results))
    return result


@app.post("/embeddings/upload")
@endpoint_error_handler("embedding upload")
async def upload_embedding(
    request: PineconeUploadRequest,
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Upload embedding vector to Pinecone."""
    return await vectordb_svc.upload_embedding(request)


@app.post("/search/similar", response_model=PineconeSearchResponse)
@endpoint_error_handler("similarity search")
async def search_similar_articles(
    request: PineconeSearchRequest,
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Search for similar articles using vector similarity."""
    result = await vectordb_svc.search_similar(request)
    logger.info(
        "Similarity search completed",
        article_id=str(request.article_id),
        results_found=len(result.results)
    )
    return result


@app.get("/index/stats")
@endpoint_error_handler("index statistics retrieval")
async def get_index_stats(
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Get Pinecone index statistics."""
    return await vectordb_svc.get_index_stats()


@app.get("/test/pinecone")
@endpoint_error_handler("Pinecone connection test")
async def test_pinecone_connection(
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Test Pinecone connection and basic operations."""
    return await vectordb_svc.test_connection()


@app.post("/embeddings/generate-and-upload")
@endpoint_error_handler("embedding generation and upload")
async def generate_and_upload_embedding(
    request: EmbeddingRequest,
    embeddings_svc: EmbeddingsService = Depends(get_embeddings_service),
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Generate embedding and upload to Pinecone in one step."""
    # Generate embedding
    embedding_response = await embeddings_svc.generate_embedding(request)
    
    # Upload to Pinecone
    upload_request = PineconeUploadRequest(
        article_id=request.article_id,
        embeddings=embedding_response.embeddings,
        metadata={
            "title": request.text[:100],  # First 100 chars as title
            "model": embedding_response.model,
            "dimensions": embedding_response.dimensions
        }
    )
    
    upload_response = await vectordb_svc.upload_embedding(upload_request)
    
    return {
        "article_id": str(request.article_id),
        "embedding": {
            "dimensions": embedding_response.dimensions,
            "model": embedding_response.model,
            "tokens_used": embedding_response.tokens_used
        },
        "upload": upload_response,
        "status": "completed"
    }




if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host=settings.ml_service_host,
        port=settings.ml_service_port,
        log_level=settings.log_level,
        reload=True
    )