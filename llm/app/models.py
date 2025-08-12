"""Pydantic models for request/response validation."""

from typing import List, Optional, Dict, Any
from uuid import UUID
from datetime import datetime
from pydantic import BaseModel, Field, validator


class EmbeddingRequest(BaseModel):
    """Request model for generating embeddings."""
    article_id: UUID = Field(..., description="Unique article identifier")
    text: str = Field(..., min_length=1, max_length=50000, description="Text content to embed")
    
    @validator('text')
    def validate_text(cls, v):
        if not v.strip():
            raise ValueError('Text cannot be empty or whitespace only')
        return v.strip()


class BatchEmbeddingRequest(BaseModel):
    """Request model for batch embedding generation."""
    items: List[EmbeddingRequest] = Field(..., min_items=1, max_items=100, description="Batch of texts to embed")


class EmbeddingResponse(BaseModel):
    """Response model for embedding generation."""
    article_id: UUID = Field(..., description="Article identifier")
    embeddings: List[float] = Field(..., description="Embedding vector")
    dimensions: int = Field(..., description="Vector dimensions")
    model: str = Field(..., description="Model used for embedding")
    tokens_used: int = Field(..., description="Number of tokens processed")


class BatchEmbeddingResponse(BaseModel):
    """Response model for batch embedding generation."""
    results: List[EmbeddingResponse] = Field(..., description="Batch embedding results")
    total_tokens: int = Field(..., description="Total tokens used")
    processing_time: float = Field(..., description="Processing time in seconds")


class PineconeUploadRequest(BaseModel):
    """Request model for uploading embeddings to Pinecone."""
    article_id: UUID = Field(..., description="Article identifier")
    embeddings: List[float] = Field(..., description="Embedding vector")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")


class PineconeSearchRequest(BaseModel):
    """Request model for Pinecone similarity search."""
    article_id: UUID = Field(..., description="Source article identifier")
    top_k: int = Field(default=5, ge=1, le=100, description="Number of similar articles to return")
    score_threshold: float = Field(default=0.7, ge=0.0, le=1.0, description="Minimum similarity score")
    include_metadata: bool = Field(default=True, description="Include metadata in results")


class SearchResult(BaseModel):
    """Individual search result model."""
    id: str = Field(..., description="Article ID")
    score: float = Field(..., description="Similarity score")
    metadata: Optional[Dict[str, Any]] = Field(default=None, description="Article metadata")


class PineconeSearchResponse(BaseModel):
    """Response model for Pinecone search results."""
    query_article_id: UUID = Field(..., description="Source article ID")
    results: List[SearchResult] = Field(..., description="Similar articles")
    total_found: int = Field(..., description="Total number of results")
    processing_time: float = Field(..., description="Search time in seconds")


class HealthStatus(BaseModel):
    """Health check status model."""
    status: str = Field(..., description="Service status")
    timestamp: datetime = Field(default_factory=datetime.utcnow, description="Health check timestamp")
    version: str = Field(default="1.0.0", description="Service version")


class DependencyHealth(BaseModel):
    """External dependency health status."""
    name: str = Field(..., description="Dependency name")
    status: str = Field(..., description="Dependency status")
    response_time: Optional[float] = Field(default=None, description="Response time in seconds")
    error: Optional[str] = Field(default=None, description="Error message if unhealthy")


class DetailedHealthResponse(BaseModel):
    """Detailed health check response."""
    service: HealthStatus = Field(..., description="Service health")
    dependencies: List[DependencyHealth] = Field(..., description="Dependency health")
    uptime: float = Field(..., description="Service uptime in seconds")


class ErrorResponse(BaseModel):
    """Standard error response model."""
    error: str = Field(..., description="Error message")
    error_code: str = Field(..., description="Error code")
    timestamp: datetime = Field(default_factory=datetime.utcnow, description="Error timestamp")
    request_id: Optional[str] = Field(default=None, description="Request identifier")


class ArticleProcessingRequest(BaseModel):
    """Request model for article processing with chunking."""
    article_id: UUID = Field(..., description="Unique article identifier")
    title: str = Field(..., min_length=1, max_length=500, description="Article title")
    content: str = Field(..., min_length=1, description="Article content")
    category: str = Field(..., min_length=1, max_length=100, description="Article category")
    created_at: datetime = Field(..., description="Article creation timestamp")


class ChunkRequest(BaseModel):
    """Request model for text chunking."""
    text: str = Field(..., min_length=1, description="Text to be chunked")
    chunk_size: Optional[int] = Field(default=500, ge=100, le=2000, description="Maximum chunk size")
    chunk_overlap: Optional[int] = Field(default=50, ge=0, le=200, description="Overlap between chunks")


class ChunkResponse(BaseModel):
    """Response model for individual text chunks."""
    article_id: UUID = Field(..., description="Source article identifier")
    chunk_index: int = Field(..., ge=0, description="Index of chunk within article")
    content: str = Field(..., description="Chunk content")
    token_count: int = Field(..., ge=0, description="Estimated token count")
    character_count: int = Field(..., ge=0, description="Character count")
    pinecone_id: Optional[str] = Field(default=None, description="Pinecone vector ID reference")


class ArticleProcessingResponse(BaseModel):
    """Response model for complete article processing."""
    article_id: UUID = Field(..., description="Article identifier")
    chunks_created: int = Field(..., ge=0, description="Number of chunks created")
    embeddings_generated: int = Field(..., ge=0, description="Number of embeddings generated")
    processing_time: float = Field(..., ge=0, description="Total processing time in seconds")
    status: str = Field(..., description="Processing status")


class ChunkSearchMessage(BaseModel):
    """Message model for chunk similarity search requests."""
    search_id: UUID = Field(..., description="Unique search identifier")
    chunk_id: UUID = Field(..., description="Source chunk ID for similarity search")
    source_article_id: UUID = Field(..., description="Source article ID to exclude from results")
    top_k: int = Field(default=5, ge=1, le=20, description="Number of similar chunks to return")
    score_threshold: float = Field(default=0.7, ge=0.0, le=1.0, description="Minimum similarity score")
    service_instance_id: Optional[str] = Field(default=None, description="Processing service instance ID")


class ChunkSearchResult(BaseModel):
    """Individual chunk search result."""
    chunk_id: str = Field(..., description="Chunk ID")
    similarity_score: float = Field(..., description="Similarity score")
    article_id: UUID = Field(..., description="Article ID containing the chunk")
    chunk_index: int = Field(..., description="Chunk index within article")
    article_title: str = Field(..., description="Article title")
    category: str = Field(..., description="Article category")
    content_preview: str = Field(..., description="Chunk content preview")


class ChunkSearchResponse(BaseModel):
    """Response model for chunk similarity search results."""
    search_id: UUID = Field(..., description="Search identifier")
    source_chunk_id: UUID = Field(..., description="Source chunk ID")
    results: List[ChunkSearchResult] = Field(..., description="Similar chunks")
    total_found: int = Field(..., description="Total results found")
    processing_time: float = Field(..., description="Search processing time")
    service_instance_id: str = Field(..., description="Processing service instance")


class ChunkSearchError(BaseModel):
    """Error model for failed chunk searches."""
    search_id: UUID = Field(..., description="Search identifier")
    error_message: str = Field(..., description="Error description")
    error_code: str = Field(..., description="Error code")
    service_instance_id: str = Field(..., description="Processing service instance")