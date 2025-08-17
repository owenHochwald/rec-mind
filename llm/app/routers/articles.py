"""Article processing endpoints."""

import time
import uuid
import structlog
from typing import List
from fastapi import APIRouter, Depends

from ..models import (
    ArticleProcessingRequest, ArticleProcessingResponse,
    EmbeddingRequest, BatchEmbeddingRequest, PineconeUploadRequest
)
from ..embeddings import get_embeddings_service, EmbeddingsService
from ..vectordb import get_vectordb_service, VectorDBService
from ..chunking import get_chunking_service, ChunkingService
from ..database import get_database_service, DatabaseService
from ..utils import endpoint_error_handler

router = APIRouter()
logger = structlog.get_logger(__name__)


@router.post("/articles/process", response_model=ArticleProcessingResponse)
@endpoint_error_handler("article processing with chunking")
async def process_article(
    request: ArticleProcessingRequest,
    chunking_svc: ChunkingService = Depends(get_chunking_service),
    database_svc: DatabaseService = Depends(get_database_service),
    embeddings_svc: EmbeddingsService = Depends(get_embeddings_service),
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Process article: chunk content, store in DB, generate embeddings, upload to Pinecone."""
    start_time = time.time()
    
    logger.info("Processing article via HTTP", article_id=str(request.article_id))
    
    # Step 1: Chunk the article
    chunks = await chunking_svc.split_article(request)
    
    # Step 2: Store chunks in PostgreSQL (if available)
    chunk_ids = []
    try:
        chunk_ids = await database_svc.store_chunks(chunks)
        logger.info("Chunks stored in PostgreSQL", count=len(chunk_ids))
    except Exception as e:
        logger.warning("Failed to store chunks in PostgreSQL - continuing without DB storage", error=str(e))
        # Generate fake chunk IDs for Pinecone upload
        chunk_ids = [uuid.uuid4() for _ in chunks]
    
    # Step 3: Generate embeddings for chunks
    embedding_requests = []
    for chunk in chunks:
        embedding_text = f"{request.title}\n\n{chunk.content}"
        embedding_requests.append(EmbeddingRequest(
            article_id=chunk.article_id,
            text=embedding_text
        ))
    
    batch_request = BatchEmbeddingRequest(items=embedding_requests)
    embedding_response = await embeddings_svc.generate_batch_embeddings(batch_request)
    
    # Step 4: Upload embeddings to Pinecone
    upload_count = 0
    for chunk, embedding_result, chunk_id in zip(chunks, embedding_response.results, chunk_ids):
        upload_request = PineconeUploadRequest(
            article_id=chunk.article_id,
            embeddings=embedding_result.embeddings,
            metadata={
                "chunk_id": str(chunk_id),
                "chunk_index": chunk.chunk_index,
                "article_title": request.title,
                "category": request.category,
                "token_count": chunk.token_count,
                "character_count": chunk.character_count,
                "created_at": request.created_at.isoformat()
            }
        )
        
        await vectordb_svc.upload_embedding(upload_request, vector_id=str(chunk_id))
        upload_count += 1
    
    processing_time = time.time() - start_time
    
    return ArticleProcessingResponse(
        article_id=request.article_id,
        chunks_created=len(chunks),
        embeddings_generated=len(embedding_response.results),
        processing_time=processing_time,
        status="completed"
    )


@router.get("/chunks/{article_id}")
@endpoint_error_handler("chunk retrieval")
async def get_article_chunks(
    article_id: str,
    database_svc: DatabaseService = Depends(get_database_service)
):
    """Get all chunks for a specific article."""
    try:
        from uuid import UUID
        article_uuid = UUID(article_id)
        chunks = await database_svc.get_chunks_by_article(article_uuid)
        return {
            "article_id": article_id,
            "chunks": chunks,
            "total_chunks": len(chunks)
        }
    except ValueError:
        return {"error": "Invalid article ID format"}
    except Exception as e:
        logger.error("Failed to retrieve chunks", article_id=article_id, error=str(e))
        return {"error": f"Failed to retrieve chunks: {str(e)}"}