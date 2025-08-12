package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"rec-mind/config"
	"rec-mind/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDBForChunks(t *testing.T) (*database.DB, ArticleRepository, ArticleChunkRepository) {
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5431,
		Name:           "postgres",
		User:           "postgres",
		Password:       "secret",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    15 * time.Minute,
	}

	db, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available (%v)", err)
		return nil, nil, nil
	}

	ctx := context.Background()
	_, err = db.Pool.Exec(ctx, `
		DROP TABLE IF EXISTS article_chunks CASCADE;
		DROP TABLE IF EXISTS articles CASCADE;
		
		CREATE TABLE articles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(500) NOT NULL,
			content TEXT NOT NULL,
			url VARCHAR(1000) UNIQUE NOT NULL,
			category VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
		
		CREATE TABLE article_chunks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
			chunk_index INTEGER NOT NULL,
			content TEXT NOT NULL,
			token_count INTEGER,
			character_count INTEGER,
			created_at TIMESTAMP DEFAULT NOW(),
			CONSTRAINT unique_article_chunk UNIQUE(article_id, chunk_index)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	articleRepo := NewArticleRepository(db.Pool)
	chunkRepo := NewArticleChunkRepository(db.Pool)
	return db, articleRepo, chunkRepo
}

func cleanupTestDBForChunks(t *testing.T, db *database.DB) {
	if db == nil {
		return
	}
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE article_chunks, articles CASCADE")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test data: %v", err)
	}
	db.Close()
}

func createTestArticle(t *testing.T, repo ArticleRepository) *database.Article {
	req := &database.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "This is test content for the article.",
		URL:         "https://example.com/test-article-" + uuid.New().String(),
		Category:    "Technology",
	}

	article, err := repo.Create(context.Background(), req)
	require.NoError(t, err)
	return article
}

func TestArticleChunkRepository_Create(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)

	req := &database.CreateArticleChunkRequest{
		ArticleID:      article.ID,
		ChunkIndex:     0,
		Content:        "This is the first chunk of the article.",
		TokenCount:     intPtr(10),
		CharacterCount: intPtr(42),
	}

	chunk, err := chunkRepo.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, chunk.ID)
	assert.Equal(t, article.ID, chunk.ArticleID)
	assert.Equal(t, 0, chunk.ChunkIndex)
	assert.Equal(t, req.Content, chunk.Content)
	assert.Equal(t, req.TokenCount, chunk.TokenCount)
	assert.Equal(t, req.CharacterCount, chunk.CharacterCount)
	assert.NotZero(t, chunk.CreatedAt)
}

func TestArticleChunkRepository_GetByID(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	chunk := createTestChunk(t, chunkRepo, article.ID, 0)

	foundChunk, err := chunkRepo.GetByID(context.Background(), chunk.ID)
	require.NoError(t, err)
	assert.Equal(t, chunk.ID, foundChunk.ID)
	assert.Equal(t, chunk.ArticleID, foundChunk.ArticleID)
	assert.Equal(t, chunk.ChunkIndex, foundChunk.ChunkIndex)
	assert.Equal(t, chunk.Content, foundChunk.Content)
}

func TestArticleChunkRepository_GetByArticleID(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	
	chunk1 := createTestChunk(t, chunkRepo, article.ID, 0)
	chunk2 := createTestChunk(t, chunkRepo, article.ID, 1)
	chunk3 := createTestChunk(t, chunkRepo, article.ID, 2)

	chunks, err := chunkRepo.GetByArticleID(context.Background(), article.ID)
	require.NoError(t, err)
	assert.Len(t, chunks, 3)
	
	assert.Equal(t, chunk1.ID, chunks[0].ID)
	assert.Equal(t, chunk2.ID, chunks[1].ID)
	assert.Equal(t, chunk3.ID, chunks[2].ID)
	
	assert.Equal(t, 0, chunks[0].ChunkIndex)
	assert.Equal(t, 1, chunks[1].ChunkIndex)
	assert.Equal(t, 2, chunks[2].ChunkIndex)
}

func TestArticleChunkRepository_GetByArticleIDAndIndex(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	chunk := createTestChunk(t, chunkRepo, article.ID, 5)

	foundChunk, err := chunkRepo.GetByArticleIDAndIndex(context.Background(), article.ID, 5)
	require.NoError(t, err)
	assert.Equal(t, chunk.ID, foundChunk.ID)
	assert.Equal(t, 5, foundChunk.ChunkIndex)
}

func TestArticleChunkRepository_Update(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	chunk := createTestChunk(t, chunkRepo, article.ID, 0)

	newContent := "Updated chunk content"
	newTokenCount := 15
	req := &database.UpdateArticleChunkRequest{
		Content:    &newContent,
		TokenCount: &newTokenCount,
	}

	updatedChunk, err := chunkRepo.Update(context.Background(), chunk.ID, req)
	require.NoError(t, err)
	assert.Equal(t, chunk.ID, updatedChunk.ID)
	assert.Equal(t, newContent, updatedChunk.Content)
	assert.Equal(t, &newTokenCount, updatedChunk.TokenCount)
}

func TestArticleChunkRepository_Delete(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	chunk := createTestChunk(t, chunkRepo, article.ID, 0)

	err := chunkRepo.Delete(context.Background(), chunk.ID)
	require.NoError(t, err)

	_, err = chunkRepo.GetByID(context.Background(), chunk.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestArticleChunkRepository_DeleteByArticleID(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	createTestChunk(t, chunkRepo, article.ID, 0)
	createTestChunk(t, chunkRepo, article.ID, 1)

	err := chunkRepo.DeleteByArticleID(context.Background(), article.ID)
	require.NoError(t, err)

	chunks, err := chunkRepo.GetByArticleID(context.Background(), article.ID)
	require.NoError(t, err)
	assert.Len(t, chunks, 0)
}

func TestArticleChunkRepository_CreateBatch(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)

	chunks := []*database.CreateArticleChunkRequest{
		{
			ArticleID:      article.ID,
			ChunkIndex:     0,
			Content:        "First chunk",
			TokenCount:     intPtr(5),
			CharacterCount: intPtr(11),
		},
		{
			ArticleID:      article.ID,
			ChunkIndex:     1,
			Content:        "Second chunk",
			TokenCount:     intPtr(6),
			CharacterCount: intPtr(12),
		},
	}

	results, err := chunkRepo.CreateBatch(context.Background(), chunks)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	for i, result := range results {
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, article.ID, result.ArticleID)
		assert.Equal(t, i, result.ChunkIndex)
		assert.Equal(t, chunks[i].Content, result.Content)
	}
}

func TestArticleChunkRepository_List(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article1 := createTestArticle(t, articleRepo)
	article2 := createTestArticle(t, articleRepo)

	createTestChunk(t, chunkRepo, article1.ID, 0)
	createTestChunk(t, chunkRepo, article1.ID, 1)
	createTestChunk(t, chunkRepo, article2.ID, 0)

	filter := &database.ArticleChunkFilter{
		ArticleID: &article1.ID,
		Limit:     10,
		Offset:    0,
	}

	chunks, err := chunkRepo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, chunks, 2)

	for _, chunk := range chunks {
		assert.Equal(t, article1.ID, chunk.ArticleID)
	}
}

func TestArticleChunkRepository_Count(t *testing.T) {
	db, articleRepo, chunkRepo := setupTestDBForChunks(t)
	if db == nil {
		return
	}
	defer cleanupTestDBForChunks(t, db)

	article := createTestArticle(t, articleRepo)
	createTestChunk(t, chunkRepo, article.ID, 0)
	createTestChunk(t, chunkRepo, article.ID, 1)

	filter := &database.ArticleChunkFilter{
		ArticleID: &article.ID,
	}

	count, err := chunkRepo.Count(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func createTestChunk(t *testing.T, repo ArticleChunkRepository, articleID uuid.UUID, index int) *database.ArticleChunk {
	req := &database.CreateArticleChunkRequest{
		ArticleID:      articleID,
		ChunkIndex:     index,
		Content:        "Test chunk content",
		TokenCount:     intPtr(10),
		CharacterCount: intPtr(20),
	}

	chunk, err := repo.Create(context.Background(), req)
	require.NoError(t, err)
	return chunk
}

func intPtr(i int) *int {
	return &i
}