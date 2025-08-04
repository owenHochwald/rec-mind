package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/owenHochwald/rec-mind-api/config"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
)

// ArticleProcessingMessage represents the message sent to RabbitMQ for article processing
type ArticleProcessingMessage struct {
	ArticleID     uuid.UUID `json:"article_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Category      string    `json:"category"`
	URL           string    `json:"url"`
	CorrelationID string    `json:"correlation_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// ScraperService handles RSS feed scraping and article processing
type ScraperService struct {
	repo      repository.ArticleRepository
	mqChannel *amqp.Channel
	config    config.ScraperConfig
	parser    *gofeed.Parser
}

// ScrapingResult represents the result of a scraping operation
type ScrapingResult struct {
	TotalFeeds      int                `json:"total_feeds"`
	TotalArticles   int                `json:"total_articles"`
	ProcessedCount  int                `json:"processed_count"`
	SkippedCount    int                `json:"skipped_count"`
	ErrorCount      int                `json:"error_count"`
	ProcessingTime  time.Duration      `json:"processing_time"`
	FeedResults     []FeedScrapingResult `json:"feed_results"`
}

// FeedScrapingResult represents the result of scraping a single feed
type FeedScrapingResult struct {
	FeedName       string        `json:"feed_name"`
	FeedURL        string        `json:"feed_url"`
	Category       string        `json:"category"`
	ArticlesFound  int           `json:"articles_found"`
	ArticlesSaved  int           `json:"articles_saved"`
	ArticlesSkipped int          `json:"articles_skipped"`
	Errors         []string      `json:"errors,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// NewScraperService creates a new scraper service
func NewScraperService(repo repository.ArticleRepository, mqChannel *amqp.Channel) *ScraperService {
	return &ScraperService{
		repo:      repo,
		mqChannel: mqChannel,
		config:    config.GetScraperConfig(),
		parser:    gofeed.NewParser(),
	}
}

// ScrapeAllFeeds scrapes all configured RSS feeds
func (s *ScraperService) ScrapeAllFeeds(ctx context.Context) (*ScrapingResult, error) {
	startTime := time.Now()
	
	result := &ScrapingResult{
		TotalFeeds:    len(s.config.Feeds),
		FeedResults:   make([]FeedScrapingResult, 0, len(s.config.Feeds)),
	}

	log.Printf("🔍 Starting to scrape %d RSS feeds", len(s.config.Feeds))

	// Declare the article_processing queue
	if err := s.declareQueue("article_processing"); err != nil {
		return nil, fmt.Errorf("failed to declare article_processing queue: %w", err)
	}

	for i, feed := range s.config.Feeds {
		// Apply rate limiting between feeds
		if i > 0 {
			time.Sleep(time.Duration(s.config.RateLimit.DelaySeconds) * time.Second)
		}

		feedResult := s.scrapeFeed(ctx, feed)
		result.FeedResults = append(result.FeedResults, feedResult)
		result.TotalArticles += feedResult.ArticlesFound
		result.ProcessedCount += feedResult.ArticlesSaved
		result.SkippedCount += feedResult.ArticlesSkipped
		result.ErrorCount += len(feedResult.Errors)

		log.Printf("📰 Feed '%s': %d articles found, %d saved, %d skipped, %d errors", 
			feed.Name, feedResult.ArticlesFound, feedResult.ArticlesSaved, 
			feedResult.ArticlesSkipped, len(feedResult.Errors))
	}

	result.ProcessingTime = time.Since(startTime)
	
	log.Printf("✅ Scraping completed: %d total articles, %d processed, %d skipped, %d errors in %v",
		result.TotalArticles, result.ProcessedCount, result.SkippedCount, result.ErrorCount, result.ProcessingTime)

	return result, nil
}

// scrapeFeed scrapes a single RSS feed
func (s *ScraperService) scrapeFeed(ctx context.Context, feedConfig config.RSSFeed) FeedScrapingResult {
	startTime := time.Now()
	
	result := FeedScrapingResult{
		FeedName: feedConfig.Name,
		FeedURL:  feedConfig.URL,
		Category: feedConfig.Category,
		Errors:   make([]string, 0),
	}

	log.Printf("📡 Scraping feed: %s (%s)", feedConfig.Name, feedConfig.URL)

	// Parse the RSS feed
	feed, err := s.parser.ParseURL(feedConfig.URL)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse RSS feed %s: %v", feedConfig.URL, err)
		result.Errors = append(result.Errors, errMsg)
		log.Printf("❌ %s", errMsg)
		result.ProcessingTime = time.Since(startTime)
		return result
	}

	result.ArticlesFound = len(feed.Items)

	// Process each article in the feed
	for _, item := range feed.Items {
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, "Context cancelled during feed processing")
			result.ProcessingTime = time.Since(startTime)
			return result
		default:
			// Continue processing
		}

		if err := s.processArticle(ctx, item, feedConfig.Category); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to process article '%s': %v", item.Title, err))
			continue
		}

		result.ArticlesSaved++
	}

	result.ArticlesSkipped = result.ArticlesFound - result.ArticlesSaved
	result.ProcessingTime = time.Since(startTime)
	
	return result
}

// processArticle processes a single article from RSS feed
func (s *ScraperService) processArticle(ctx context.Context, item *gofeed.Item, category string) error {
	// Clean and extract content
	title := s.cleanText(item.Title)
	content := s.extractContent(item)
	url := item.Link

	// Validate article
	if !s.validateArticle(title, content) {
		return fmt.Errorf("article validation failed")
	}

	// Check for duplicates by URL
	exists, err := s.repo.ExistsByURL(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to check duplicate: %w", err)
	}
	if exists {
		return fmt.Errorf("article already exists with URL: %s", url)
	}

	// Create article in database
	articleReq := &database.CreateArticleRequest{
		Title:    title,
		Content:  content,
		URL:      url,
		Category: category,
	}

	article, err := s.repo.Create(ctx, articleReq)
	if err != nil {
		return fmt.Errorf("failed to create article: %w", err)
	}

	// Publish to RabbitMQ for ML processing
	if err := s.publishArticleProcessingMessage(article); err != nil {
		log.Printf("⚠️ Failed to publish article processing message for %s: %v", article.ID, err)
		// Don't return error - article is already saved
	}

	log.Printf("✅ Processed article: %s (ID: %s)", title, article.ID)
	return nil
}

// validateArticle validates article title and content length
func (s *ScraperService) validateArticle(title, content string) bool {
	titleLen := len(title)
	contentLen := len(content)

	if titleLen < s.config.Validation.MinTitleLength || titleLen > s.config.Validation.MaxTitleLength {
		log.Printf("⚠️ Title length validation failed: %d chars (min: %d, max: %d)", 
			titleLen, s.config.Validation.MinTitleLength, s.config.Validation.MaxTitleLength)
		return false
	}

	if contentLen < s.config.Validation.MinContentLength || contentLen > s.config.Validation.MaxContentLength {
		log.Printf("⚠️ Content length validation failed: %d chars (min: %d, max: %d)", 
			contentLen, s.config.Validation.MinContentLength, s.config.Validation.MaxContentLength)
		return false
	}

	return true
}

// cleanText removes HTML tags and cleans text content
func (s *ScraperService) cleanText(text string) string {
	// Remove HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]+>`)
	cleaned := htmlTagRegex.ReplaceAllString(text, "")
	
	// Remove extra whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")
	
	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// extractContent extracts and cleans article content from RSS item
func (s *ScraperService) extractContent(item *gofeed.Item) string {
	var content string
	
	// Prefer content over description
	if item.Content != "" {
		content = item.Content
	} else if item.Description != "" {
		content = item.Description
	} else {
		content = item.Title // Fallback to title if no content
	}

	// Clean the content
	cleaned := s.cleanText(content)
	
	// Remove common ads and navigation text patterns
	adPatterns := []string{
		"Advertisement",
		"Click here",
		"Read more",
		"Subscribe",
		"Newsletter",
		"Follow us",
		"Share this",
		"Related articles",
		"Trending now",
	}
	
	for _, pattern := range adPatterns {
		cleaned = strings.ReplaceAll(cleaned, pattern, "")
	}
	
	// Final cleanup
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// publishArticleProcessingMessage publishes article to RabbitMQ for ML processing
func (s *ScraperService) publishArticleProcessingMessage(article *database.Article) error {
	message := ArticleProcessingMessage{
		ArticleID:     article.ID,
		Title:         article.Title,
		Content:       article.Content,
		Category:      article.Category,
		URL:           article.URL,
		CorrelationID: uuid.New().String(),
		CreatedAt:     time.Now(),
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = s.mqChannel.Publish(
		"",                   // exchange
		"article_processing", // routing key (queue name)
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
			Headers: amqp.Table{
				"correlation_id": message.CorrelationID,
				"article_id":     message.ArticleID.String(),
				"category":       message.Category,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("📤 Published article processing message for article %s", article.ID)
	return nil
}

// declareQueue declares a RabbitMQ queue
func (s *ScraperService) declareQueue(queueName string) error {
	_, err := s.mqChannel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err
}