"""RabbitMQ consumer for query-based semantic search."""

import asyncio
import json
import os
import time
from typing import Dict, Any, Optional
from uuid import UUID

import structlog
import aio_pika
from aio_pika import connect_robust, IncomingMessage
from aio_pika.abc import AbstractConnection, AbstractChannel, AbstractQueue
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

from .config import get_settings
from .models import (
    QuerySearchMessage, QuerySearchResponse, QuerySearchResult, 
    QuerySearchError, EmbeddingRequest
)
from .database import get_database_service
from .embeddings import get_embeddings_service
from .vectordb import get_vectordb_service

logger = structlog.get_logger(__name__)
settings = get_settings()


class QuerySearchConsumer:
    """RabbitMQ consumer for query-based semantic search."""
    
    def __init__(self):
        self.connection: AbstractConnection = None
        self.channel: AbstractChannel = None
        self.search_queue: AbstractQueue = None
        self.results_queue: AbstractQueue = None
        self.is_running = False
        self.instance_id = os.getenv("ML_SERVICE_INSTANCE_ID", "query_search_default")
        
        # Circuit breaker state
        self.failure_count = 0
        self.max_failures = 5
        self.circuit_open_time = None
        self.circuit_timeout = 60  # seconds
        
    async def connect(self):
        """Establish RabbitMQ connection and declare queues."""
        try:
            logger.info("Connecting to RabbitMQ for query search", 
                       instance_id=self.instance_id, host=settings.rabbitmq_host)
            
            connection_url = f"amqp://{settings.rabbitmq_user}:{settings.rabbitmq_password}@{settings.rabbitmq_host}:{settings.rabbitmq_port}{settings.rabbitmq_vhost}"
            self.connection = await connect_robust(connection_url)
            
            self.channel = await self.connection.channel()
            await self.channel.set_qos(prefetch_count=1)
            
            # Declare search queue (consume from)
            self.search_queue = await self.channel.declare_queue(
                "query_search", durable=True
            )
            
            # Declare results queue (publish to)
            self.results_queue = await self.channel.declare_queue(
                "query_search_results", durable=True
            )
            
            logger.info("RabbitMQ connection established for query search", 
                       instance_id=self.instance_id)
            
        except Exception as e:
            logger.error("Failed to connect to RabbitMQ for query search", 
                        instance_id=self.instance_id, error=str(e))
            raise
    
    async def disconnect(self):
        """Close RabbitMQ connection."""
        try:
            if self.connection and not self.connection.is_closed:
                await self.connection.close()
                logger.info("RabbitMQ connection closed", instance_id=self.instance_id)
        except Exception as e:
            logger.error("Error closing RabbitMQ connection", 
                        instance_id=self.instance_id, error=str(e))
    
    def is_circuit_open(self) -> bool:
        """Check if circuit breaker is open."""
        if self.circuit_open_time is None:
            return False
        
        if time.time() - self.circuit_open_time > self.circuit_timeout:
            # Reset circuit breaker after timeout
            self.circuit_open_time = None
            self.failure_count = 0
            logger.info("Circuit breaker reset", instance_id=self.instance_id)
            return False
        
        return True

    def record_success(self):
        """Record successful operation."""
        if self.failure_count > 0:
            self.failure_count = 0
            logger.info("Circuit breaker success recorded", instance_id=self.instance_id)

    def record_failure(self):
        """Record failed operation and potentially open circuit."""
        self.failure_count += 1
        if self.failure_count >= self.max_failures and self.circuit_open_time is None:
            self.circuit_open_time = time.time()
            logger.error("Circuit breaker opened due to consecutive failures", 
                        failure_count=self.failure_count, instance_id=self.instance_id)

    async def publish_search_results(self, search_response: QuerySearchResponse):
        """Publish search results to results queue."""
        try:
            message_body = search_response.json()
            await self.channel.default_exchange.publish(
                aio_pika.Message(
                    message_body.encode(),
                    delivery_mode=aio_pika.DeliveryMode.PERSISTENT
                ),
                routing_key="query_search_results"
            )
            logger.info("Published query search results", 
                       search_id=str(search_response.search_id),
                       results_count=len(search_response.results),
                       query=search_response.query[:50] + "..." if len(search_response.query) > 50 else search_response.query,
                       instance_id=self.instance_id)
        except Exception as e:
            logger.error("Failed to publish query search results", 
                        search_id=str(search_response.search_id),
                        instance_id=self.instance_id, error=str(e))
    
    async def publish_search_error(self, search_id: UUID, query: str, error_message: str, error_code: str = "SEARCH_ERROR"):
        """Publish search error to results queue."""
        try:
            error_response = QuerySearchError(
                search_id=search_id,
                query=query,
                error_message=error_message,
                error_code=error_code,
                service_instance_id=self.instance_id
            )
            
            message_body = error_response.json()
            await self.channel.default_exchange.publish(
                aio_pika.Message(
                    message_body.encode(),
                    delivery_mode=aio_pika.DeliveryMode.PERSISTENT
                ),
                routing_key="query_search_results"
            )
            logger.error("Published query search error", 
                        search_id=str(search_id),
                        query=query[:50] + "..." if len(query) > 50 else query,
                        instance_id=self.instance_id, error=error_message)
        except Exception as e:
            logger.error("Failed to publish query search error", 
                        search_id=str(search_id),
                        instance_id=self.instance_id, error=str(e))
    
    async def process_query_search(self, message: IncomingMessage):
        """Process query-based search request."""
        async with message.process():
            start_time = time.time()
            search_message = None
            
            try:
                # Check circuit breaker
                if self.is_circuit_open():
                    logger.warning("Circuit breaker is open, rejecting request", 
                                  instance_id=self.instance_id)
                    return
                
                # Parse message
                message_data = json.loads(message.body.decode())
                search_message = QuerySearchMessage(**message_data)
                
                logger.info("Processing query search", 
                           search_id=str(search_message.search_id),
                           query=search_message.query[:50] + "..." if len(search_message.query) > 50 else search_message.query,
                           instance_id=self.instance_id)
                
                # Generate embedding for query
                embeddings_svc = await get_embeddings_service()
                embedding_request = EmbeddingRequest(
                    article_id=search_message.search_id,  # Use search_id as placeholder
                    text=search_message.query
                )
                embedding_response = await embeddings_svc.generate_embedding(embedding_request)
                
                # Search Pinecone for similar chunks (no exclusions for query search)
                vectordb_svc = await get_vectordb_service()
                pinecone_results = await vectordb_svc.search_similar_vectors(
                    embedding=embedding_response.embeddings,
                    top_k=search_message.max_results,
                    score_threshold=search_message.score_threshold,
                    exclude_article_id=None  # No exclusions for query-based search
                )
                
                # Convert Pinecone results to structured format
                search_results = []
                for result in pinecone_results.get("matches", []):
                    metadata = result.get("metadata", {})
                    
                    # Extract chunk content preview (first 200 chars)
                    content_preview = metadata.get("content_preview", "")
                    if not content_preview:
                        # If no preview in metadata, truncate chunk content
                        content = metadata.get("content", "")
                        content_preview = content[:200] + "..." if len(content) > 200 else content
                    
                    query_result = QuerySearchResult(
                        chunk_id=metadata.get("chunk_id", result["id"]),
                        similarity_score=result["score"],
                        article_id=UUID(metadata.get("article_id", str(search_message.search_id))),
                        chunk_index=metadata.get("chunk_index", 0),
                        article_title=metadata.get("article_title", "Unknown"),
                        category=metadata.get("category", "Unknown"),
                        content_preview=content_preview,
                        url=metadata.get("url", "")
                    )
                    search_results.append(query_result)
                
                processing_time = time.time() - start_time
                
                # Create response
                search_response = QuerySearchResponse(
                    search_id=search_message.search_id,
                    query=search_message.query,
                    results=search_results,
                    total_found=len(search_results),
                    processing_time=processing_time,
                    service_instance_id=self.instance_id
                )
                
                # Publish results
                await self.publish_search_results(search_response)
                self.record_success()
                
                logger.info("Query search completed successfully",
                           search_id=str(search_message.search_id),
                           query=search_message.query[:50] + "..." if len(search_message.query) > 50 else search_message.query,
                           results_found=len(search_results),
                           processing_time=processing_time,
                           instance_id=self.instance_id)
                
            except Exception as e:
                self.record_failure()
                error_msg = f"Query search failed: {str(e)}"
                
                if search_message:
                    await self.publish_search_error(
                        search_message.search_id,
                        search_message.query,
                        error_msg,
                        "PROCESSING_ERROR"
                    )
                
                logger.error("Failed to process query search",
                           search_id=str(search_message.search_id) if search_message else "unknown",
                           query=search_message.query[:50] + "..." if search_message and len(search_message.query) > 50 else search_message.query if search_message else "unknown",
                           instance_id=self.instance_id,
                           error=str(e))

    async def start_consuming(self):
        """Start consuming messages from the query search queue."""
        try:
            self.is_running = True
            
            async with self.search_queue.iterator() as queue_iter:
                logger.info("Started consuming query search messages", instance_id=self.instance_id)
                
                async for message in queue_iter:
                    if not self.is_running:
                        break
                    
                    await self.process_query_search(message)
                    
        except Exception as e:
            logger.error("Error in query search consumer", instance_id=self.instance_id, error=str(e))
            raise
    
    async def stop_consuming(self):
        """Stop consuming messages."""
        self.is_running = False
        logger.info("Stopped consuming query search messages", instance_id=self.instance_id)
    
    async def health_check(self) -> Dict[str, Any]:
        """Check consumer health."""
        try:
            if not self.connection or self.connection.is_closed:
                return {"status": "unhealthy", "error": "No connection"}
            
            if self.is_circuit_open():
                return {"status": "degraded", "error": "Circuit breaker open"}
            
            return {"status": "healthy"}
            
        except Exception as e:
            return {"status": "unhealthy", "error": str(e)}


# Singleton instance
_query_search_consumer = None

async def get_query_search_consumer() -> QuerySearchConsumer:
    """Get or create the query search consumer instance."""
    global _query_search_consumer
    if _query_search_consumer is None:
        _query_search_consumer = QuerySearchConsumer()
    return _query_search_consumer

async def start_query_search_processing():
    """Start the query search processing background task."""
    consumer = await get_query_search_consumer()
    await consumer.start_consuming()