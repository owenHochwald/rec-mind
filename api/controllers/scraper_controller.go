package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"rec-mind/internal/services"
)

// ScrapeArticles triggers RSS feed scraping
func ScrapeArticles(scraperService *services.ScraperService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set timeout for scraping operation
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
		defer cancel()

		// Run scraper
		result, err := scraperService.ScrapeAllFeeds(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scrape articles",
				"details": err.Error(),
			})
			return
		}

		// Return comprehensive results
		response := gin.H{
			"success": true,
			"message": "Scraping completed successfully",
			"summary": gin.H{
				"total_feeds": result.TotalFeeds,
				"total_articles_found": result.TotalArticles,
				"articles_processed": result.ProcessedCount,
				"articles_skipped": result.SkippedCount,
				"errors": result.ErrorCount,
				"processing_time": result.ProcessingTime.String(),
			},
			"feed_results": result.FeedResults,
		}

		// Set status based on results
		if result.ProcessedCount > 0 {
			c.JSON(http.StatusOK, response)
		} else if result.ErrorCount > 0 {
			response["success"] = false
			response["message"] = "Scraping completed but no articles were processed due to errors"
			c.JSON(http.StatusPartialContent, response)
		} else {
			response["success"] = false  
			response["message"] = "No new articles found to process"
			c.JSON(http.StatusOK, response)
		}
	}
}