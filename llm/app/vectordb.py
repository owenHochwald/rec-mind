"""Pinecone vector database operations with direct host connection."""

import time
from typing import List, Dict, Any, Optional
from uuid import UUID

import pinecone
import structlog
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

from .config import get_settings
from .models import (
    PineconeUploadRequest, 
    PineconeSearchRequest, 
    PineconeSearchResponse, 
    SearchResult
)


logger = structlog.get_logger(__name__)
settings = get_settings()


class VectorDBService:
    """Service for Pinecone vector database operations using direct host connection."""
    
    def __init__(self):
        self.api_key = settings.pinecone_api_key
        self.environment = settings.pinecone_environment
        self.host = settings.pinecone_host
        self.index_name = settings.pinecone_index_name
        self.dimension = settings.embedding_dimensions
        self._index = None
        self._initialized = False
    
    async def _initialize(self):
        """Initialize Pinecone connection using modern SDK approach."""
        if self._initialized:
            return
            
        try:
            logger.info("Initializing Pinecone connection", 
                       index_name=self.index_name,
                       host_configured=bool(self.host))
            
            if not self.host:
                raise ValueError("PINECONE_HOST must be configured for direct connection")
            
            logger.info("Connecting to Pinecone index", 
                       host=self.host, 
                       index_name=self.index_name)
            
            # Initialize Pinecone client with host configuration (modern SDK v7+ approach)
            pc = pinecone.Pinecone(api_key=self.api_key, host=self.host)
            
            # Connect to index by name - host is already configured in client
            self._index = pc.Index(self.index_name)
            
            self._initialized = True
            logger.info("Pinecone connection established successfully", 
                       index_name=self.index_name,
                       host=self.host)
            
        except Exception as e:
            logger.error("Failed to initialize Pinecone connection", error=str(e))
            raise
    
    @retry(
        retry=retry_if_exception_type((pinecone.exceptions.PineconeException,)),
        stop=stop_after_attempt(settings.max_retries),
        wait=wait_exponential(multiplier=1, min=2, max=8),
        reraise=True
    )
    async def upload_embedding(self, request: PineconeUploadRequest) -> Dict[str, Any]:
        """Upload embedding vector to Pinecone with retry logic."""
        await self._initialize()
        
        try:
            start_time = time.time()
            
            # Prepare vector data
            vector_data = {
                "id": str(request.article_id),
                "values": request.embeddings,
                "metadata": {
                    **request.metadata,
                    "article_id": str(request.article_id),
                    "uploaded_at": time.time()
                }
            }
            
            logger.info(
                "Uploading vector to Pinecone",
                article_id=str(request.article_id),
                dimension=len(request.embeddings)
            )
            
            # Upload vector
            upsert_response = self._index.upsert([vector_data])
            
            upload_time = time.time() - start_time
            
            logger.info(
                "Vector uploaded successfully",
                article_id=str(request.article_id),
                upload_time=upload_time,
                upserted_count=upsert_response.upserted_count
            )
            
            return {
                "article_id": str(request.article_id),
                "status": "uploaded",
                "upload_time": upload_time,
                "upserted_count": upsert_response.upserted_count
            }
            
        except Exception as e:
            logger.error(
                "Failed to upload vector",
                article_id=str(request.article_id),
                error=str(e)
            )
            raise
    
    @retry(
        retry=retry_if_exception_type((pinecone.exceptions.PineconeException,)),
        stop=stop_after_attempt(settings.max_retries),
        wait=wait_exponential(multiplier=1, min=2, max=8),
        reraise=True
    )
    async def search_similar(self, request: PineconeSearchRequest) -> PineconeSearchResponse:
        """Search for similar vectors in Pinecone."""
        await self._initialize()
        
        try:
            start_time = time.time()
            
            logger.info(
                "Searching for similar vectors",
                article_id=str(request.article_id),
                top_k=request.top_k,
                threshold=request.score_threshold
            )
            
            # First, get the query vector for the article
            query_response = self._index.fetch([str(request.article_id)])
            
            if not query_response.vectors:
                raise ValueError(f"Article {request.article_id} not found in vector database")
            
            query_vector = query_response.vectors[str(request.article_id)]["values"]
            
            # Perform similarity search
            search_response = self._index.query(
                vector=query_vector,
                top_k=request.top_k + 1,  # +1 to exclude self
                include_metadata=request.include_metadata,
                include_values=False
            )
            
            # Filter out the query article itself and apply threshold
            results = []
            for match in search_response.matches:
                if (match.id != str(request.article_id) and 
                    match.score >= request.score_threshold):
                    
                    result = SearchResult(
                        id=match.id,
                        score=match.score,
                        metadata=match.metadata if request.include_metadata else None
                    )
                    results.append(result)
            
            # Limit to requested number
            results = results[:request.top_k]
            
            search_time = time.time() - start_time
            
            logger.info(
                "Similar vectors search completed",
                article_id=str(request.article_id),
                results_found=len(results),
                search_time=search_time
            )
            
            return PineconeSearchResponse(
                query_article_id=request.article_id,
                results=results,
                total_found=len(results),
                processing_time=search_time
            )
            
        except Exception as e:
            logger.error(
                "Failed to search similar vectors",
                article_id=str(request.article_id),
                error=str(e)
            )
            raise
    
    async def delete_vector(self, article_id: UUID) -> Dict[str, Any]:
        """Delete vector from Pinecone."""
        await self._initialize()
        
        try:
            logger.info("Deleting vector", article_id=str(article_id))
            
            delete_response = self._index.delete([str(article_id)])
            
            logger.info("Vector deleted successfully", article_id=str(article_id))
            
            return {
                "article_id": str(article_id),
                "status": "deleted"
            }
            
        except Exception as e:
            logger.error("Failed to delete vector", article_id=str(article_id), error=str(e))
            raise
    
    async def get_index_stats(self) -> Dict[str, Any]:
        """Get Pinecone index statistics."""
        await self._initialize()
        
        try:
            stats = self._index.describe_index_stats()
            
            return {
                "total_vectors": stats.total_vector_count,
                "dimension": stats.dimension,
                "index_fullness": stats.index_fullness,
                "namespaces": dict(stats.namespaces) if stats.namespaces else {}
            }
            
        except Exception as e:
            logger.error("Failed to get index stats", error=str(e))
            raise
    
    async def health_check(self) -> Dict[str, Any]:
        """Check Pinecone connectivity and index health."""
        try:
            start_time = time.time()
            
            await self._initialize()
            
            # Simple health check by getting index stats
            stats = await self.get_index_stats()
            
            response_time = time.time() - start_time
            
            return {
                "status": "healthy",
                "response_time": response_time,
                "index_name": self.index_name,
                "total_vectors": stats["total_vectors"],
                "dimension": stats["dimension"]
            }
            
        except Exception as e:
            logger.error("Pinecone health check failed", error=str(e))
            return {
                "status": "unhealthy",
                "error": str(e),
                "index_name": self.index_name
            }


# Global service instance
vectordb_service = VectorDBService()


async def get_vectordb_service() -> VectorDBService:
    """Dependency injection for vector database service."""
    return vectordb_service