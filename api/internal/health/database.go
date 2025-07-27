package health

import (
	"context"
	"time"

	"github.com/owenHochwald/rec-mind-api/internal/database"
)

type DatabaseHealth struct {
	Status      string                 `json:"status"`
	ResponseTime string                `json:"response_time"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

func CheckDatabase(db *database.DB) *DatabaseHealth {
	start := time.Now()
	
	health := &DatabaseHealth{
		Status: "unknown",
		Stats:  db.GetStats(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
	} else {
		health.Status = "healthy"
	}

	health.ResponseTime = time.Since(start).String()
	return health
}

type SystemHealth struct {
	Status   string          `json:"status"`
	Database *DatabaseHealth `json:"database"`
	Uptime   string          `json:"uptime"`
	Version  string          `json:"version,omitempty"`
}

func CheckSystemHealth(db *database.DB, startTime time.Time) *SystemHealth {
	dbHealth := CheckDatabase(db)
	
	status := "healthy"
	if dbHealth.Status != "healthy" {
		status = "unhealthy"
	}

	return &SystemHealth{
		Status:   status,
		Database: dbHealth,
		Uptime:   time.Since(startTime).String(),
		Version:  "1.0.0",
	}
}