package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/owenHochwald/rec-mind-api/config"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewConnection(cfg *config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MaxConnIdleTime = cfg.MaxIdleTime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Database connection established (Max Connections: %d)", cfg.MaxConnections)

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("🔒 Database connection closed")
	}
}

func (db *DB) HealthCheck(ctx context.Context) error {
	if db.Pool == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	stats := db.Pool.Stat()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no database connections available")
	}

	return nil
}

func (db *DB) GetStats() map[string]interface{} {
	if db.Pool == nil {
		return map[string]interface{}{
			"status": "disconnected",
		}
	}

	stats := db.Pool.Stat()
	return map[string]interface{}{
		"status":           "connected",
		"total_conns":      stats.TotalConns(),
		"idle_conns":       stats.IdleConns(),
		"acquired_conns":   stats.AcquiredConns(),
		"constructing_conns": stats.ConstructingConns(),
	}
}