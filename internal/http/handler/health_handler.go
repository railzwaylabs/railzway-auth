package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health and readiness checks
type HealthHandler struct {
	db    *pgxpool.Pool
	redis redis.UniversalClient
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool, redis redis.UniversalClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Health returns a simple liveness check
// This endpoint should always return 200 if the process is running
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Ready returns readiness status including dependency checks
// This endpoint returns 503 if any critical dependency is unavailable
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	// Check PostgreSQL
	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		checks["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		checks["redis"] = "healthy"
	}

	if !allHealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": checks,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": checks,
	})
}
