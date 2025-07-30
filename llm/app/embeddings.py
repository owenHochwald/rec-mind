"""LangChain-based embeddings service with batch processing."""

import asyncio
import time
from typing import List, Dict, Any
from uuid import UUID

import structlog
from langchain_openai import OpenAIEmbeddings
from langchain_core.documents import Document

from .config import get_settings
from .models import EmbeddingRequest, EmbeddingResponse, BatchEmbeddingRequest, BatchEmbeddingResponse


logger = structlog.get_logger(__name__)
settings = get_settings()


class EmbeddingsService:
    """LangChain-based embeddings service for generating embeddings."""
    
    def __init__(self):
        self.embeddings = OpenAIEmbeddings(
            api_key=settings.openai_api_key,
            model=settings.openai_model
        )
        self.batch_size = settings.batch_size
        
    async def generate_embedding(self, request: EmbeddingRequest) -> EmbeddingResponse:
        """Generate embedding for a single article using LangChain."""
        start_time = time.time()
        
        try:
            logger.info(
                "Generating embedding with LangChain",
                article_id=str(request.article_id),
                text_length=len(request.text),
                model=settings.openai_model
            )
            
            # Use LangChain's embed_query method for single text
            embedding_vector = await asyncio.to_thread(
                self.embeddings.embed_query, 
                request.text
            )
            
            processing_time = time.time() - start_time
            
            # Estimate tokens (rough approximation: 1 token ≈ 4 characters)
            estimated_tokens = len(request.text) // 4
            
            logger.info(
                "Single embedding completed with LangChain",
                article_id=str(request.article_id),
                processing_time=processing_time,
                dimensions=len(embedding_vector),
                estimated_tokens=estimated_tokens
            )
            
            return EmbeddingResponse(
                article_id=request.article_id,
                embeddings=embedding_vector,
                dimensions=len(embedding_vector),
                model=settings.openai_model,
                tokens_used=estimated_tokens
            )
            
        except Exception as e:
            logger.error(
                "Failed to generate embedding with LangChain",
                article_id=str(request.article_id),
                error=str(e)
            )
            raise
    
    async def generate_batch_embeddings(self, request: BatchEmbeddingRequest) -> BatchEmbeddingResponse:
        """Generate embeddings for multiple articles using LangChain batch processing."""
        start_time = time.time()
        total_tokens = 0
        results = []
        
        try:
            logger.info(
                "Starting batch embedding generation with LangChain",
                batch_size=len(request.items)
            )
            
            # Process items in chunks to respect rate limits
            for i in range(0, len(request.items), self.batch_size):
                chunk = request.items[i:i + self.batch_size]
                logger.info(f"Processing chunk {i//self.batch_size + 1}/{(len(request.items) + self.batch_size - 1)//self.batch_size}")
                
                # Extract texts for batch processing
                texts = [item.text for item in chunk]
                
                try:
                    # Use LangChain's embed_documents method for batch processing
                    chunk_embeddings = await asyncio.to_thread(
                        self.embeddings.embed_documents,
                        texts
                    )
                    
                    # Create response objects for each embedding
                    for item, embedding_vector in zip(chunk, chunk_embeddings):
                        estimated_tokens = len(item.text) // 4
                        total_tokens += estimated_tokens
                        
                        result = EmbeddingResponse(
                            article_id=item.article_id,
                            embeddings=embedding_vector,
                            dimensions=len(embedding_vector),
                            model=settings.openai_model,
                            tokens_used=estimated_tokens
                        )
                        results.append(result)
                    
                    logger.info(
                        "Chunk processed successfully",
                        chunk_size=len(chunk),
                        chunk_tokens=sum(len(item.text) // 4 for item in chunk)
                    )
                    
                except Exception as chunk_error:
                    logger.error(
                        "Chunk processing failed",
                        chunk_index=i//self.batch_size,
                        error=str(chunk_error)
                    )
                    raise chunk_error
                
                # Rate limiting between chunks
                if i + self.batch_size < len(request.items):
                    await asyncio.sleep(0.1)
            
            processing_time = time.time() - start_time
            logger.info(
                "Batch embedding completed with LangChain",
                total_items=len(results),
                total_tokens=total_tokens,
                processing_time=processing_time
            )
            
            return BatchEmbeddingResponse(
                results=results,
                total_tokens=total_tokens,
                processing_time=processing_time
            )
            
        except Exception as e:
            logger.error("Batch embedding failed with LangChain", error=str(e))
            raise
    
    def create_document(self, article_id: UUID, text: str, metadata: Dict[str, Any] = None) -> Document:
        """Create a LangChain Document object for article processing."""
        doc_metadata = {
            "article_id": str(article_id),
            "source": f"article_{article_id}",
            **(metadata or {})
        }
        
        return Document(
            page_content=text,
            metadata=doc_metadata
        )
    
    def generate_article_id(self, article_id: UUID, chunk_num: int = 0) -> str:
        """Generate standardized ID for vector storage following LangChain pattern."""
        return f"article_{article_id}#chunk_{chunk_num}"
    
    async def embed_documents_for_vectorstore(self, articles: List[Dict[str, Any]]) -> List[Document]:
        """Convert articles to LangChain Documents ready for vector store upload."""
        documents = []
        
        for article in articles:
            doc = self.create_document(
                article_id=article["article_id"],
                text=article["text"],
                metadata=article.get("metadata", {})
            )
            documents.append(doc)
        
        logger.info(
            "Prepared documents for vector store",
            document_count=len(documents)
        )
        
        return documents
    
    async def health_check(self) -> Dict[str, Any]:
        """Check LangChain OpenAI embeddings connectivity."""
        try:
            start_time = time.time()
            
            # Simple health check with minimal API call
            test_embedding = await asyncio.to_thread(
                self.embeddings.embed_query,
                "health check"
            )
            
            response_time = time.time() - start_time
            
            return {
                "status": "healthy",
                "response_time": response_time,
                "model": settings.openai_model,
                "embedding_dimensions": len(test_embedding),
                "provider": "langchain_openai"
            }
            
        except Exception as e:
            logger.error("LangChain OpenAI health check failed", error=str(e))
            return {
                "status": "unhealthy",
                "error": str(e),
                "model": settings.openai_model,
                "provider": "langchain_openai"
            }


# Global service instance
embeddings_service = EmbeddingsService()


async def get_embeddings_service() -> EmbeddingsService:
    """Dependency injection for embeddings service."""
    return embeddings_service