package middleware

import (
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		log.Printf(
			"[%s] %s %s | Status: %d | Duration: %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			statusCode,
			duration,
		)
	}
}
