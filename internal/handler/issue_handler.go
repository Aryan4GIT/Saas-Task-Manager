package handler

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"
	"saas-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type IssueHandler struct {
	issueService *service.IssueService
}

func NewIssueHandler(issueService *service.IssueService) *IssueHandler {
	return &IssueHandler{
		issueService: issueService,
	}
}

func (h *IssueHandler) CreateIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	var req models.CreateIssueRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	issue, err := h.issueService.CreateIssue(orgID, userID, &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to create issue", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, issue)
}

func (h *IssueHandler) GetIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	issueID, ok := utils.ParseUUID(c, "id", "issue ID")
	if !ok {
		return
	}

	issue, err := h.issueService.GetIssueForRole(orgID, issueID, userID, role)
	if err != nil {
		utils.HandlePermissionError(c, err, "issue not found")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, issue)
}

func (h *IssueHandler) ListIssues(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	status := c.Query("status")
	severity := c.Query("severity")

	issues, err := h.issueService.ListIssuesForRole(orgID, userID, role, status, severity)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list issues", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, issues)
}

func (h *IssueHandler) UpdateIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	issueID, ok := utils.ParseUUID(c, "id", "issue ID")
	if !ok {
		return
	}

	var req models.UpdateIssueRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	issue, err := h.issueService.UpdateIssueForRole(orgID, issueID, userID, role, &req)
	if err != nil {
		utils.HandlePermissionError(c, err, "failed to update issue")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, issue)
}

func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	issueID, ok := utils.ParseUUID(c, "id", "issue ID")
	if !ok {
		return
	}

	if err := h.issueService.DeleteIssueForRole(orgID, issueID, userID, role); err != nil {
		utils.HandlePermissionError(c, err, "failed to delete issue")
		return
	}

	utils.RespondWithMessage(c, http.StatusOK, "issue deleted successfully")
}
