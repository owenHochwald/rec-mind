"""Health check endpoints."""

import time
from fastapi import APIRouter, Depends

from ..models import HealthStatus, DependencyHealth, DetailedHealthResponse
from ..embeddings import get_embeddings_service, EmbeddingsService
from ..vectordb import get_vectordb_service, VectorDBService
from ..database import get_database_service, DatabaseService
from ..mq_consumer import get_article_consumer
from ..query_search_consumer import get_query_search_consumer

router = APIRouter()

# Service startup time for uptime calculation
SERVICE_START_TIME = time.time()


@router.get("/health", response_model=HealthStatus)
async def health_check():
    """Basic health check endpoint."""
    return HealthStatus(status="healthy")


@router.get("/health/detail", response_model=DetailedHealthResponse)
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