package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Status      HealthStatus       `json:"status"`
	Timestamp   time.Time          `json:"timestamp"`
	Version     string             `json:"version"`
	Checks      map[string]Check   `json:"checks"`
}

// Check represents an individual component check
type Check struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	db      *gorm.DB
	redis   *redis.Client
	version string
}

func NewHealthChecker(db *gorm.DB, redis *redis.Client, version string) *HealthChecker {
	return &HealthChecker{
		db:      db,
		redis:   redis,
		version: version,
	}
}

// CheckHealth performs all health checks
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthCheck {
	checks := make(map[string]Check)
	
	// Check database
	checks["database"] = h.checkDatabase(ctx)
	
	// Check Redis (if available)
	if h.redis != nil {
		checks["redis"] = h.checkRedis(ctx)
	}
	
	// Determine overall status
	overallStatus := StatusHealthy
	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if check.Status == StatusDegraded {
			overallStatus = StatusDegraded
		}
	}
	
	return &HealthCheck{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   h.version,
		Checks:    checks,
	}
}

func (h *HealthChecker) checkDatabase(ctx context.Context) Check {
	start := time.Now()
	
	sqlDB, err := h.db.DB()
	if err != nil {
		return Check{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("failed to get DB: %v", err),
		}
	}
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return Check{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}
	
	latency := time.Since(start)
	
	// Check connection pool
	stats := sqlDB.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections {
		return Check{
			Status:  StatusDegraded,
			Message: "connection pool exhausted",
			Latency: latency.String(),
		}
	}
	
	return Check{
		Status:  StatusHealthy,
		Latency: latency.String(),
	}
}

func (h *HealthChecker) checkRedis(ctx context.Context) Check {
	start := time.Now()
	
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return Check{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}
	
	latency := time.Since(start)
	
	return Check{
		Status:  StatusHealthy,
		Latency: latency.String(),
	}
}

// CheckReadiness performs readiness checks (stricter than health)
func (h *HealthChecker) CheckReadiness(ctx context.Context) *HealthCheck {
	health := h.CheckHealth(ctx)
	
	// Additional readiness checks
	// For example, check if migrations are complete
	if err := h.checkMigrations(ctx); err != nil {
		health.Checks["migrations"] = Check{
			Status:  StatusUnhealthy,
			Message: err.Error(),
		}
		health.Status = StatusUnhealthy
	}
	
	return health
}

func (h *HealthChecker) checkMigrations(ctx context.Context) error {
	// Simple check: ensure key tables exist
	var count int64
	if err := h.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").
		Scan(&count).Error; err != nil {
		return fmt.Errorf("migration check failed: %w", err)
	}
	
	if count == 0 {
		return fmt.Errorf("no tables found, migrations may not be complete")
	}
	
	return nil
}
