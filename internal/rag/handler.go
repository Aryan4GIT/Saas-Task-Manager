package rag

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service         *Service
	backfillService *BackfillService
}

func NewHandler(service *Service, backfillService *BackfillService) *Handler {
	return &Handler{
		service:         service,
		backfillService: backfillService,
	}
}

type QueryRequestDTO struct {
	Question string `json:"question" binding:"required"`
}

func (h *Handler) Query(c *gin.Context) {
	if h.service == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "RAG service not available", "")
		return
	}

	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req QueryRequestDTO
	if !utils.BindJSON(c, &req) {
		return
	}

	if req.Question == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "question is required", "")
		return
	}

	resp, err := h.service.Query(c.Request.Context(), QueryRequest{
		OrgID:    orgID,
		UserID:   userID,
		Role:     role,
		Question: req.Question,
	})
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to process query", "")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, resp)
}

func (h *Handler) Backfill(c *gin.Context) {
	if h.backfillService == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "backfill service not available", "")
		return
	}

	orgID, _ := middleware.GetOrgID(c)

	result, err := h.backfillService.BackfillOrganization(c.Request.Context(), orgID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to backfill documents", "")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, result)
}
