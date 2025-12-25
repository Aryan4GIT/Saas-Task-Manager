package middleware

import (
	"saas-backend/config"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		originNormalized := strings.TrimRight(origin, "/")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			allowedOrigin = strings.TrimRight(strings.TrimSpace(allowedOrigin), "/")
			if allowedOrigin == "*" {
				allowed = true
				break
			}
			if originNormalized != "" && originNormalized == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			// Echo the origin to support credentials while allowing configured origins.
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
