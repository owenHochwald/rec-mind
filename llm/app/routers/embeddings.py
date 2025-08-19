"""Embeddings API endpoints."""

import time
from typing import List, Optional
import structlog
from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel, Field
import uuid

from ..embeddings import get_embeddings_service, EmbeddingsService
from ..vectordb import get_vectordb_service, VectorDBService
from ..models import ErrorResponse

logger = structlog.get_logger(__name__)

router = APIRouter(prefix="/embeddings")


class EmbeddingItem(BaseModel):
    """Single embedding request item."""
    article_id: str = Field(..., description="Article UUID")
    text: str = Field(..., description="Text content to embed")


class BatchEmbeddingRequest(BaseModel):
    """Batch embedding request."""
    items: List[EmbeddingItem] = Field(..., description="List of articles to process")


class EmbeddingResult(BaseModel):
    """Single embedding result."""
    article_id: str
    embeddings: List[float]
    dimensions: int
    model: str
    tokens_used: int


class BatchEmbeddingResponse(BaseModel):
    """Batch embedding response."""
    results: List[EmbeddingResult]
    total_tokens: int
    processing_time: float


class UploadResult(BaseModel):
    """Upload result for a single article."""
    article_id: str
    status: str
    upload_time: float
    upserted_count: int


class ProcessingSummary(BaseModel):
    """Processing summary."""
    total_processed: int
    total_tokens: int
    processing_time: float


class BatchAndUploadResponse(BaseModel):
    """Complete batch and upload response."""
    embeddings: BatchEmbeddingResponse
    uploads: List[UploadResult]
    summary: ProcessingSummary


@router.post("/batch-and-upload", response_model=BatchAndUploadResponse)
async def batch_and_upload(
    request: BatchEmbeddingRequest,
    embeddings_service: EmbeddingsService = Depends(get_embeddings_service),
    vectordb_service: VectorDBService = Depends(get_vectordb_service)
):
    """Generate embeddings for multiple articles and upload to Pinecone."""
    start_time = time.time()
    
    if not request.items:
        raise HTTPException(status_code=400, detail="No items provided for processing")
    
    logger.info("Processing batch embedding request", item_count=len(request.items))
    
    try:
        embedding_results = []
        upload_results = []
        total_tokens = 0
        total_processed = 0
        
        for item in request.items:
            try:
                # Validate UUID
                article_uuid = uuid.UUID(item.article_id)
                
                # Generate embedding
                embedding_result = await embeddings_service.generate_embedding(item.text)
                
                embedding_response = EmbeddingResult(
                    article_id=item.article_id,
                    embeddings=embedding_result.embedding,
                    dimensions=len(embedding_result.embedding),
                    model=embedding_result.model,
                    tokens_used=embedding_result.tokens_used
                )
                embedding_results.append(embedding_response)
                total_tokens += embedding_result.tokens_used
                
                # Upload to Pinecone
                upload_start = time.time()
                await vectordb_service.upsert_article_embedding(
                    article_id=str(article_uuid),
                    embedding=embedding_result.embedding,
                    metadata={
                        "article_id": str(article_uuid),
                        "text_preview": item.text[:200] + "..." if len(item.text) > 200 else item.text,
                        "embedding_model": embedding_result.model,
                        "tokens_used": embedding_result.tokens_used
                    }
                )
                upload_time = time.time() - upload_start
                
                upload_results.append(UploadResult(
                    article_id=item.article_id,
                    status="success",
                    upload_time=upload_time,
                    upserted_count=1
                ))
                
                total_processed += 1
                logger.info("Article processed successfully", article_id=item.article_id)
                
            except ValueError as e:
                logger.error("Invalid article ID", article_id=item.article_id, error=str(e))
                upload_results.append(UploadResult(
                    article_id=item.article_id,
                    status="error",
                    upload_time=0.0,
                    upserted_count=0
                ))
            except Exception as e:
                logger.error("Failed to process article", article_id=item.article_id, error=str(e))
                upload_results.append(UploadResult(
                    article_id=item.article_id,
                    status="error", 
                    upload_time=0.0,
                    upserted_count=0
                ))
        
        processing_time = time.time() - start_time
        
        # Build response
        embeddings_response = BatchEmbeddingResponse(
            results=embedding_results,
            total_tokens=total_tokens,
            processing_time=processing_time
        )
        
        summary = ProcessingSummary(
            total_processed=total_processed,
            total_tokens=total_tokens,
            processing_time=processing_time
        )
        
        response = BatchAndUploadResponse(
            embeddings=embeddings_response,
            uploads=upload_results,
            summary=summary
        )
        
        logger.info(
            "Batch processing completed",
            total_items=len(request.items),
            successful=total_processed,
            processing_time=processing_time
        )
        
        return response
        
    except Exception as e:
        logger.error("Batch processing failed", error=str(e))
        raise HTTPException(status_code=500, detail=f"Batch processing failed: {str(e)}")


@router.post("/generate", response_model=EmbeddingResult)
async def generate_embedding(
    article_id: str,
    text: str,
    embeddings_service: EmbeddingsService = Depends(get_embeddings_service)
):
    """Generate embedding for a single text."""
    try:
        # Validate UUID
        uuid.UUID(article_id)
        
        # Generate embedding
        result = await embeddings_service.generate_embedding(text)
        
        return EmbeddingResult(
            article_id=article_id,
            embeddings=result.embedding,
            dimensions=len(result.embedding),
            model=result.model,
            tokens_used=result.tokens_used
        )
        
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid article ID format")
    except Exception as e:
        logger.error("Failed to generate embedding", article_id=article_id, error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to generate embedding: {str(e)}")


@router.post("/upload")
async def upload_embedding(
    article_id: str,
    embedding: List[float],
    metadata: Optional[dict] = None,
    vectordb_service: VectorDBService = Depends(get_vectordb_service)
):
    """Upload embedding to Pinecone."""
    try:
        # Validate UUID
        uuid.UUID(article_id)
        
        # Upload to Pinecone
        await vectordb_service.upsert_article_embedding(
            article_id=article_id,
            embedding=embedding,
            metadata=metadata or {}
        )
        
        return {"status": "success", "article_id": article_id}
        
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid article ID format")
    except Exception as e:
        logger.error("Failed to upload embedding", article_id=article_id, error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to upload embedding: {str(e)}")