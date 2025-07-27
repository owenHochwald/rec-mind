"""Configuration management for the ML service."""

import os
from typing import Optional
from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""
    
    # OpenAI Configuration
    openai_api_key: str = Field(..., description="OpenAI API key")
    openai_model: str = Field(default="text-embedding-3-small", description="OpenAI embedding model")
    embedding_dimensions: int = Field(default=1536, description="Embedding vector dimensions")
    max_tokens: int = Field(default=8192, description="Maximum tokens per request")
    
    # Pinecone Configuration
    pinecone_api_key: str = Field(..., description="Pinecone API key")
    pinecone_environment: str = Field(default="us-east-1", description="Pinecone environment")
    pinecone_host: Optional[str] = Field(default=None, description="Pinecone index host URL")
    pinecone_index_name: str = Field(default="news-articles", description="Pinecone index name")
    
    # Service Configuration
    ml_service_host: str = Field(default="0.0.0.0", description="Service host")
    ml_service_port: int = Field(default=8000, description="Service port")
    log_level: str = Field(default="info", description="Logging level")
    
    # Processing Configuration
    batch_size: int = Field(default=100, description="Batch processing size")
    request_timeout: int = Field(default=30, description="Request timeout in seconds")
    max_retries: int = Field(default=3, description="Maximum retry attempts")
    
    # CORS Configuration
    cors_origins: list[str] = Field(
        default=["http://localhost:8080", "http://localhost:3000"],
        description="Allowed CORS origins"
    )
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False


# Global settings instance
settings = Settings()


def get_settings() -> Settings:
    """Get application settings."""
    return settings