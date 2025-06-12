package middleware

import (
	"time"

	"cinema-reservation/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rate_limit:" + c.ClientIP()

		// Get current count
		count, err := rl.redis.Get(c.Request.Context(), key).Int()
		if err != nil && err != redis.Nil {
			// If Redis is down, allow the request
			c.Next()
			return
		}

		if count >= rl.limit {
			utils.ErrorResponse(c, utils.ErrRateLimitExceeded)
			c.Abort()
			return
		}

		// Increment counter
		pipe := rl.redis.Pipeline()
		pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, rl.window)
		_, err = pipe.Exec(c.Request.Context())
		if err != nil {
			// If Redis is down, allow the request
			c.Next()
			return
		}

		c.Next()
	}
}
