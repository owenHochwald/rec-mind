package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() error {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPortStr := os.Getenv("REDIS_PORT")
	if redisPortStr == "" {
		redisPortStr = "6379"
	}

	redisPort, err := strconv.Atoi(redisPortStr)
	if err != nil {
		return fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		redisDB, err = strconv.Atoi(dbStr)
		if err != nil {
			return fmt.Errorf("invalid REDIS_DB: %w", err)
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password:     redisPassword,
		DB:           redisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	return nil
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

func HealthCheck(ctx context.Context) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	_, err := RedisClient.Ping(ctx).Result()
	return err
}