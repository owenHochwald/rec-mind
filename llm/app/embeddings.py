"""OpenAI embeddings service with async operations and retry logic."""

import asyncio
import time
from typing import List, Dict, Any
from uuid import UUID

import openai
import structlog
from openai import AsyncOpenAI
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

from .config import get_settings
from .models import EmbeddingRequest, EmbeddingResponse, BatchEmbeddingRequest, BatchEmbeddingResponse


logger = structlog.get_logger(__name__)
settings = get_settings()


class EmbeddingsService:
    """Service for generating embeddings using OpenAI API."""
    
    def __init__(self):
        self.client = AsyncOpenAI(api_key=settings.openai_api_key)
        self.model = settings.openai_model
        self.max_tokens = settings.max_tokens
        self.batch_size = settings.batch_size
        
    @retry(
        retry=retry_if_exception_type((openai.RateLimitError, openai.APITimeoutError, openai.APIConnectionError)),
        stop=stop_after_attempt(settings.max_retries),
        wait=wait_exponential(multiplier=1, min=4, max=10),
        reraise=True
    )
    async def _generate_embedding(self, text: str) -> tuple[List[float], int]:
        """Generate embedding for a single text with retry logic."""
        try:
            logger.info("Generating embedding", text_length=len(text), model=self.model)
            
            response = await self.client.embeddings.create(
                model=self.model,
                input=text,
                encoding_format="float"
            )
            
            embedding = response.data[0].embedding
            tokens_used = response.usage.total_tokens
            
            logger.info(
                "Embedding generated successfully",
                dimensions=len(embedding),
                tokens_used=tokens_used
            )
            
            return embedding, tokens_used
            
        except openai.RateLimitError as e:
            logger.warning("Rate limit exceeded, retrying", error=str(e))
            raise
        except openai.APITimeoutError as e:
            logger.warning("API timeout, retrying", error=str(e))
            raise
        except openai.APIConnectionError as e:
            logger.warning("API connection error, retrying", error=str(e))
            raise
        except Exception as e:
            logger.error("Unexpected error generating embedding", error=str(e))
            raise
    
    async def generate_embedding(self, request: EmbeddingRequest) -> EmbeddingResponse:
        """Generate embedding for a single article."""
        start_time = time.time()
        
        try:
            embedding, tokens_used = await self._generate_embedding(request.text)
            
            processing_time = time.time() - start_time
            logger.info(
                "Single embedding completed",
                article_id=str(request.article_id),
                processing_time=processing_time,
                tokens_used=tokens_used
            )
            
            return EmbeddingResponse(
                article_id=request.article_id,
                embeddings=embedding,
                dimensions=len(embedding),
                model=self.model,
                tokens_used=tokens_used
            )
            
        except Exception as e:
            logger.error(
                "Failed to generate embedding",
                article_id=str(request.article_id),
                error=str(e)
            )
            raise
    
    async def generate_batch_embeddings(self, request: BatchEmbeddingRequest) -> BatchEmbeddingResponse:
        """Generate embeddings for multiple articles in parallel."""
        start_time = time.time()
        total_tokens = 0
        results = []
        
        try:
            logger.info("Starting batch embedding generation", batch_size=len(request.items))
            
            # Process items in chunks to respect rate limits
            for i in range(0, len(request.items), self.batch_size):
                chunk = request.items[i:i + self.batch_size]
                
                # Create tasks for parallel processing within chunk
                tasks = [self.generate_embedding(item) for item in chunk]
                chunk_results = await asyncio.gather(*tasks, return_exceptions=True)
                
                for result in chunk_results:
                    if isinstance(result, Exception):
                        logger.error("Batch item failed", error=str(result))
                        raise result
                    else:
                        results.append(result)
                        total_tokens += result.tokens_used
                
                # Rate limiting between chunks
                if i + self.batch_size < len(request.items):
                    await asyncio.sleep(0.1)
            
            processing_time = time.time() - start_time
            logger.info(
                "Batch embedding completed",
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
            logger.error("Batch embedding failed", error=str(e))
            raise
    
    async def health_check(self) -> Dict[str, Any]:
        """Check OpenAI API connectivity."""
        try:
            start_time = time.time()
            
            # Simple health check with minimal API call
            await self.client.embeddings.create(
                model=self.model,
                input="health check",
                encoding_format="float"
            )
            
            response_time = time.time() - start_time
            
            return {
                "status": "healthy",
                "response_time": response_time,
                "model": self.model
            }
            
        except Exception as e:
            logger.error("OpenAI health check failed", error=str(e))
            return {
                "status": "unhealthy",
                "error": str(e),
                "model": self.model
            }


# Global service instance
embeddings_service = EmbeddingsService()


async def get_embeddings_service() -> EmbeddingsService:
    """Dependency injection for embeddings service."""
    return embeddings_service