package router

import (
	"saas-backend/config"
	"saas-backend/internal/handler"
	"saas-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	taskHandler *handler.TaskHandler,
	issueHandler *handler.IssueHandler,
	userHandler *handler.UserHandler,
	reportHandler *handler.ReportHandler,
	auditLogHandler *handler.AuditLogHandler,
) {
	// Apply global middleware
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Auth routes
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.Me)

			// Task routes
			tasks := protected.Group("/tasks")
			{
				tasks.POST("", middleware.RequireRole("admin", "manager"), taskHandler.CreateTask)
				tasks.GET("", taskHandler.ListTasks)
				tasks.GET("/my", taskHandler.ListMyTasks)
				tasks.GET("/ai-report", middleware.RequireRole("admin"), taskHandler.AdminAIReport)
				tasks.GET("/:id", taskHandler.GetTask)
				tasks.PATCH("/:id", taskHandler.UpdateTask)
				tasks.DELETE("/:id", middleware.RequireRole("admin", "manager"), taskHandler.DeleteTask)
				// Workflow actions
				tasks.POST("/:id/done", taskHandler.MarkDone)
				tasks.POST("/:id/verify", taskHandler.VerifyTask)
				tasks.POST("/:id/approve", taskHandler.ApproveTask)
				tasks.POST("/:id/reject", taskHandler.RejectTask)
			}

			// Issue routes
			issues := protected.Group("/issues")
			{
				issues.POST("", issueHandler.CreateIssue)
				issues.GET("", issueHandler.ListIssues)
				issues.GET("/:id", issueHandler.GetIssue)
				issues.PATCH("/:id", issueHandler.UpdateIssue)
				issues.DELETE("/:id", middleware.RequireRole("admin", "manager"), issueHandler.DeleteIssue)
			}

			// Reports (admin/manager)
			reports := protected.Group("/reports")
			reports.Use(middleware.RequireRole("admin", "manager"))
			{
				reports.GET("/weekly-summary", reportHandler.WeeklySummary)
			}

			// Audit logs (admin only)
			audit := protected.Group("/audit-logs")
			audit.Use(middleware.RequireRole("admin"))
			{
				audit.GET("", auditLogHandler.List)
			}

			// User routes (admin/manager only)
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin", "manager"))
			{
				users.POST("", userHandler.CreateUser)
				users.GET("", userHandler.ListUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PATCH("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", middleware.RequireRole("admin"), userHandler.DeleteUser)
			}
		}
	}
}
