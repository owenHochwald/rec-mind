package main

import (
	"context"
	"flag"
	"log"
	"time"

	"rec-mind/config"
	"rec-mind/internal/database"
	"rec-mind/internal/migrations"
)

func main() {
	var migrationsDir = flag.String("dir", "migrations", "Directory containing migration files")
	flag.Parse()

	log.Println("🚀 Starting database migrations...")

	// Load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create migration runner
	runner := migrations.NewMigrationRunner(db.Pool)

	// Run migrations with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = runner.RunMigrations(ctx, *migrationsDir)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ All migrations completed successfully!")
}