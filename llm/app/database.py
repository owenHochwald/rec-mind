"""PostgreSQL database connection and operations for chunk storage."""

import asyncio
import time
from typing import List, Dict, Any, Optional
from uuid import UUID
from datetime import datetime

import structlog
import asyncpg
from asyncpg import Pool, Connection

from .config import get_settings
from .models import ChunkResponse

logger = structlog.get_logger(__name__)
settings = get_settings()


class DatabaseService:
    """Service for PostgreSQL database operations."""
    
    def __init__(self):
        self.pool: Optional[Pool] = None
        
    async def initialize(self):
        """Initialize database connection pool."""
        try:
            logger.info("Initializing PostgreSQL connection pool")
            
            self.pool = await asyncpg.create_pool(
                host=settings.db_host,
                port=settings.db_port,
                user=settings.db_user,
                password=settings.db_password,
                database=settings.db_name,
                min_size=1,
                max_size=10,
                command_timeout=60
            )
            
            # Test connection
            async with self.pool.acquire() as conn:
                await conn.fetch("SELECT 1")
            
            logger.info("PostgreSQL connection pool initialized successfully")
            
        except Exception as e:
            logger.error("Failed to initialize PostgreSQL connection pool", error=str(e))
            raise
    
    async def close(self):
        """Close database connection pool."""
        if self.pool:
            await self.pool.close()
            logger.info("PostgreSQL connection pool closed")
    
    async def store_chunks(self, chunks: List[ChunkResponse]) -> List[UUID]:
        """Store article chunks in PostgreSQL."""
        if not self.pool:
            logger.error("Database pool not initialized - cannot store chunks")
            raise RuntimeError("Database pool not initialized")
        
        if not chunks:
            logger.warning("No chunks provided to store")
            return []
        
        chunk_ids = []
        
        try:
            logger.info("Acquiring database connection for chunk storage")
            async with self.pool.acquire() as conn:
                logger.info("Starting database transaction for chunks", chunks_count=len(chunks))
                async with conn.transaction():
                    for i, chunk in enumerate(chunks):
                        logger.debug(f"Inserting chunk {i+1}/{len(chunks)}", 
                                   chunk_index=chunk.chunk_index, 
                                   content_length=len(chunk.content))
                        
                        # Insert chunk and return generated ID
                        result = await conn.fetchrow("""
                            INSERT INTO article_chunks (article_id, chunk_index, content, token_count, character_count, pinecone_id, created_at)
                            VALUES ($1, $2, $3, $4, $5, $6, NOW())
                            RETURNING id
                        """, 
                        chunk.article_id, 
                        chunk.chunk_index, 
                        chunk.content,
                        chunk.token_count,
                        chunk.character_count,
                        chunk.pinecone_id
                        )
                        
                        if result and 'id' in result:
                            chunk_ids.append(result['id'])
                            logger.debug(f"Chunk {i+1} stored with ID: {result['id']}")
                        else:
                            logger.error(f"Failed to get ID for chunk {i+1}")
            
            logger.info(
                "All chunks stored successfully in PostgreSQL",
                article_id=str(chunks[0].article_id) if chunks else None,
                chunks_stored=len(chunk_ids),
                expected_chunks=len(chunks)
            )
            
            return chunk_ids
            
        except Exception as e:
            logger.error(
                "Failed to store chunks in PostgreSQL",
                article_id=str(chunks[0].article_id) if chunks else None,
                error=str(e),
                error_type=type(e).__name__,
                chunks_attempted=len(chunks)
            )
            raise
    
    async def update_chunk_pinecone_id(self, chunk_id: UUID, pinecone_id: str) -> bool:
        """Update the pinecone_id for a specific chunk."""
        if not self.pool:
            raise RuntimeError("Database pool not initialized")
        
        try:
            async with self.pool.acquire() as conn:
                result = await conn.execute("""
                    UPDATE article_chunks 
                    SET pinecone_id = $1 
                    WHERE id = $2
                """, pinecone_id, chunk_id)
                
                if result == "UPDATE 1":
                    logger.debug(f"Updated chunk {chunk_id} with pinecone_id: {pinecone_id}")
                    return True
                else:
                    logger.warning(f"No chunk found with ID {chunk_id} to update")
                    return False
                    
        except Exception as e:
            logger.error("Failed to update chunk pinecone_id", chunk_id=str(chunk_id), error=str(e))
            raise

    async def get_chunk_by_id(self, chunk_id: UUID) -> Optional[Dict[str, Any]]:
        """Retrieve a specific chunk by ID."""
        if not self.pool:
            raise RuntimeError("Database pool not initialized")
        
        try:
            async with self.pool.acquire() as conn:
                row = await conn.fetchrow("""
                    SELECT id, article_id, chunk_index, content, token_count, character_count, pinecone_id, created_at
                    FROM article_chunks 
                    WHERE id = $1
                """, chunk_id)
                
                if row:
                    chunk = {
                        "id": row['id'],
                        "article_id": row['article_id'],
                        "chunk_index": row['chunk_index'],
                        "content": row['content'],
                        "token_count": row['token_count'],
                        "character_count": row['character_count'],
                        "pinecone_id": row['pinecone_id'],
                        "created_at": row['created_at']
                    }
                    
                    logger.debug("Retrieved chunk by ID", chunk_id=str(chunk_id))
                    return chunk
                else:
                    logger.warning("Chunk not found", chunk_id=str(chunk_id))
                    return None
                
        except Exception as e:
            logger.error("Failed to retrieve chunk by ID", chunk_id=str(chunk_id), error=str(e))
            raise

    async def get_chunks_by_article(self, article_id: UUID) -> List[Dict[str, Any]]:
        """Retrieve all chunks for a specific article."""
        if not self.pool:
            raise RuntimeError("Database pool not initialized")
        
        try:
            async with self.pool.acquire() as conn:
                rows = await conn.fetch("""
                    SELECT id, article_id, chunk_index, content, token_count, character_count, pinecone_id, created_at
                    FROM article_chunks 
                    WHERE article_id = $1
                    ORDER BY chunk_index
                """, article_id)
                
                chunks = []
                for row in rows:
                    chunks.append({
                        "id": row['id'],
                        "article_id": row['article_id'],
                        "chunk_index": row['chunk_index'],
                        "content": row['content'],
                        "token_count": row['token_count'],
                        "character_count": row['character_count'],
                        "pinecone_id": row['pinecone_id'],
                        "created_at": row['created_at']
                    })
                
                logger.info(
                    "Retrieved chunks for article",
                    article_id=str(article_id),
                    chunks_found=len(chunks)
                )
                
                return chunks
                
        except Exception as e:
            logger.error(
                "Failed to retrieve chunks",
                article_id=str(article_id), 
                error=str(e)
            )
            raise
    
    async def health_check(self) -> Dict[str, Any]:
        """Check database connectivity and performance."""
        try:
            start_time = time.time()
            
            if not self.pool:
                return {
                    "status": "unhealthy",
                    "error": "Database pool not initialized"
                }
            
            async with self.pool.acquire() as conn:
                # Test basic connectivity
                result = await conn.fetchrow("SELECT version(), current_timestamp")
                
                # Test article_chunks table access
                count_result = await conn.fetchrow("SELECT COUNT(*) as count FROM article_chunks")
                
            response_time = time.time() - start_time
            
            return {
                "status": "healthy",
                "response_time": response_time,
                "postgresql_version": result['version'],
                "chunks_in_db": count_result['count'],
                "pool_size": self.pool.get_size(),
                "pool_free_size": self.pool.get_idle_size()
            }
            
        except Exception as e:
            logger.error("Database health check failed", error=str(e))
            return {
                "status": "unhealthy",
                "error": str(e)
            }


# Global service instance
database_service = DatabaseService()


async def get_database_service() -> DatabaseService:
    """Dependency injection for database service."""
    return database_service