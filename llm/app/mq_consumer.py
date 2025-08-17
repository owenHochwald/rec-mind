"""RabbitMQ consumer for article processing with chunking and embeddings."""

import asyncio
import json
import time
from typing import Dict, Any

import structlog
import aio_pika
from aio_pika import connect_robust, IncomingMessage
from aio_pika.abc import AbstractConnection, AbstractChannel, AbstractQueue

from .config import get_settings
from .models import ArticleProcessingRequest, BatchEmbeddingRequest, EmbeddingRequest, PineconeUploadRequest
from .chunking import get_chunking_service
from .database import get_database_service
from .embeddings import get_embeddings_service
from .vectordb import get_vectordb_service

logger = structlog.get_logger(__name__)
settings = get_settings()


class ArticleProcessingConsumer:
    """RabbitMQ consumer for processing articles with chunking and embeddings."""
    
    def __init__(self):
        self.connection: AbstractConnection = None
        self.channel: AbstractChannel = None
        self.queue: AbstractQueue = None
        self.is_running = False
        
    async def connect(self):
        """Establish RabbitMQ connection and declare queue."""
        try:
            logger.info("Connecting to RabbitMQ", host=settings.rabbitmq_host)
            
            # Create connection
            connection_url = f"amqp://{settings.rabbitmq_user}:{settings.rabbitmq_password}@{settings.rabbitmq_host}:{settings.rabbitmq_port}{settings.rabbitmq_vhost}"
            self.connection = await connect_robust(connection_url)
            
            # Create channel
            self.channel = await self.connection.channel()
            await self.channel.set_qos(prefetch_count=1)  # Process one message at a time
            
            # Declare queue
            self.queue = await self.channel.declare_queue(
                "article_processing",
                durable=True
            )
            
            logger.info("RabbitMQ connection established successfully")
            
        except Exception as e:
            logger.error("Failed to connect to RabbitMQ", error=str(e))
            raise
    
    async def disconnect(self):
        """Close RabbitMQ connection."""
        try:
            if self.connection and not self.connection.is_closed:
                await self.connection.close()
                logger.info("RabbitMQ connection closed")
        except Exception as e:
            logger.error("Error closing RabbitMQ connection", error=str(e))
    
    async def process_article_message(self, message: IncomingMessage):
        """Process individual article message with chunking and embeddings."""
        async with message.process():
            try:
                start_time = time.time()
                
                # Parse message
                message_data = json.loads(message.body.decode())
                logger.info("Processing article message", message_data=message_data)
                
                # Create processing request
                processing_request = ArticleProcessingRequest(**message_data)
                
                # Get services
                chunking_svc = await get_chunking_service()
                database_svc = await get_database_service()
                embeddings_svc = await get_embeddings_service()
                vectordb_svc = await get_vectordb_service()
                
                # Step 1: Chunk the article content
                logger.info("Starting article chunking", article_id=str(processing_request.article_id))
                chunks = await chunking_svc.split_article(processing_request)
                
                # Step 2: Store chunks in PostgreSQL
                logger.info("Storing chunks in database", chunks_count=len(chunks))
                try:
                    chunk_ids = await database_svc.store_chunks(chunks)
                    logger.info("Successfully stored chunks in PostgreSQL", chunk_ids_count=len(chunk_ids))
                except Exception as db_error:
                    logger.error("Failed to store chunks in PostgreSQL", error=str(db_error))
                    # Generate fake chunk IDs so we can still proceed with embeddings
                    from uuid import uuid4
                    chunk_ids = [uuid4() for _ in chunks]
                    logger.warning("Using generated chunk IDs for Pinecone upload", generated_count=len(chunk_ids))
                
                # Step 3: Generate embeddings for each chunk
                logger.info("Generating embeddings for chunks", chunks_count=len(chunks))
                embedding_requests = []
                for chunk in chunks:
                    # Combine title context with chunk content for better embeddings
                    embedding_text = f"{processing_request.title}\n\n{chunk.content}"
                    embedding_requests.append(EmbeddingRequest(
                        article_id=chunk.article_id,
                        text=embedding_text
                    ))
                
                # Batch process embeddings
                batch_request = BatchEmbeddingRequest(items=embedding_requests)
                embedding_response = await embeddings_svc.generate_batch_embeddings(batch_request)
                
                # Step 4: Upload embeddings to Pinecone with chunk metadata
                logger.info("Uploading embeddings to Pinecone", embeddings_count=len(embedding_response.results))
                upload_count = 0
                for i, (chunk, embedding_result, chunk_id) in enumerate(zip(chunks, embedding_response.results, chunk_ids)):
                    # Generate unique Pinecone vector ID
                    pinecone_vector_id = str(chunk_id)  # Use chunk_id as Pinecone vector ID
                    
                    upload_request = PineconeUploadRequest(
                        article_id=chunk.article_id,
                        embeddings=embedding_result.embeddings,
                        metadata={
                            "chunk_id": str(chunk_id),
                            "chunk_index": chunk.chunk_index,
                            "article_title": processing_request.title,
                            "category": processing_request.category,
                            "token_count": chunk.token_count,
                            "character_count": chunk.character_count,
                            "created_at": processing_request.created_at.isoformat(),
                            "content_preview": chunk.content[:100] + "..." if len(chunk.content) > 100 else chunk.content
                        }
                    )
                    
                    # Upload to Pinecone with specific vector ID
                    await vectordb_svc.upload_embedding(upload_request, vector_id=pinecone_vector_id)
                    
                    # Update the chunk in PostgreSQL with the Pinecone vector ID
                    try:
                        await database_svc.update_chunk_pinecone_id(chunk_id, pinecone_vector_id)
                        logger.debug(f"Updated chunk {chunk_id} with pinecone_id: {pinecone_vector_id}")
                    except Exception as update_error:
                        logger.error(f"Failed to update chunk {chunk_id} with pinecone_id", error=str(update_error))
                        # Don't fail the entire process for this
                    
                    upload_count += 1
                
                processing_time = time.time() - start_time
                
                logger.info(
                    "Article processing completed successfully",
                    article_id=str(processing_request.article_id),
                    chunks_created=len(chunks),
                    embeddings_generated=len(embedding_response.results),
                    vectors_uploaded=upload_count,
                    processing_time=processing_time
                )
                
            except Exception as e:
                logger.error(
                    "Article processing failed",
                    error=str(e),
                    message_body=message.body.decode()
                )
                # Message will be rejected and potentially requeued based on RabbitMQ configuration
                raise
    
    async def start_consuming(self):
        """Start consuming messages from the article_processing queue."""
        if not self.connection or self.connection.is_closed:
            await self.connect()
        
        try:
            logger.info("Starting article processing consumer")
            self.is_running = True
            
            # Start consuming messages
            await self.queue.consume(self.process_article_message)
            
            logger.info("Article processing consumer started successfully")
            
            # Keep the consumer running
            while self.is_running:
                await asyncio.sleep(1)
                
        except Exception as e:
            logger.error("Consumer error", error=str(e))
            raise
        finally:
            await self.disconnect()
    
    async def stop_consuming(self):
        """Stop consuming messages."""
        logger.info("Stopping article processing consumer")
        self.is_running = False
    
    async def health_check(self) -> Dict[str, Any]:
        """Check RabbitMQ consumer health."""
        try:
            if not self.connection or self.connection.is_closed:
                return {
                    "status": "unhealthy",
                    "error": "RabbitMQ connection not established"
                }
            
            return {
                "status": "healthy",
                "connection_state": "open" if not self.connection.is_closed else "closed",
                "is_consuming": self.is_running,
                "queue_name": "article_processing"
            }
            
        except Exception as e:
            return {
                "status": "unhealthy",
                "error": str(e)
            }


# Global consumer instance
article_consumer = ArticleProcessingConsumer()


async def get_article_consumer() -> ArticleProcessingConsumer:
    """Dependency injection for article consumer."""
    return article_consumer


async def start_article_processing():
    """Start the article processing consumer as a background task."""
    try:
        await article_consumer.start_consuming()
    except Exception as e:
        logger.error("Failed to start article processing consumer", error=str(e))
        raise