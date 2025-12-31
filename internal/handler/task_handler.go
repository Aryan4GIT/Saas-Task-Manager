package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"
	"saas-backend/internal/utils"

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
	if !utils.BindJSON(c, &req) {
		return
	}

	task, err := h.taskService.CreateTaskForRole(orgID, userID, role, &req)
	if err != nil {
		utils.HandlePermissionError(c, err, "failed to create task")
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	task, err := h.taskService.GetTaskForRole(orgID, taskID, userID, role)
	if err != nil {
		utils.HandlePermissionError(c, err, "task not found")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, task)
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	status := c.Query("status")
	priority := c.Query("priority")

	tasks, err := h.taskService.ListTasksForRole(orgID, userID, role, status, priority)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list tasks", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, tasks)
}

func (h *TaskHandler) ListMyTasks(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	tasks, err := h.taskService.ListMyTasks(orgID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list tasks", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	var req models.UpdateTaskRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	task, err := h.taskService.UpdateTaskForRole(orgID, taskID, userID, role, &req)
	if err != nil {
		utils.HandlePermissionError(c, err, "failed to update task")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	if err := h.taskService.DeleteTaskForRole(orgID, taskID, userID, role); err != nil {
		utils.HandlePermissionError(c, err, "failed to delete task")
		return
	}

	utils.RespondWithMessage(c, http.StatusOK, "task deleted successfully")
}

// MarkDone - Member marks task as done
func (h *TaskHandler) MarkDone(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	// Check if there's a file upload
	file, err := c.FormFile("document")
	if err == nil && file != nil {
		// Handle file upload
		uploadDir := "cmd/server/uploads/tasks"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to create upload directory", err.Error())
			return
		}

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s_%s%s", taskID.String(), uuid.New().String()[:8], ext)
		filepath := filepath.Join(uploadDir, filename)

		// Save file
		if err := c.SaveUploadedFile(file, filepath); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to save file", err.Error())
			return
		}

		// Extract text content from file for AI processing
		content := ""
		if ext == ".txt" || ext == ".md" || ext == ".log" {
			data, err := os.ReadFile(filepath)
			if err == nil {
				content = string(data)
			}
		} else if ext == ".pdf" {
			// Extract text from PDF
			pdfContent, err := utils.ExtractTextFromPDF(filepath)
			if err == nil {
				content = pdfContent
			} else {
				fmt.Printf("Failed to extract PDF text: %v\n", err)
			}
		}

		// Limit content size for AI processing
		if len(content) > 50000 {
			content = content[:50000]
		}

		// Mark done with document
		task, err := h.taskService.MarkDoneWithDocument(orgID, taskID, userID, file.Filename, filepath, content)
		if err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "failed to mark task as done", err.Error())
			return
		}

		utils.RespondWithSuccess(c, http.StatusOK, task)
		return
	}

	// No file upload - regular mark done
	task, err := h.taskService.MarkDone(orgID, taskID, userID)
	if err != nil {
		status := http.StatusBadRequest
		errMsg := "failed to mark task as done"
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
			errMsg = "insufficient permissions"
		}
		if err.Error() == "task not found" {
			status = http.StatusNotFound
			errMsg = "task not found"
		}
		utils.RespondWithError(c, status, errMsg, err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, task)
}

// VerifyTask - Manager verifies a completed task
func (h *TaskHandler) VerifyTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	// Only managers and admins can verify
	if role != "manager" && role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "insufficient permissions", "only managers and admins can verify tasks")
		return
	}

	task, err := h.taskService.VerifyTask(orgID, taskID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to verify task", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, task)
}

// ApproveTask - Admin approves a verified task
func (h *TaskHandler) ApproveTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	// Only admins can approve
	if role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "insufficient permissions", "only admins can approve tasks")
		return
	}

	task, err := h.taskService.ApproveTask(orgID, taskID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to approve task", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, task)
}

// RejectTask - Manager/Admin rejects a task back to in_progress
func (h *TaskHandler) RejectTask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetRole(c)
	taskID, ok := utils.ParseUUID(c, "id", "task ID")
	if !ok {
		return
	}

	// Only managers and admins can reject
	if role != "manager" && role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "insufficient permissions", "only managers and admins can reject tasks")
		return
	}

	task, err := h.taskService.RejectTask(orgID, taskID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to reject task", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, task)
}

// AdminAIReport - Admin generates an AI summary report of organization tasks
func (h *TaskHandler) AdminAIReport(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	role, _ := middleware.GetRole(c)

	if role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "insufficient permissions", "only admins can generate AI task reports")
		return
	}

	report, err := h.taskService.GenerateAdminTaskReport(orgID)
	if err != nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "AI report unavailable", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{
		"report": report,
	})
}
