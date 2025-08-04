package config

import (
    "log"
    "github.com/joho/godotenv"
)

type RSSFeed struct {
    Name     string `json:"name"`
    URL      string `json:"url"`
    Category string `json:"category"`
}

type ScraperConfig struct {
    Feeds []RSSFeed `json:"feeds"`
    RateLimit struct {
        DelaySeconds int `json:"delay_seconds"`
    } `json:"rate_limit"`
    Validation struct {
        MinTitleLength   int `json:"min_title_length"`
        MaxTitleLength   int `json:"max_title_length"`
        MinContentLength int `json:"min_content_length"`
        MaxContentLength int `json:"max_content_length"`
    } `json:"validation"`
}

func GetScraperConfig() ScraperConfig {
    return ScraperConfig{
        Feeds: []RSSFeed{
            {
                Name:     "TechCrunch",
                URL:      "https://techcrunch.com/feed/",
                Category: "technology",
            },
            {
                Name:     "BBC News",
                URL:      "http://feeds.bbci.co.uk/news/rss.xml",
                Category: "general",
            },
            {
                Name:     "Reuters Business",
                URL:      "https://feeds.reuters.com/reuters/businessNews",
                Category: "business",
            },
            {
                Name:     "ESPN",
                URL:      "https://www.espn.com/espn/rss/news",
                Category: "sports",
            },
        },
        RateLimit: struct {
            DelaySeconds int `json:"delay_seconds"`
        }{
            DelaySeconds: 1,
        },
        Validation: struct {
            MinTitleLength   int `json:"min_title_length"`
            MaxTitleLength   int `json:"max_title_length"`
            MinContentLength int `json:"min_content_length"`
            MaxContentLength int `json:"max_content_length"`
        }{
            MinTitleLength:   10,
            MaxTitleLength:   500,
            MinContentLength: 50,  // Reduced from 200 to handle RSS descriptions
            MaxContentLength: 10000,
        },
    }
}

func LoadEnv() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found (skipping), using OS env vars")
    }
}
