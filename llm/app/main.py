"""FastAPI ML Service for embeddings and vector operations."""

import asyncio
import uuid
from contextlib import asynccontextmanager

import structlog
import uvicorn
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse

from .config import get_settings
from .models import ErrorResponse
from .embeddings import get_embeddings_service
from .vectordb import get_vectordb_service
from .chunking import get_chunking_service
from .database import get_database_service
from .mq_consumer import get_article_consumer, start_article_processing
from .query_search_consumer import get_query_search_consumer, start_query_search_processing
from .routers import health, vectordb, articles, text, embeddings

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


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan events."""
    logger.info("Starting ML service", version="1.0.0")
    
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
        logger.info("Shutting down ML service")
        
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

# Include routers
app.include_router(health.router, tags=["health"])
app.include_router(vectordb.router, tags=["vectordb"])
app.include_router(articles.router, tags=["articles"])
app.include_router(text.router, tags=["text"])
app.include_router(embeddings.router, tags=["embeddings"])


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


if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host=settings.ml_service_host,
        port=settings.ml_service_port,
        log_level=settings.log_level,
        reload=True
    )