package handler

import (
	"net/http"
	"strconv"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"
	"saas-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	documentService *service.DocumentService
}

func NewDocumentHandler(documentService *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "missing file", err.Error())
		return
	}

	titleStr := c.PostForm("title")
	var title *string
	if titleStr != "" {
		title = &titleStr
	}

	// Optional task_id association
	var taskID *uuid.UUID
	if taskIDStr := c.PostForm("task_id"); taskIDStr != "" {
		parsed, err := uuid.Parse(taskIDStr)
		if err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "invalid task_id", err.Error())
			return
		}
		taskID = &parsed
	}

	doc, err := h.documentService.Upload(c.Request.Context(), orgID, userID, taskID, file, title)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "upload failed", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	docs, err := h.documentService.List(c.Request.Context(), orgID, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list documents", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, docs)
}

func (h *DocumentHandler) ListByTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	taskID, ok := utils.ParseUUID(c, "id", "task_id")
	if !ok {
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	docs, err := h.documentService.ListByTask(c.Request.Context(), orgID, taskID, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list documents", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, docs)
}

func (h *DocumentHandler) ListPending(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	docs, err := h.documentService.ListPending(c.Request.Context(), orgID, limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list documents", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, docs)
}

func (h *DocumentHandler) GetByID(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	docID, ok := utils.ParseUUID(c, "id", "document id")
	if !ok {
		return
	}

	doc, err := h.documentService.GetByID(c.Request.Context(), orgID, docID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get document", err.Error())
		return
	}
	if doc == nil {
		utils.RespondWithError(c, http.StatusNotFound, "not found", "document not found")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, doc)
}

func (h *DocumentHandler) UpdateStatus(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	docID, ok := utils.ParseUUID(c, "id", "document id")
	if !ok {
		return
	}

	var req models.UpdateDocumentStatusRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	if err := h.documentService.UpdateStatus(c.Request.Context(), orgID, docID, userID, req.Status, req.Notes); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "update failed", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "status updated"})
}

func (h *DocumentHandler) GenerateSummary(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	docID, ok := utils.ParseUUID(c, "id", "document id")
	if !ok {
		return
	}

	resp, err := h.documentService.GenerateSummary(c.Request.Context(), orgID, docID)
	if err != nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "summary generation failed", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, resp)
}

func (h *DocumentHandler) Verify(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	docID, ok := utils.ParseUUID(c, "id", "document id")
	if !ok {
		return
	}

	var req models.VerifyDocumentRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	resp, err := h.documentService.Verify(c.Request.Context(), orgID, docID, req.Question)
	if err != nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "verification unavailable", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, resp)
}
