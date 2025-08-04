package migrations

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MigrationRunner struct {
	db *pgxpool.Pool
}

func NewMigrationRunner(db *pgxpool.Pool) *MigrationRunner {
	return &MigrationRunner{db: db}
}

// RunMigrations executes all SQL files in the migrations directory
func (mr *MigrationRunner) RunMigrations(ctx context.Context, migrationsDir string) error {
	// Create migrations tracking table if it doesn't exist
	err := mr.createMigrationsTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	migrationFiles, err := mr.getMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Execute each migration
	for _, filename := range migrationFiles {
		executed, err := mr.isMigrationExecuted(ctx, filename)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		}

		if executed {
			fmt.Printf("⏭️  Migration %s already executed, skipping\n", filename)
			continue
		}

		fmt.Printf("🔄 Running migration: %s\n", filename)
		err = mr.executeMigration(ctx, migrationsDir, filename)
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		err = mr.markMigrationExecuted(ctx, filename)
		if err != nil {
			return fmt.Errorf("failed to mark migration as executed %s: %w", filename, err)
		}

		fmt.Printf("✅ Migration %s completed successfully\n", filename)
	}

	fmt.Println("🎉 All migrations completed successfully!")
	return nil
}

func (mr *MigrationRunner) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := mr.db.Exec(ctx, query)
	return err
}

func (mr *MigrationRunner) getMigrationFiles(migrationsDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, d.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files to ensure they run in order
	sort.Strings(files)
	return files, nil
}

func (mr *MigrationRunner) isMigrationExecuted(ctx context.Context, filename string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1"
	err := mr.db.QueryRow(ctx, query, filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (mr *MigrationRunner) executeMigration(ctx context.Context, migrationsDir, filename string) error {
	filePath := filepath.Join(migrationsDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute the SQL content
	_, err = mr.db.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}

func (mr *MigrationRunner) markMigrationExecuted(ctx context.Context, filename string) error {
	query := "INSERT INTO schema_migrations (filename) VALUES ($1)"
	_, err := mr.db.Exec(ctx, query, filename)
	return err
}