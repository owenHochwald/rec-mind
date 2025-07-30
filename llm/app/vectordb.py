"""Simplified Pinecone vector database operations."""

import time
from typing import List, Dict, Any, Optional
from uuid import UUID

import structlog
from pinecone import Pinecone, ServerlessSpec
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
    """Simplified Pinecone vector database service."""
    
    def __init__(self):
        self.api_key = settings.pinecone_api_key
        self.index_name = settings.pinecone_index_name
        self.dimension = settings.embedding_dimensions
        self._pc = None
        self._index = None
        
    def _get_client(self):
        """Get or create Pinecone client."""
        if not self._pc:
            self._pc = Pinecone(api_key=self.api_key)
        return self._pc
    
    def _get_index(self):
        """Get or create index connection."""
        if not self._index:
            pc = self._get_client()
            
            # Create index if it doesn't exist
            if not pc.has_index(self.index_name):
                logger.info(f"Creating index {self.index_name}")
                pc.create_index(
                    name=self.index_name,
                    dimension=self.dimension,
                    metric="cosine",
                    spec=ServerlessSpec(
                        cloud="aws",
                        region="us-east-1"
                    )
                )
                logger.info(f"Index {self.index_name} created successfully")
            
            # Initialize index client
            self._index = pc.Index(name=self.index_name)
            logger.info(f"Connected to index {self.index_name}")
            
        return self._index
    
    @retry(
        stop=stop_after_attempt(settings.max_retries),
        wait=wait_exponential(multiplier=1, min=2, max=8),
        reraise=True
    )
    async def upload_embedding(self, request: PineconeUploadRequest) -> Dict[str, Any]:
        """Upload embedding vector to Pinecone."""
        try:
            start_time = time.time()
            index = self._get_index()
            
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
            upsert_response = index.upsert([vector_data])
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
        stop=stop_after_attempt(settings.max_retries),
        wait=wait_exponential(multiplier=1, min=2, max=8),
        reraise=True
    )
    async def search_similar(self, request: PineconeSearchRequest) -> PineconeSearchResponse:
        """Search for similar vectors in Pinecone."""
        try:
            start_time = time.time()
            index = self._get_index()
            
            logger.info(
                "Searching for similar vectors",
                article_id=str(request.article_id),
                top_k=request.top_k,
                threshold=request.score_threshold
            )
            
            # First, get the query vector for the article
            query_response = index.fetch([str(request.article_id)])
            
            if not query_response.vectors:
                raise ValueError(f"Article {request.article_id} not found in vector database")
            
            query_vector = query_response.vectors[str(request.article_id)]["values"]
            
            # Perform similarity search
            search_response = index.query(
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
        try:
            index = self._get_index()
            logger.info("Deleting vector", article_id=str(article_id))
            
            index.delete([str(article_id)])
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
        try:
            index = self._get_index()
            stats = index.describe_index_stats()
            
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
    
    async def test_connection(self) -> Dict[str, Any]:
        """Test Pinecone connection with basic operations."""
        try:
            start_time = time.time()
            test_results = {}
            
            # Test 1: Get index and stats
            logger.info("Testing Pinecone connection")
            index = self._get_index()
            test_results["connection"] = "success"
            
            stats = await self.get_index_stats()
            test_results["index_stats"] = {
                "status": "success",
                "total_vectors": stats["total_vectors"],
                "dimension": stats["dimension"]
            }
            
            # Test 2: Basic operations with dummy data
            test_vector_id = "test-connection-vector"
            test_vector = [0.1] * self.dimension
            
            try:
                # Test upsert
                logger.info("Testing vector operations")
                upsert_result = index.upsert([{
                    "id": test_vector_id,
                    "values": test_vector,
                    "metadata": {"test": True, "timestamp": time.time()}
                }])
                test_results["upsert"] = {
                    "status": "success",
                    "count": upsert_result.upserted_count
                }
                
                # Test fetch
                fetch_result = index.fetch([test_vector_id])
                test_results["fetch"] = {
                    "status": "success" if fetch_result.vectors else "failed",
                    "found": len(fetch_result.vectors)
                }
                
                # Test query
                query_result = index.query(
                    vector=test_vector,
                    top_k=1,
                    include_metadata=True
                )
                test_results["query"] = {
                    "status": "success",
                    "matches": len(query_result.matches)
                }
                
                # Cleanup
                index.delete([test_vector_id])
                test_results["cleanup"] = "success"
                
            except Exception as op_error:
                test_results["operations"] = {
                    "status": "failed",
                    "error": str(op_error)
                }
            
            total_time = time.time() - start_time
            overall_status = "success" if all(
                r.get("status", r) == "success" 
                for r in test_results.values()
            ) else "partial"
            
            return {
                "overall_status": overall_status,
                "test_duration": total_time,
                "index_name": self.index_name,
                "tests": test_results
            }
            
        except Exception as e:
            logger.error("Pinecone connection test failed", error=str(e))
            return {
                "overall_status": "failed",
                "error": str(e),
                "index_name": self.index_name
            }


# Global service instance
vectordb_service = VectorDBService()


async def get_vectordb_service() -> VectorDBService:
    """Dependency injection for vector database service."""
    return vectordb_service