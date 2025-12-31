package handler

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/service"
	"saas-backend/internal/utils"

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
		utils.RespondWithError(c, http.StatusServiceUnavailable, "AI summary unavailable", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"summary": summary})
}
