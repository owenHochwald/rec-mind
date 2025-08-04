-- Migration: Create article_chunks table for chunked embeddings support
-- Author: Claude Code
-- Date: 2025-08-04

-- Create article_chunks table with proper foreign key constraints
CREATE TABLE IF NOT EXISTS article_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER,
    character_count INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_article_chunk UNIQUE(article_id, chunk_index)
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_article_chunks_article_id ON article_chunks(article_id);
CREATE INDEX IF NOT EXISTS idx_article_chunks_chunk_index ON article_chunks(chunk_index);
CREATE INDEX IF NOT EXISTS idx_article_chunks_created_at ON article_chunks(created_at);