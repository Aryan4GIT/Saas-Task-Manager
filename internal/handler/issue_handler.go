package handler

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	issue, err := h.issueService.CreateIssue(orgID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "failed to create issue",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, issue)
}

func (h *IssueHandler) GetIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid issue ID",
			Message: err.Error(),
		})
		return
	}

	issue, err := h.issueService.GetIssueForRole(orgID, issueID, userID, role)
	if err != nil {
		status := http.StatusNotFound
		errMsg := "issue not found"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, issue)
}

func (h *IssueHandler) ListIssues(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	status := c.Query("status")
	severity := c.Query("severity")

	issues, err := h.issueService.ListIssuesForRole(orgID, userID, role, status, severity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to list issues",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, issues)
}

func (h *IssueHandler) UpdateIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid issue ID",
			Message: err.Error(),
		})
		return
	}

	var req models.UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	issue, err := h.issueService.UpdateIssueForRole(orgID, issueID, userID, role, &req)
	if err != nil {
		status := http.StatusBadRequest
		errMsg := "failed to update issue"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, issue)
}

func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid issue ID",
			Message: err.Error(),
		})
		return
	}

	if err := h.issueService.DeleteIssueForRole(orgID, issueID, userID, role); err != nil {
		status := http.StatusBadRequest
		errMsg := "failed to delete issue"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "issue deleted successfully",
	})
}
