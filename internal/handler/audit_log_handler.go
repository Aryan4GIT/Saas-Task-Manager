package handler

import (
	"net/http"
	"strconv"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	auditRepo *repository.AuditLogRepository
}

func NewAuditLogHandler(auditRepo *repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{auditRepo: auditRepo}
}

func (h *AuditLogHandler) List(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 200 {
		limit = 200
	}

	logs, err := h.auditRepo.List(orgID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to list audit logs",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}
