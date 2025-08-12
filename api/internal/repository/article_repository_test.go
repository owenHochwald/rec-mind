package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rec-mind/config"
	"rec-mind/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, ArticleRepository) {
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
		return nil, nil
	}

	ctx := context.Background()
	_, err = db.Pool.Exec(ctx, `
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
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	repo := NewArticleRepository(db.Pool)
	return db, repo
}

func cleanupTestDB(t *testing.T, db *database.DB) {
	if db == nil {
		return
	}
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE articles CASCADE")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test data: %v", err)
	}
	db.Close()
}

func TestArticleRepository_Create(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "This is test content for the article.",
		URL:         "https://example.com/test-article",
		Category:    "Technology",
	}

	article, err := repo.Create(context.Background(), req)
	
	require.NoError(t, err)
	require.NotNil(t, article)
	assert.NotEqual(t, uuid.Nil, article.ID)
	assert.Equal(t, req.Title, article.Title)
	assert.Equal(t, req.Content, article.Content)
	assert.Equal(t, req.URL, article.URL)
	assert.Equal(t, req.Category, article.Category)
	assert.False(t, article.CreatedAt.IsZero())
	assert.False(t, article.UpdatedAt.IsZero())
}

func TestArticleRepository_Create_DuplicateURL(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "Content",
		URL:         "https://example.com/duplicate",
		Category:    "Technology",
	}

	_, err := repo.Create(context.Background(), req)
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestArticleRepository_GetByID(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "Content",
		URL:         "https://example.com/test",
		Category:    "Technology",
	}

	created, err := repo.Create(context.Background(), req)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Title, found.Title)
}

func TestArticleRepository_GetByID_NotFound(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	id := uuid.New()
	_, err := repo.GetByID(context.Background(), id)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "article not found")
}

func TestArticleRepository_GetByURL(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "Content",
		URL:         "https://example.com/unique-url",
		Category:    "Technology",
	}

	created, err := repo.Create(context.Background(), req)
	require.NoError(t, err)

	found, err := repo.GetByURL(context.Background(), req.URL)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestArticleRepository_List(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	articles := []*database.CreateArticleRequest{
		{
			Title:       "Tech Article 1",
			Content:     "Tech content 1",
			URL:         "https://example.com/tech1",
			Category:    "Technology",
			},
		{
			Title:       "Sports Article 1",
			Content:     "Sports content 1",
			URL:         "https://example.com/sports1",
			Category:    "Sports",
			},
	}

	for _, req := range articles {
		_, err := repo.Create(context.Background(), req)
		require.NoError(t, err)
	}

	filter := &database.ArticleFilter{
		Limit:  10,
		Offset: 0,
	}

	results, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestArticleRepository_List_WithCategoryFilter(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	techReq := &database.CreateArticleRequest{
		Title:       "Tech Article",
		Content:     "Tech content",
		URL:         "https://example.com/tech-filter",
		Category:    "Technology",
	}

	sportsReq := &database.CreateArticleRequest{
		Title:       "Sports Article",
		Content:     "Sports content",
		URL:         "https://example.com/sports-filter",
		Category:    "Sports",
	}

	_, err := repo.Create(context.Background(), techReq)
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), sportsReq)
	require.NoError(t, err)

	category := "Technology"
	filter := &database.ArticleFilter{
		Category: &category,
		Limit:    10,
		Offset:   0,
	}

	results, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Technology", results[0].Category)
}

func TestArticleRepository_Update(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "Original Title",
		Content:     "Original content",
		URL:         "https://example.com/update-test",
		Category:    "Technology",
	}

	created, err := repo.Create(context.Background(), req)
	require.NoError(t, err)

	newTitle := "Updated Title"
	updateReq := &database.UpdateArticleRequest{
		Title: &newTitle,
	}

	updated, err := repo.Update(context.Background(), created.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
	assert.Equal(t, created.Content, updated.Content)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt))
}

func TestArticleRepository_Delete(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "To Delete",
		Content:     "Content to delete",
		URL:         "https://example.com/to-delete",
		Category:    "Technology",
	}

	created, err := repo.Create(context.Background(), req)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), created.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "article not found")
}

func TestArticleRepository_Count(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	req := &database.CreateArticleRequest{
		Title:       "Count Test",
		Content:     "Content for counting",
		URL:         "https://example.com/count-test",
		Category:    "Technology",
	}

	_, err := repo.Create(context.Background(), req)
	require.NoError(t, err)

	filter := &database.ArticleFilter{}
	count, err := repo.Count(context.Background(), filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestArticleRepository_GetRecent(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	articles := []*database.CreateArticleRequest{
		{
			Title:       "Recent 1",
			Content:     "Content 1",
			URL:         "https://example.com/recent1",
			Category:    "Technology",
			},
		{
			Title:       "Recent 2",
			Content:     "Content 2", 
			URL:         "https://example.com/recent2",
			Category:    "Technology",
			},
	}

	for _, req := range articles {
		_, err := repo.Create(context.Background(), req)
		require.NoError(t, err)
	}

	results, err := repo.GetRecent(context.Background(), 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
	assert.True(t, results[0].CreatedAt.After(results[1].CreatedAt) || results[0].CreatedAt.Equal(results[1].CreatedAt))
}