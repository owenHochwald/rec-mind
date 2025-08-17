package database

// Re-export models for backward compatibility
// TODO: Update all imports to use models package directly and remove this file

import "rec-mind/models"

// Core domain models
type Article = models.Article
type ArticleChunk = models.ArticleChunk

// Request/Response DTOs
type CreateArticleRequest = models.CreateArticleRequest
type UpdateArticleRequest = models.UpdateArticleRequest
type CreateArticleChunkRequest = models.CreateArticleChunkRequest
type UpdateArticleChunkRequest = models.UpdateArticleChunkRequest
type ArticleFilter = models.ArticleFilter
type ArticleChunkFilter = models.ArticleChunkFilter

// Search models
type QuerySearchJob = models.QuerySearchJob
type QuerySearchMessage = models.QuerySearchMessage
type QuerySearchResult = models.QuerySearchResult
type QuerySearchResponse = models.QuerySearchResponse
type QuerySearchError = models.QuerySearchError
type QueryRecommendationResult = models.QueryRecommendationResult

// Recommendation models
type RecommendationJob = models.RecommendationJob
type ChunkSearchMessage = models.ChunkSearchMessage
type ChunkSearchResult = models.ChunkSearchResult
type ChunkSearchResponse = models.ChunkSearchResponse
type ChunkSearchError = models.ChunkSearchError
type ChunkMatch = models.ChunkMatch
type ArticleRecommendation = models.ArticleRecommendation
type RecommendationResult = models.RecommendationResult