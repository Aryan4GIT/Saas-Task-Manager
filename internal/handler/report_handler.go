package handler

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) WeeklySummary(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	summary, err := h.reportService.GenerateWeeklySummary(orgID, userID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "AI summary unavailable",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}
