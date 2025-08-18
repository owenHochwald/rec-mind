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

func LoadEnv() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found (skipping), using OS env vars")
    }
}
