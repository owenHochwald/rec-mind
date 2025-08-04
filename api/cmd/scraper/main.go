package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/owenHochwald/rec-mind-api/config"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
	"github.com/owenHochwald/rec-mind-api/internal/services"
	"github.com/owenHochwald/rec-mind-api/mq"
)

func main() {
	log.Println("🚀 Starting RSS Article Scraper")

	// Load environment variables
	config.LoadEnv()

	// Initialize database connection
	log.Println("🔌 Connecting to database...")
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize RabbitMQ
	log.Println("🐰 Connecting to RabbitMQ...")
	mq.InitRabbitMQ()
	defer mq.MQConn.Close()
	defer mq.MQChannel.Close()

	// Initialize repository
	articleRepo := repository.NewArticleRepository(db.Pool)

	// Initialize scraper service
	scraperService := services.NewScraperService(articleRepo, mq.MQChannel)

	// Run scraper
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Println("🔍 Starting article scraping process...")
	result, err := scraperService.ScrapeAllFeeds(ctx)
	if err != nil {
		log.Fatalf("Scraping failed: %v", err)
	}

	// Print results
	log.Println("\n📊 SCRAPING RESULTS:")
	log.Printf("Total Feeds: %d", result.TotalFeeds)
	log.Printf("Total Articles Found: %d", result.TotalArticles)
	log.Printf("Articles Processed: %d", result.ProcessedCount)
	log.Printf("Articles Skipped: %d", result.SkippedCount)
	log.Printf("Errors: %d", result.ErrorCount)
	log.Printf("Processing Time: %v", result.ProcessingTime)

	log.Println("\n📰 FEED DETAILS:")
	for _, feedResult := range result.FeedResults {
		log.Printf("Feed: %s (%s)", feedResult.FeedName, feedResult.Category)
		log.Printf("  - URL: %s", feedResult.FeedURL)
		log.Printf("  - Articles Found: %d", feedResult.ArticlesFound)
		log.Printf("  - Articles Saved: %d", feedResult.ArticlesSaved)
		log.Printf("  - Articles Skipped: %d", feedResult.ArticlesSkipped)
		log.Printf("  - Processing Time: %v", feedResult.ProcessingTime)
		if len(feedResult.Errors) > 0 {
			log.Printf("  - Errors: %d", len(feedResult.Errors))
			for _, err := range feedResult.Errors {
				log.Printf("    • %s", err)
			}
		}
		log.Println()
	}

	if result.ProcessedCount > 0 {
		log.Printf("✅ Successfully processed %d articles", result.ProcessedCount)
	} else {
		log.Println("⚠️ No articles were processed")
		os.Exit(1)
	}
}