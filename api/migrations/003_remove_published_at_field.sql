-- Migration: Remove published_at field from articles table
-- Author: Claude Code
-- Date: 2025-08-04

-- Drop the published_at index first
DROP INDEX IF EXISTS idx_articles_published_at;

-- Remove the published_at column from articles table
ALTER TABLE articles DROP COLUMN IF EXISTS published_at;