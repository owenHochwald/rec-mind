"""FastAPI ML Service for embeddings and vector operations."""

import asyncio
import time
import uuid
from contextlib import asynccontextmanager
from typing import List

import structlog
import uvicorn
from fastapi import FastAPI, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse

from .config import get_settings
from .models import (
    HealthStatus, DependencyHealth, DetailedHealthResponse,
    ErrorResponse, ArticleProcessingRequest, ArticleProcessingResponse,
    ChunkRequest, ChunkResponse, EmbeddingRequest, BatchEmbeddingRequest,
    PineconeUploadRequest
)
from .embeddings import get_embeddings_service, EmbeddingsService
from .vectordb import get_vectordb_service, VectorDBService
from .chunking import get_chunking_service, ChunkingService
from .database import get_database_service, DatabaseService
from .mq_consumer import get_article_consumer, start_article_processing
from .query_search_consumer import get_query_search_consumer, start_query_search_processing
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
        # Initialize core services (always required)
        embeddings_svc = await get_embeddings_service()
        vectordb_svc = await get_vectordb_service()
        chunking_svc = await get_chunking_service()
        
        logger.info("Core services initialized successfully")
        
        # Initialize optional services (graceful degradation)
        database_svc = await get_database_service()
        try:
            await database_svc.initialize()
            logger.info("PostgreSQL connection established")
        except Exception as db_error:
            logger.warning("PostgreSQL connection failed - article chunking storage disabled", error=str(db_error))
        
        # Initialize RabbitMQ consumers
        consumer = await get_article_consumer()
        query_search_consumer = await get_query_search_consumer()
        try:
            await consumer.connect()
            await query_search_consumer.connect()
            # Start consumers as background tasks
            asyncio.create_task(start_article_processing())
            asyncio.create_task(start_query_search_processing())
            logger.info("RabbitMQ consumers started")
        except Exception as mq_error:
            logger.warning("RabbitMQ connection failed - message processing disabled", error=str(mq_error))
        
        logger.info("Service startup completed (some features may be disabled)")
        yield
        
    except Exception as e:
        logger.error("Failed to initialize core services", error=str(e))
        raise
    finally:
        # Shutdown
        logger.info("Shutting down ML service")
        
        # Clean shutdown of services
        try:
            database_svc = await get_database_service()
            await database_svc.close()
            
            consumer = await get_article_consumer()
            query_search_consumer = await get_query_search_consumer()
            await consumer.stop_consuming()
            await query_search_consumer.stop_consuming()
            await consumer.disconnect()
            await query_search_consumer.disconnect()
        except Exception as e:
            logger.error("Error during service shutdown", error=str(e))


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
    vectordb_svc: VectorDBService = Depends(get_vectordb_service),
    database_svc: DatabaseService = Depends(get_database_service),
    consumer = Depends(get_article_consumer),
    query_search_consumer = Depends(get_query_search_consumer)
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
    
    # Check PostgreSQL connectivity
    try:
        database_health = await database_svc.health_check()
        dependencies.append(DependencyHealth(
            name="postgresql",
            status=database_health["status"],
            response_time=database_health.get("response_time"),
            error=database_health.get("error")
        ))
    except Exception as e:
        dependencies.append(DependencyHealth(
            name="postgresql",
            status="unavailable",
            response_time=None,
            error=f"Service not initialized: {str(e)}"
        ))
    
    # Check RabbitMQ consumers status
    try:
        consumer_health = await consumer.health_check()
        dependencies.append(DependencyHealth(
            name="rabbitmq_article_consumer",
            status=consumer_health["status"],
            response_time=None,
            error=consumer_health.get("error")
        ))
    except Exception as e:
        dependencies.append(DependencyHealth(
            name="rabbitmq_article_consumer",
            status="unavailable",
            response_time=None,
            error=f"Service not initialized: {str(e)}"
        ))
    
    try:
        query_search_health = await query_search_consumer.health_check()
        dependencies.append(DependencyHealth(
            name="rabbitmq_query_search_consumer",
            status=query_search_health["status"],
            response_time=None,
            error=query_search_health.get("error")
        ))
    except Exception as e:
        dependencies.append(DependencyHealth(
            name="rabbitmq_query_search_consumer",
            status="unavailable",
            response_time=None,
            error=f"Service not initialized: {str(e)}"
        ))
    
    # Determine overall service status (core services must be healthy)
    core_services = ["openai", "pinecone"]
    core_healthy = all(
        dep.status == "healthy" 
        for dep in dependencies 
        if dep.name in core_services
    )
    service_status = "healthy" if core_healthy else "unhealthy"
    
    return DetailedHealthResponse(
        service=HealthStatus(status=service_status),
        dependencies=dependencies,
        uptime=time.time() - SERVICE_START_TIME
    )
    

@app.get("/index/stats")
@endpoint_error_handler("index statistics retrieval")
async def get_index_stats(
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Get Pinecone index statistics."""
    return await vectordb_svc.get_index_stats()





@app.post("/articles/process", response_model=ArticleProcessingResponse)
@endpoint_error_handler("article processing with chunking")
async def process_article(
    request: ArticleProcessingRequest,
    chunking_svc: ChunkingService = Depends(get_chunking_service),
    database_svc: DatabaseService = Depends(get_database_service),
    embeddings_svc: EmbeddingsService = Depends(get_embeddings_service),
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Process article: chunk content, store in DB, generate embeddings, upload to Pinecone."""
    start_time = time.time()
    
    logger.info("Processing article via HTTP", article_id=str(request.article_id))
    
    # Step 1: Chunk the article
    chunks = await chunking_svc.split_article(request)
    
    # Step 2: Store chunks in PostgreSQL (if available)
    chunk_ids = []
    try:
        chunk_ids = await database_svc.store_chunks(chunks)
        logger.info("Chunks stored in PostgreSQL", count=len(chunk_ids))
    except Exception as e:
        logger.warning("Failed to store chunks in PostgreSQL - continuing without DB storage", error=str(e))
        # Generate fake chunk IDs for Pinecone upload
        chunk_ids = [uuid.uuid4() for _ in chunks]
    
    # Step 3: Generate embeddings for chunks
    embedding_requests = []
    for chunk in chunks:
        embedding_text = f"{request.title}\n\n{chunk.content}"
        embedding_requests.append(EmbeddingRequest(
            article_id=chunk.article_id,
            text=embedding_text
        ))
    
    batch_request = BatchEmbeddingRequest(items=embedding_requests)
    embedding_response = await embeddings_svc.generate_batch_embeddings(batch_request)
    
    # Step 4: Upload embeddings to Pinecone
    upload_count = 0
    for chunk, embedding_result, chunk_id in zip(chunks, embedding_response.results, chunk_ids):
        upload_request = PineconeUploadRequest(
            article_id=chunk.article_id,
            embeddings=embedding_result.embeddings,
            metadata={
                "chunk_id": str(chunk_id),
                "chunk_index": chunk.chunk_index,
                "article_title": request.title,
                "category": request.category,
                "token_count": chunk.token_count,
                "character_count": chunk.character_count,
                "created_at": request.created_at.isoformat()
            }
        )
        
        await vectordb_svc.upload_embedding(upload_request, vector_id=str(chunk_id))
        upload_count += 1
    
    processing_time = time.time() - start_time
    
    return ArticleProcessingResponse(
        article_id=request.article_id,
        chunks_created=len(chunks),
        embeddings_generated=len(embedding_response.results),
        processing_time=processing_time,
        status="completed"
    )


@app.post("/text/chunk", response_model=List[ChunkResponse])
@endpoint_error_handler("text chunking")
async def chunk_text(
    request: ChunkRequest,
    chunking_svc: ChunkingService = Depends(get_chunking_service)
):
    """Split text into chunks using LangChain RecursiveCharacterTextSplitter."""
    from uuid import uuid4
    from .models import ArticleProcessingRequest
    from datetime import datetime
    
    # Create temporary processing request for chunking
    temp_request = ArticleProcessingRequest(
        article_id=uuid4(),
        title="Text Chunking Request",
        content=request.text,
        category="test",
        created_at=datetime.utcnow()
    )
    
    # Use custom chunk parameters from request
    chunks = await chunking_svc.split_article(
        temp_request, 
        chunk_size=request.chunk_size or 500,
        chunk_overlap=request.chunk_overlap or 50
    )
    return chunks


@app.get("/chunks/parameters")
async def get_chunk_parameters(
    chunking_svc: ChunkingService = Depends(get_chunking_service)
):
    """Get current chunking parameters."""
    return await chunking_svc.validate_chunk_parameters()


@app.get("/chunks/{article_id}")
@endpoint_error_handler("chunk retrieval")
async def get_article_chunks(
    article_id: str,
    database_svc: DatabaseService = Depends(get_database_service)
):
    """Get all chunks for a specific article."""
    try:
        from uuid import UUID
        article_uuid = UUID(article_id)
        chunks = await database_svc.get_chunks_by_article(article_uuid)
        return {
            "article_id": article_id,
            "chunks": chunks,
            "total_chunks": len(chunks)
        }
    except ValueError:
        return {"error": "Invalid article ID format"}
    except Exception as e:
        logger.error("Failed to retrieve chunks", article_id=article_id, error=str(e))
        return {"error": f"Failed to retrieve chunks: {str(e)}"}


if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host=settings.ml_service_host,
        port=settings.ml_service_port,
        log_level=settings.log_level,
        reload=True
    )