package handlers

import (
	"context"
	"net/http"
	"time"

	"cinema-reservation/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

type HealthStatus struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
	Time     time.Time         `json:"time"`
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := HealthStatus{
		Status:   "healthy",
		Services: make(map[string]string),
		Time:     time.Now(),
	}

	// Check PostgreSQL
	sqlDB, err := h.db.DB()
	if err != nil {
		health.Status = "unhealthy"
		health.Services["postgres"] = "error: " + err.Error()
	} else {
		err = sqlDB.PingContext(ctx)
		if err != nil {
			health.Status = "unhealthy"
			health.Services["postgres"] = "error: " + err.Error()
		} else {
			health.Services["postgres"] = "healthy"
		}
	}

	// Check Redis
	_, err = h.redis.Ping(ctx).Result()
	if err != nil {
		health.Status = "unhealthy"
		health.Services["redis"] = "error: " + err.Error()
	} else {
		health.Services["redis"] = "healthy"
	}

	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	utils.SuccessResponse(c, statusCode, "Health check completed", health)
}
