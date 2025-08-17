"""Text processing endpoints."""

from typing import List
from fastapi import APIRouter, Depends

from ..models import ChunkRequest, ChunkResponse
from ..chunking import get_chunking_service, ChunkingService
from ..utils import endpoint_error_handler

router = APIRouter()


@router.post("/text/chunk", response_model=List[ChunkResponse])
@endpoint_error_handler("text chunking")
async def chunk_text(
    request: ChunkRequest,
    chunking_svc: ChunkingService = Depends(get_chunking_service)
):
    """Split text into chunks using LangChain RecursiveCharacterTextSplitter."""
    from uuid import uuid4
    from ..models import ArticleProcessingRequest
    from datetime import datetime
    
    # Create temporary processing request for chunking
    temp_request = ArticleProcessingRequest(
        article_id=uuid4(),
        title="Text Chunking Request",
        content=request.text,
        category="test",
        created_at=datetime.utcnow()
    )
    
    # Use custom chunk parameters from request
    chunks = await chunking_svc.split_article(
        temp_request, 
        chunk_size=request.chunk_size or 500,
        chunk_overlap=request.chunk_overlap or 50
    )
    return chunks


@router.get("/chunks/parameters")
async def get_chunk_parameters(
    chunking_svc: ChunkingService = Depends(get_chunking_service)
):
    """Get current chunking parameters."""
    return await chunking_svc.validate_chunk_parameters()