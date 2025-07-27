package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/owenHochwald/rec-mind-api/internal/database"
)

type ArticleRepository interface {
	Create(ctx context.Context, req *database.CreateArticleRequest) (*database.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*database.Article, error)
	GetByURL(ctx context.Context, url string) (*database.Article, error)
	List(ctx context.Context, filter *database.ArticleFilter) ([]*database.Article, error)
	Update(ctx context.Context, id uuid.UUID, req *database.UpdateArticleRequest) (*database.Article, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter *database.ArticleFilter) (int64, error)
	GetByCategory(ctx context.Context, category string, limit int) ([]*database.Article, error)
	GetRecent(ctx context.Context, limit int) ([]*database.Article, error)
}

type articleRepository struct {
	db *pgxpool.Pool
}

func NewArticleRepository(db *pgxpool.Pool) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(ctx context.Context, req *database.CreateArticleRequest) (*database.Article, error) {
	query := `
		INSERT INTO articles (title, content, url, category, published_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, content, url, category, published_at, created_at, updated_at`

	var article database.Article
	err := r.db.QueryRow(ctx, query, req.Title, req.Content, req.URL, req.Category, req.PublishedAt).
		Scan(&article.ID, &article.Title, &article.Content, &article.URL, &article.Category,
			&article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	return &article, nil
}

func (r *articleRepository) GetByID(ctx context.Context, id uuid.UUID) (*database.Article, error) {
	query := `
		SELECT id, title, content, url, category, published_at, created_at, updated_at
		FROM articles
		WHERE id = $1`

	var article database.Article
	err := r.db.QueryRow(ctx, query, id).
		Scan(&article.ID, &article.Title, &article.Content, &article.URL, &article.Category,
			&article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("article not found")
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return &article, nil
}

func (r *articleRepository) GetByURL(ctx context.Context, url string) (*database.Article, error) {
	query := `
		SELECT id, title, content, url, category, published_at, created_at, updated_at
		FROM articles
		WHERE url = $1`

	var article database.Article
	err := r.db.QueryRow(ctx, query, url).
		Scan(&article.ID, &article.Title, &article.Content, &article.URL, &article.Category,
			&article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("article not found")
		}
		return nil, fmt.Errorf("failed to get article by URL: %w", err)
	}

	return &article, nil
}

func (r *articleRepository) List(ctx context.Context, filter *database.ArticleFilter) ([]*database.Article, error) {
	filter.SetDefaults()

	query := `
		SELECT id, title, content, url, category, published_at, created_at, updated_at
		FROM articles`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *filter.Category)
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("published_at >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("published_at <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filter.SearchTerm+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY published_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list articles: %w", err)
	}
	defer rows.Close()

	var articles []*database.Article
	for rows.Next() {
		var article database.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.URL,
			&article.Category, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, &article)
	}

	return articles, nil
}

func (r *articleRepository) Update(ctx context.Context, id uuid.UUID, req *database.UpdateArticleRequest) (*database.Article, error) {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *req.Content)
		argIndex++
	}

	if req.URL != nil {
		setParts = append(setParts, fmt.Sprintf("url = $%d", argIndex))
		args = append(args, *req.URL)
		argIndex++
	}

	if req.Category != nil {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *req.Category)
		argIndex++
	}

	if req.PublishedAt != nil {
		setParts = append(setParts, fmt.Sprintf("published_at = $%d", argIndex))
		args = append(args, *req.PublishedAt)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf(`
		UPDATE articles
		SET %s
		WHERE id = $%d
		RETURNING id, title, content, url, category, published_at, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	var article database.Article
	err := r.db.QueryRow(ctx, query, args...).
		Scan(&article.ID, &article.Title, &article.Content, &article.URL, &article.Category,
			&article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("article not found")
		}
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	return &article, nil
}

func (r *articleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM articles WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	return nil
}

func (r *articleRepository) Count(ctx context.Context, filter *database.ArticleFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM articles"

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *filter.Category)
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("published_at >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("published_at <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filter.SearchTerm+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count articles: %w", err)
	}

	return count, nil
}

func (r *articleRepository) GetByCategory(ctx context.Context, category string, limit int) ([]*database.Article, error) {
	query := `
		SELECT id, title, content, url, category, published_at, created_at, updated_at
		FROM articles
		WHERE category = $1
		ORDER BY published_at DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, category, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles by category: %w", err)
	}
	defer rows.Close()

	var articles []*database.Article
	for rows.Next() {
		var article database.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.URL,
			&article.Category, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, &article)
	}

	return articles, nil
}

func (r *articleRepository) GetRecent(ctx context.Context, limit int) ([]*database.Article, error) {
	query := `
		SELECT id, title, content, url, category, published_at, created_at, updated_at
		FROM articles
		ORDER BY published_at DESC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent articles: %w", err)
	}
	defer rows.Close()

	var articles []*database.Article
	for rows.Next() {
		var article database.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.URL,
			&article.Category, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, &article)
	}

	return articles, nil
}