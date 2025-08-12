package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rec-mind/config"
)

func TestNewConnection_Success(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5431,
		Name:           "postgres", 
		User:           "postgres",
		Password:       "secret",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    15 * time.Minute,
	}

	db, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available (%v)", err)
		return
	}
	defer db.Close()

	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.Pool)
}

func TestNewConnection_InvalidCredentials(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5431,
		Name:           "nonexistent_db",
		User:           "invalid_user",
		Password:       "wrong_password",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    15 * time.Minute,
	}

	db, err := NewConnection(cfg)
	
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to")
}

func TestDB_HealthCheck(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5431,
		Name:           "postgres",
		User:           "postgres", 
		Password:       "secret",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    15 * time.Minute,
	}

	db, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available (%v)", err)
		return
	}
	defer db.Close()

	ctx := context.Background()
	err = db.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestDB_HealthCheck_Timeout(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5431,
		Name:           "postgres",
		User:           "postgres",
		Password:       "secret", 
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    15 * time.Minute,
	}

	db, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available (%v)", err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	err = db.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestDB_GetStats(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5431,
		Name:           "postgres",
		User:           "postgres",
		Password:       "secret",
		SSLMode:        "disable", 
		MaxConnections: 5,
		MaxIdleTime:    15 * time.Minute,
	}

	db, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available (%v)", err)
		return
	}
	defer db.Close()

	stats := db.GetStats()
	
	assert.NotNil(t, stats)
	assert.Equal(t, "connected", stats["status"])
	assert.Contains(t, stats, "total_conns")
	assert.Contains(t, stats, "idle_conns")
	assert.Contains(t, stats, "acquired_conns")
}

func TestDB_GetStats_Disconnected(t *testing.T) {
	db := &DB{Pool: nil}
	
	stats := db.GetStats()
	
	assert.NotNil(t, stats)
	assert.Equal(t, "disconnected", stats["status"])
}