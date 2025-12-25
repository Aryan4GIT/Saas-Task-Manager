package handler

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	task, err := h.taskService.CreateTaskForRole(orgID, userID, role, &req)
	if err != nil {
		status := http.StatusBadRequest
		errMsg := "failed to create task"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	task, err := h.taskService.GetTaskForRole(orgID, taskID, userID, role)
	if err != nil {
		status := http.StatusNotFound
		errMsg := "task not found"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	status := c.Query("status")
	priority := c.Query("priority")

	tasks, err := h.taskService.ListTasksForRole(orgID, userID, role, status, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to list tasks",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) ListMyTasks(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	tasks, err := h.taskService.ListMyTasks(orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to list tasks",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	task, err := h.taskService.UpdateTaskForRole(orgID, taskID, userID, role, &req)
	if err != nil {
		status := http.StatusBadRequest
		errMsg := "failed to update task"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	if err := h.taskService.DeleteTaskForRole(orgID, taskID, userID, role); err != nil {
		status := http.StatusBadRequest
		errMsg := "failed to delete task"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "task deleted successfully",
	})
}

// MarkDone - Member marks task as done
func (h *TaskHandler) MarkDone(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	task, err := h.taskService.MarkDone(orgID, taskID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "failed to mark task as done",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// VerifyTask - Manager verifies a completed task
func (h *TaskHandler) VerifyTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	// Only managers and admins can verify
	if role != "manager" && role != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "insufficient permissions",
			Message: "only managers and admins can verify tasks",
		})
		return
	}

	task, err := h.taskService.VerifyTask(orgID, taskID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "failed to verify task",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ApproveTask - Admin approves a verified task
func (h *TaskHandler) ApproveTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	// Only admins can approve
	if role != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "insufficient permissions",
			Message: "only admins can approve tasks",
		})
		return
	}

	task, err := h.taskService.ApproveTask(orgID, taskID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "failed to approve task",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// RejectTask - Manager/Admin rejects a task back to in_progress
func (h *TaskHandler) RejectTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid task ID",
			Message: err.Error(),
		})
		return
	}

	// Only managers and admins can reject
	if role != "manager" && role != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "insufficient permissions",
			Message: "only managers and admins can reject tasks",
		})
		return
	}

	task, err := h.taskService.RejectTask(orgID, taskID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "failed to reject task",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// AdminAIReport - Admin generates an AI summary report of organization tasks
func (h *TaskHandler) AdminAIReport(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	role, _ := middleware.GetRole(c)

	if role != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "insufficient permissions",
			Message: "only admins can generate AI task reports",
		})
		return
	}

	report, err := h.taskService.GenerateAdminTaskReport(orgID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "AI report unavailable",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}
