package main

import (
	"fmt"
	"log"

	"saas-backend/config"
	"saas-backend/database"
	"saas-backend/internal/handler"
	"saas-backend/internal/repository"
	"saas-backend/internal/router"
	"saas-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)

	// Initialize Gemini service (optional - will not fail if API key is missing)
	var geminiService *service.GeminiService
	geminiService, err = service.NewGeminiService(cfg)
	if err != nil {
		log.Printf("Warning: Gemini service not initialized: %v", err)
	}

	// Initialize services
	authService := service.NewAuthService(userRepo, orgRepo, refreshTokenRepo, cfg)
	taskService := service.NewTaskService(taskRepo, auditLogRepo, geminiService)
	issueService := service.NewIssueService(issueRepo, auditLogRepo, geminiService)
	reportService := service.NewReportService(taskRepo, issueRepo, auditLogRepo, geminiService)
	userService := service.NewUserService(userRepo, auditLogRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, cfg)
	taskHandler := handler.NewTaskHandler(taskService)
	issueHandler := handler.NewIssueHandler(issueService)
	userHandler := handler.NewUserHandler(userService)
	reportHandler := handler.NewReportHandler(reportService)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Setup routes
	router.SetupRoutes(r, cfg, authHandler, taskHandler, issueHandler, userHandler, reportHandler, auditLogHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
