package health

import (
	"context"
	"time"

	"rec-mind/internal/database"
)

type DatabaseHealth struct {
	Status      string                 `json:"status"`
	ResponseTime string                `json:"response_time"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	Tables      *TableHealth           `json:"tables,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

type TableHealth struct {
	Articles      *TableInfo `json:"articles"`
	ArticleChunks *TableInfo `json:"article_chunks"`
}

type TableInfo struct {
	Exists bool  `json:"exists"`
	Count  int64 `json:"count"`
	Error  string `json:"error,omitempty"`
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
		health.Tables = checkTableHealth(ctx, db)
	}

	health.ResponseTime = time.Since(start).String()
	return health
}

func checkTableHealth(ctx context.Context, db *database.DB) *TableHealth {
	return &TableHealth{
		Articles:      checkTableInfo(ctx, db, "articles"),
		ArticleChunks: checkTableInfo(ctx, db, "article_chunks"),
	}
}

func checkTableInfo(ctx context.Context, db *database.DB, tableName string) *TableInfo {
	info := &TableInfo{
		Exists: false,
		Count:  0,
	}

	existsQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`
	
	var exists bool
	err := db.Pool.QueryRow(ctx, existsQuery, tableName).Scan(&exists)
	if err != nil {
		info.Error = "failed to check table existence: " + err.Error()
		return info
	}

	info.Exists = exists
	if !exists {
		return info
	}

	countQuery := `SELECT COUNT(*) FROM ` + tableName
	err = db.Pool.QueryRow(ctx, countQuery).Scan(&info.Count)
	if err != nil {
		info.Error = "failed to count table rows: " + err.Error()
	}

	return info
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