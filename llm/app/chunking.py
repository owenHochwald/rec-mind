"""LangChain-based text chunking service for article processing."""

import asyncio
import time
from typing import List, Dict, Any, Optional
from uuid import UUID

import structlog
from langchain_text_splitters import RecursiveCharacterTextSplitter

from .config import get_settings
from .models import ChunkRequest, ChunkResponse, ArticleProcessingRequest

logger = structlog.get_logger(__name__)
settings = get_settings()


class ChunkingService:
    """Service for splitting article content into chunks using LangChain."""
    
    def __init__(self):
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=500,
            chunk_overlap=50,
            separators=["\n\n", "\n", ". ", " "]
        )
        
    async def split_article(self, request: ArticleProcessingRequest, chunk_size: int = 500, chunk_overlap: int = 50) -> List[ChunkResponse]:
        """Split article content into chunks with metadata."""
        start_time = time.time()
        
        try:
            logger.info(
                "Starting article chunking",
                article_id=str(request.article_id),
                content_length=len(request.content),
                title_length=len(request.title),
                chunk_size=chunk_size,
                chunk_overlap=chunk_overlap
            )
            
            # Create text splitter with custom parameters if different from default
            if chunk_size != 500 or chunk_overlap != 50:
                text_splitter = RecursiveCharacterTextSplitter(
                    chunk_size=chunk_size,
                    chunk_overlap=chunk_overlap,
                    separators=["\n\n", "\n", ". ", " "]
                )
            else:
                text_splitter = self.text_splitter
            
            # Split ONLY the content into chunks (do NOT include title in chunks)
            # Title will be used separately for embedding context
            chunks = await asyncio.to_thread(
                text_splitter.split_text,
                request.content
            )
            
            # Create chunk responses with metadata
            chunk_responses = []
            for i, chunk_content in enumerate(chunks):
                # Calculate basic statistics
                token_count = len(chunk_content) // 4  # Rough approximation
                character_count = len(chunk_content)
                
                chunk_response = ChunkResponse(
                    article_id=request.article_id,
                    chunk_index=i,
                    content=chunk_content,
                    token_count=token_count,
                    character_count=character_count
                )
                chunk_responses.append(chunk_response)
            
            processing_time = time.time() - start_time
            
            logger.info(
                "Article chunking completed",
                article_id=str(request.article_id),
                total_chunks=len(chunk_responses),
                processing_time=processing_time,
                avg_chunk_size=sum(c.character_count for c in chunk_responses) / len(chunk_responses) if chunk_responses else 0
            )
            
            return chunk_responses
            
        except Exception as e:
            logger.error(
                "Article chunking failed",
                article_id=str(request.article_id),
                error=str(e)
            )
            raise
    
    async def validate_chunk_parameters(self) -> Dict[str, Any]:
        """Validate current chunking parameters."""
        return {
            "chunk_size": self.text_splitter._chunk_size,
            "chunk_overlap": self.text_splitter._chunk_overlap,
            "separators": self.text_splitter._separators,
            "length_function": str(self.text_splitter._length_function),
            "is_separator_regex": self.text_splitter._is_separator_regex
        }


# Global service instance
chunking_service = ChunkingService()


async def get_chunking_service() -> ChunkingService:
    """Dependency injection for chunking service."""
    return chunking_service