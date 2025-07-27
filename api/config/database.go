package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	
	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host           string
	Port           int
	Name           string
	User           string
	Password       string
	SSLMode        string
	MaxConnections int32
	MaxIdleTime    time.Duration
}

func LoadDatabaseConfig() (*DatabaseConfig, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default values")
	}

	port, err := strconv.Atoi(getEnvOrDefault("DB_PORT", "5431"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	maxConnections, err := strconv.Atoi(getEnvOrDefault("DB_MAX_CONNECTIONS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONNECTIONS: %w", err)
	}

	maxIdleTime, err := time.ParseDuration(getEnvOrDefault("DB_MAX_IDLE_TIME", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_TIME: %w", err)
	}

	config := &DatabaseConfig{
		Host:           getEnvOrDefault("DB_HOST", "localhost"),
		Port:           port,
		Name:           getEnvOrDefault("DB_NAME", "postgres"),
		User:           getEnvOrDefault("DB_USER", "postgres"),
		Password:       getEnvOrDefault("DB_PASSWORD", "secret"),
		SSLMode:        getEnvOrDefault("DB_SSL_MODE", "disable"),
		MaxConnections: int32(maxConnections),
		MaxIdleTime:    maxIdleTime,
	}

	return config, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}