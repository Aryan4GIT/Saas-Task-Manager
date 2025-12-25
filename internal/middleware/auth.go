package middleware

import (
	"context"
	"net/http"
	"strings"

	"saas-backend/config"
	"saas-backend/internal/models"
	"saas-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	OrgIDKey  contextKey = "org_id"
	RoleKey   contextKey = "role"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, cfg.JWT.AccessSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Add claims to context
		ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, OrgIDKey, claims.OrgID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		c.Request = c.Request.WithContext(ctx)

		// Also set in Gin context for easier access
		c.Set("user_id", claims.UserID)
		c.Set("org_id", claims.OrgID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "role not found in context",
			})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "invalid role type",
			})
			c.Abort()
			return
		}

		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "insufficient permissions",
		})
		c.Abort()
	}
}

// Helper functions to extract values from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}

func GetOrgID(c *gin.Context) (uuid.UUID, bool) {
	orgID, exists := c.Get("org_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := orgID.(uuid.UUID)
	return id, ok
}

func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}
