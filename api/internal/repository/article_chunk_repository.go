package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"rec-mind/models"
)

type ArticleChunkRepository interface {
	Create(ctx context.Context, req *models.CreateArticleChunkRequest) (*models.ArticleChunk, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.ArticleChunk, error)
	GetByArticleID(ctx context.Context, articleID uuid.UUID) ([]*models.ArticleChunk, error)
	GetByArticleIDAndIndex(ctx context.Context, articleID uuid.UUID, chunkIndex int) (*models.ArticleChunk, error)
	List(ctx context.Context, filter *models.ArticleChunkFilter) ([]*models.ArticleChunk, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateArticleChunkRequest) (*models.ArticleChunk, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByArticleID(ctx context.Context, articleID uuid.UUID) error
	Count(ctx context.Context, filter *models.ArticleChunkFilter) (int64, error)
	CreateBatch(ctx context.Context, chunks []*models.CreateArticleChunkRequest) ([]*models.ArticleChunk, error)
}

type articleChunkRepository struct {
	db *pgxpool.Pool
}

func NewArticleChunkRepository(db *pgxpool.Pool) ArticleChunkRepository {
	return &articleChunkRepository{db: db}
}

func (r *articleChunkRepository) Create(ctx context.Context, req *models.CreateArticleChunkRequest) (*models.ArticleChunk, error) {
	query := `
		INSERT INTO article_chunks (article_id, chunk_index, content, token_count, character_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, article_id, chunk_index, content, token_count, character_count, created_at`

	var chunk models.ArticleChunk
	err := r.db.QueryRow(ctx, query, req.ArticleID, req.ChunkIndex, req.Content, req.TokenCount, req.CharacterCount).
		Scan(&chunk.ID, &chunk.ArticleID, &chunk.ChunkIndex, &chunk.Content, &chunk.TokenCount, &chunk.CharacterCount, &chunk.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create article chunk: %w", err)
	}

	return &chunk, nil
}

func (r *articleChunkRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ArticleChunk, error) {
	query := `
		SELECT id, article_id, chunk_index, content, token_count, character_count, created_at
		FROM article_chunks
		WHERE id = $1`

	var chunk models.ArticleChunk
	err := r.db.QueryRow(ctx, query, id).
		Scan(&chunk.ID, &chunk.ArticleID, &chunk.ChunkIndex, &chunk.Content, &chunk.TokenCount, &chunk.CharacterCount, &chunk.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("article chunk not found")
		}
		return nil, fmt.Errorf("failed to get article chunk: %w", err)
	}

	return &chunk, nil
}

func (r *articleChunkRepository) GetByArticleID(ctx context.Context, articleID uuid.UUID) ([]*models.ArticleChunk, error) {
	query := `
		SELECT id, article_id, chunk_index, content, token_count, character_count, created_at
		FROM article_chunks
		WHERE article_id = $1
		ORDER BY chunk_index ASC`

	rows, err := r.db.Query(ctx, query, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article chunks by article ID: %w", err)
	}
	defer rows.Close()

	var chunks []*models.ArticleChunk
	for rows.Next() {
		var chunk models.ArticleChunk
		err := rows.Scan(&chunk.ID, &chunk.ArticleID, &chunk.ChunkIndex, &chunk.Content, &chunk.TokenCount, &chunk.CharacterCount, &chunk.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article chunk: %w", err)
		}
		chunks = append(chunks, &chunk)
	}

	return chunks, nil
}

func (r *articleChunkRepository) GetByArticleIDAndIndex(ctx context.Context, articleID uuid.UUID, chunkIndex int) (*models.ArticleChunk, error) {
	query := `
		SELECT id, article_id, chunk_index, content, token_count, character_count, created_at
		FROM article_chunks
		WHERE article_id = $1 AND chunk_index = $2`

	var chunk models.ArticleChunk
	err := r.db.QueryRow(ctx, query, articleID, chunkIndex).
		Scan(&chunk.ID, &chunk.ArticleID, &chunk.ChunkIndex, &chunk.Content, &chunk.TokenCount, &chunk.CharacterCount, &chunk.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("article chunk not found")
		}
		return nil, fmt.Errorf("failed to get article chunk: %w", err)
	}

	return &chunk, nil
}

func (r *articleChunkRepository) List(ctx context.Context, filter *models.ArticleChunkFilter) ([]*models.ArticleChunk, error) {
	filter.SetDefaults()

	query := `
		SELECT id, article_id, chunk_index, content, token_count, character_count, created_at
		FROM article_chunks`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ArticleID != nil {
		conditions = append(conditions, fmt.Sprintf("article_id = $%d", argIndex))
		args = append(args, *filter.ArticleID)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY article_id, chunk_index ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list article chunks: %w", err)
	}
	defer rows.Close()

	var chunks []*models.ArticleChunk
	for rows.Next() {
		var chunk models.ArticleChunk
		err := rows.Scan(&chunk.ID, &chunk.ArticleID, &chunk.ChunkIndex, &chunk.Content, &chunk.TokenCount, &chunk.CharacterCount, &chunk.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article chunk: %w", err)
		}
		chunks = append(chunks, &chunk)
	}

	return chunks, nil
}

func (r *articleChunkRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateArticleChunkRequest) (*models.ArticleChunk, error) {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if req.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *req.Content)
		argIndex++
	}

	if req.TokenCount != nil {
		setParts = append(setParts, fmt.Sprintf("token_count = $%d", argIndex))
		args = append(args, *req.TokenCount)
		argIndex++
	}

	if req.CharacterCount != nil {
		setParts = append(setParts, fmt.Sprintf("character_count = $%d", argIndex))
		args = append(args, *req.CharacterCount)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE article_chunks
		SET %s
		WHERE id = $%d
		RETURNING id, article_id, chunk_index, content, token_count, character_count, created_at`,
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	var chunk models.ArticleChunk
	err := r.db.QueryRow(ctx, query, args...).
		Scan(&chunk.ID, &chunk.ArticleID, &chunk.ChunkIndex, &chunk.Content, &chunk.TokenCount, &chunk.CharacterCount, &chunk.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("article chunk not found")
		}
		return nil, fmt.Errorf("failed to update article chunk: %w", err)
	}

	return &chunk, nil
}

func (r *articleChunkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM article_chunks WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete article chunk: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article chunk not found")
	}

	return nil
}

func (r *articleChunkRepository) DeleteByArticleID(ctx context.Context, articleID uuid.UUID) error {
	query := `DELETE FROM article_chunks WHERE article_id = $1`

	_, err := r.db.Exec(ctx, query, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete article chunks by article ID: %w", err)
	}

	return nil
}

func (r *articleChunkRepository) Count(ctx context.Context, filter *models.ArticleChunkFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM article_chunks"

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ArticleID != nil {
		conditions = append(conditions, fmt.Sprintf("article_id = $%d", argIndex))
		args = append(args, *filter.ArticleID)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count article chunks: %w", err)
	}

	return count, nil
}

func (r *articleChunkRepository) CreateBatch(ctx context.Context, chunks []*models.CreateArticleChunkRequest) ([]*models.ArticleChunk, error) {
	if len(chunks) == 0 {
		return []*models.ArticleChunk{}, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO article_chunks (article_id, chunk_index, content, token_count, character_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, article_id, chunk_index, content, token_count, character_count, created_at`

	var results []*models.ArticleChunk
	for _, chunk := range chunks {
		var result models.ArticleChunk
		err := tx.QueryRow(ctx, query, chunk.ArticleID, chunk.ChunkIndex, chunk.Content, chunk.TokenCount, chunk.CharacterCount).
			Scan(&result.ID, &result.ArticleID, &result.ChunkIndex, &result.Content, &result.TokenCount, &result.CharacterCount, &result.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to create article chunk in batch: %w", err)
		}

		results = append(results, &result)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}