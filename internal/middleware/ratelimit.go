package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"saas-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type rateLimiter struct {
	requests map[string]*clientRate
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientRate struct {
	count     int
	resetTime time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string]*clientRate),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, client := range rl.requests {
			if now.After(client.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.requests[key]

	if !exists || now.After(client.resetTime) {
		rl.requests[key] = &clientRate{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if client.count >= rl.limit {
		return false
	}

	client.count++
	return true
}

var (
	defaultLimiter = newRateLimiter(100, time.Minute)
	uploadLimiter  = newRateLimiter(10, time.Minute)
	aiLimiter      = newRateLimiter(20, time.Minute)
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getClientKey(c)

		if !defaultLimiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getClientKey(c)

		if !uploadLimiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "upload rate limit exceeded, please wait before uploading again",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitAI() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getClientKey(c)

		if !aiLimiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "AI request rate limit exceeded, please wait before trying again",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getClientKey(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		switch v := userID.(type) {
		case string:
			if v != "" {
				return v
			}
		case uuid.UUID:
			return v.String()
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return c.ClientIP()
}
