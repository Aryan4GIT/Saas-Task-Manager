package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"saas-backend/internal/ai"
	"saas-backend/internal/models"
	"saas-backend/internal/rag"
	"saas-backend/internal/repository"

	"github.com/google/uuid"
)

type TaskService struct {
	taskRepo      *repository.TaskRepository
	auditLogRepo  *repository.AuditLogRepository
	geminiService *GeminiService
	langChainSvc  *ai.LangChainService
	ragIndexer    *rag.Indexer
}

func NewTaskService(taskRepo *repository.TaskRepository, auditLogRepo *repository.AuditLogRepository, geminiService *GeminiService, langChainSvc *ai.LangChainService, ragIndexer *rag.Indexer) *TaskService {
	return &TaskService{
		taskRepo:      taskRepo,
		auditLogRepo:  auditLogRepo,
		geminiService: geminiService,
		langChainSvc:  langChainSvc,
		ragIndexer:    ragIndexer,
	}
}

func (s *TaskService) GenerateAdminTaskReport(orgID uuid.UUID) (string, error) {
	if s.langChainSvc == nil && s.geminiService == nil {
		return "", fmt.Errorf("AI service not configured")
	}

	tasks, err := s.taskRepo.List(orgID, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}

	statusCounts := map[string]int{}
	priorityCounts := map[string]int{}
	now := time.Now()
	overdueCount := 0
	dueSoonCount := 0

	for _, t := range tasks {
		statusCounts[t.Status]++
		priorityCounts[t.Priority]++

		if t.DueDate != nil {
			if t.DueDate.Before(now) && t.Status != "approved" {
				overdueCount++
			} else if t.DueDate.After(now) && t.DueDate.Before(now.Add(48*time.Hour)) {
				dueSoonCount++
			}
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	maxTasks := 30
	if len(tasks) < maxTasks {
		maxTasks = len(tasks)
	}

	lines := make([]string, 0, maxTasks)
	for i := 0; i < maxTasks; i++ {
		t := tasks[i]
		assignee := "unassigned"
		if t.AssignedToName != nil && strings.TrimSpace(*t.AssignedToName) != "" {
			assignee = strings.TrimSpace(*t.AssignedToName)
		}
		due := "none"
		if t.DueDate != nil {
			due = t.DueDate.Format(time.RFC3339)
		}
		lines = append(lines, fmt.Sprintf("%d. [%s/%s] %s (assignee: %s, due: %s)", i+1, t.Status, t.Priority, t.Title, assignee, due))
	}

	taskData := fmt.Sprintf(`Counts by status: %v
Counts by priority: %v
Overdue (non-approved): %d
Due within 48h: %d
Most recently updated tasks:
%s`, statusCounts, priorityCounts, overdueCount, dueSoonCount, strings.Join(lines, "\n"))

	var report string
	if s.langChainSvc != nil {
		prompt := ai.AdminReportPrompt(taskData)
		report, err = s.langChainSvc.GenerateText(context.Background(), prompt)
	} else {
		// Fallback to old Gemini service
		prompt := fmt.Sprintf(`You are an operations assistant helping an admin summarize the organization's task workload.

Write a concise admin report in Markdown format using only headings (##, ###) and numbered lists (1., 2., 3., ...).
Do not use asterisks, dashes, or bullet points anywhere.
Use minimal punctuation. Avoid exclamation marks, question marks, semicolons, ellipses, emojis, and decorative symbols.
Prefer short, clear sentences.

Sections:
## Overall Status
Summarize the current state of tasks in 2-4 sentences.

## Key Risks
List up to 5 key risks as a numbered list.

## Recommended Next Actions
List up to 5 next actions as a numbered list.

## Task Stats
Show status and priority counts, overdue and due-soon counts, as a short Markdown table or numbered list.

## Most Recently Updated Tasks
List up to %d tasks as a numbered list, using this format for each:
    [status/priority] title (assignee: name, due: date)

Data:
%s
`, maxTasks, taskData)
		report, err = s.geminiService.GenerateText(prompt)
	}

	if err != nil {
		return "", err
	}
	return sanitizeAIReport(report), nil
}

func (s *TaskService) CreateTask(orgID, createdBy uuid.UUID, req *models.CreateTaskRequest) (*models.Task, error) {
	task := &models.Task{
		ID:          uuid.New(),
		OrgID:       orgID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "todo",
		Priority:    req.Priority,
		CreatedBy:   createdBy,
	}

	if task.Priority == "" {
		task.Priority = "medium"
	}

	if req.AssignedTo != nil && *req.AssignedTo != "" {
		assignedID, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			return nil, fmt.Errorf("invalid assigned_to UUID: %w", err)
		}
		task.AssignedTo = &assignedID
	}

	if req.DueDate != nil && *req.DueDate != "" {
		dueDate, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format: %w", err)
		}
		task.DueDate = &dueDate
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Index task for RAG
	if s.ragIndexer != nil {
		s.ragIndexer.IndexTask(context.Background(), task.OrgID, task.ID, task.Title, task.Description)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &createdBy,
		Action:     "create",
		EntityType: "task",
		EntityID:   &task.ID,
		Details: map[string]interface{}{
			"title": task.Title,
		},
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}

func (s *TaskService) CreateTaskForRole(orgID, createdBy uuid.UUID, role string, req *models.CreateTaskRequest) (*models.Task, error) {
	if role != "admin" && role != "manager" {
		return nil, fmt.Errorf("insufficient permissions")
	}
	return s.CreateTask(orgID, createdBy, req)
}

func (s *TaskService) GetTask(orgID, taskID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}
	return task, nil
}

func (s *TaskService) GetTaskForRole(orgID, taskID, userID uuid.UUID, role string) (*models.Task, error) {
	task, err := s.GetTask(orgID, taskID)
	if err != nil {
		return nil, err
	}

	if role == "member" {
		if task.AssignedTo == nil || *task.AssignedTo != userID {
			return nil, fmt.Errorf("insufficient permissions")
		}
	}

	return task, nil
}

func (s *TaskService) ListTasks(orgID uuid.UUID, status string, priority string) ([]models.Task, error) {
	tasks, err := s.taskRepo.List(orgID, status, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

func (s *TaskService) ListTasksForRole(orgID, userID uuid.UUID, role string, status string, priority string) ([]models.Task, error) {
	if role != "member" {
		return s.ListTasks(orgID, status, priority)
	}

	assigned, err := s.taskRepo.ListByAssignee(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	filtered := make([]models.Task, 0, len(assigned))
	for _, t := range assigned {
		if status != "" {
			if status == "completed" {
				if t.Status != "done" && t.Status != "verified" && t.Status != "approved" {
					continue
				}
			} else if t.Status != status {
				continue
			}
		}
		if priority != "" && t.Priority != priority {
			continue
		}
		filtered = append(filtered, t)
	}

	return filtered, nil
}

func (s *TaskService) ListMyTasks(orgID, userID uuid.UUID) ([]models.Task, error) {
	tasks, err := s.taskRepo.ListByAssignee(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

func (s *TaskService) UpdateTask(orgID, taskID, userID uuid.UUID, req *models.UpdateTaskRequest) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Update fields if provided
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			task.AssignedTo = nil
		} else {
			assignedID, err := uuid.Parse(*req.AssignedTo)
			if err != nil {
				return nil, fmt.Errorf("invalid assigned_to UUID: %w", err)
			}
			task.AssignedTo = &assignedID
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			task.DueDate = nil
		} else {
			dueDate, err := time.Parse(time.RFC3339, *req.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due_date format: %w", err)
			}
			task.DueDate = &dueDate
		}
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Re-index task for RAG
	if s.ragIndexer != nil {
		s.ragIndexer.IndexTask(context.Background(), task.OrgID, task.ID, task.Title, task.Description)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "update",
		EntityType: "task",
		EntityID:   &task.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}

func (s *TaskService) UpdateTaskForRole(orgID, taskID, userID uuid.UUID, role string, req *models.UpdateTaskRequest) (*models.Task, error) {
	task, err := s.GetTask(orgID, taskID)
	if err != nil {
		return nil, err
	}

	if role == "member" {
		if task.AssignedTo == nil || *task.AssignedTo != userID {
			return nil, fmt.Errorf("insufficient permissions")
		}
		// Members can only update status and description (used as comments/notes).
		if req.Title != nil || req.Priority != nil || req.AssignedTo != nil || req.DueDate != nil {
			return nil, fmt.Errorf("insufficient permissions")
		}
		return s.UpdateTask(orgID, taskID, userID, req)
	}

	// Managers/Admins can update any fields.
	if role != "admin" && role != "manager" {
		return nil, fmt.Errorf("insufficient permissions")
	}
	return s.UpdateTask(orgID, taskID, userID, req)
}

func (s *TaskService) DeleteTask(orgID, taskID, userID uuid.UUID) error {
	if err := s.taskRepo.Delete(orgID, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "delete",
		EntityType: "task",
		EntityID:   &taskID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	// Delete from RAG index
	if s.ragIndexer != nil {
		s.ragIndexer.DeleteTask(context.Background(), orgID, taskID)
	}

	return nil
}

func (s *TaskService) DeleteTaskForRole(orgID, taskID, userID uuid.UUID, role string) error {
	if role != "admin" && role != "manager" {
		return fmt.Errorf("insufficient permissions")
	}
	return s.DeleteTask(orgID, taskID, userID)
}

// MarkDone - Member marks task as done
func (s *TaskService) MarkDone(orgID, taskID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}
	if task.AssignedTo == nil || *task.AssignedTo != userID {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Only allow marking as done if in_progress or todo
	if task.Status != "todo" && task.Status != "in_progress" {
		return nil, fmt.Errorf("task must be in todo or in_progress status to mark as done")
	}

	task.Status = "done"
	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "mark_done",
		EntityType: "task",
		EntityID:   &task.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}

// MarkDoneWithDocument - Member marks task as done with document upload
func (s *TaskService) MarkDoneWithDocument(orgID, taskID, userID uuid.UUID, filename, filepath, content string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}
	if task.AssignedTo == nil || *task.AssignedTo != userID {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Only allow marking as done if in_progress or todo
	if task.Status != "todo" && task.Status != "in_progress" {
		return nil, fmt.Errorf("task must be in todo or in_progress status to mark as done")
	}

	// Update task immediately with document info
	task.Status = "done"
	task.DocumentFilename = &filename
	task.DocumentPath = &filepath
	
	// Set initial processing message
	if content != "" {
		processingMsg := "⏳ AI summary is being generated..."
		task.DocumentSummary = &processingMsg
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Generate AI summary asynchronously (don't block HTTP response)
	if content != "" && (s.langChainSvc != nil || s.geminiService != nil) {
		// Capture variables for goroutine
		capturedOrgID := orgID
		capturedTaskID := taskID
		capturedFilename := filename
		capturedContent := content
		capturedTitle := task.Title

		go func() {
			// Recover from panics in goroutine
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic in AI summary generation for task %s: %v\n", capturedTaskID, r)
				}
			}()

			ctx := context.Background()
			var summary string

			fmt.Printf("Starting AI summary generation for task %s (content length: %d bytes)\n", capturedTaskID, len(capturedContent))

			// Try LangChain first
			if s.langChainSvc != nil {
				prompt := ai.TaskSummaryPrompt(capturedTitle, capturedContent)
				aiSummary, err := s.langChainSvc.GenerateText(ctx, prompt)
				if err == nil {
					summary = strings.TrimSpace(aiSummary)
					summary = strings.TrimPrefix(summary, "Summary:")
					summary = strings.TrimPrefix(summary, "SUMMARY:")
					summary = strings.TrimPrefix(summary, "RESPONSE:")
					summary = strings.TrimSpace(summary)
					fmt.Printf("✓ LangChain generated summary for task %s\n", capturedTaskID)
				} else {
					fmt.Printf("LangChain AI summary failed for task %s: %v\n", capturedTaskID, err)
				}
			}

			// Fallback to Gemini service
			if summary == "" && s.geminiService != nil {
				prompt := fmt.Sprintf(`You are an assistant helping managers review task completion documents. 
Analyze the following document submitted for task "%s" and provide a concise summary (3-5 sentences) covering:
1. Main deliverables or outcomes
2. Key findings or results
3. Any notable points for review

Document content:
%s

Provide ONLY the summary text, no additional formatting or preamble.`, capturedTitle, capturedContent)

				aiSummary, err := s.geminiService.GenerateText(prompt)
				if err == nil {
					summary = strings.TrimSpace(aiSummary)
					summary = strings.TrimPrefix(summary, "Summary:")
					summary = strings.TrimPrefix(summary, "SUMMARY:")
					summary = strings.TrimSpace(summary)
					fmt.Printf("✓ Gemini generated summary for task %s\n", capturedTaskID)
				} else {
					fmt.Printf("Gemini AI summary failed for task %s: %v\n", capturedTaskID, err)
				}
			}

			// Update task with generated summary
			if summary != "" {
				updatedTask, err := s.taskRepo.GetByID(capturedOrgID, capturedTaskID)
				if err != nil {
					fmt.Printf("Failed to fetch task %s for summary update: %v\n", capturedTaskID, err)
					return
				}
				if updatedTask != nil {
					updatedTask.DocumentSummary = &summary
					if err := s.taskRepo.Update(updatedTask); err != nil {
						fmt.Printf("Failed to update task %s with AI summary: %v\n", capturedTaskID, err)
					} else {
						fmt.Printf("✓ Successfully updated task %s with AI summary\n", capturedTaskID)
					}
				}
			} else {
				// Failed to generate - update with fallback message
				fmt.Printf("No summary generated for task %s, using fallback\n", capturedTaskID)
				updatedTask, err := s.taskRepo.GetByID(capturedOrgID, capturedTaskID)
				if err == nil && updatedTask != nil {
					fallback := fmt.Sprintf("Document uploaded: %s", capturedFilename)
					updatedTask.DocumentSummary = &fallback
					s.taskRepo.Update(updatedTask)
				}
			}
		}()
	}

	// Index document in RAG if available (async, don't block on errors)
	if s.ragIndexer != nil && content != "" {
		go func() {
			ctx := context.Background()
			s.ragIndexer.IndexTaskDocument(ctx, orgID, taskID, filename, content)
		}()
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "mark_done_with_document",
		EntityType: "task",
		EntityID:   &task.ID,
		Details: map[string]interface{}{
			"filename": filename,
		},
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}

// VerifyTask - Manager verifies a completed task
func (s *TaskService) VerifyTask(orgID, taskID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Only allow verifying if status is "done"
	if task.Status != "done" {
		return nil, fmt.Errorf("task must be marked as done before verification")
	}

	now := time.Now()
	task.Status = "verified"
	task.VerifiedBy = &userID
	task.VerifiedAt = &now

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "verify",
		EntityType: "task",
		EntityID:   &task.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}

// ApproveTask - Admin approves a verified task
func (s *TaskService) ApproveTask(orgID, taskID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Only allow approving if status is "verified"
	if task.Status != "verified" {
		return nil, fmt.Errorf("task must be verified before approval")
	}

	now := time.Now()
	task.Status = "approved"
	task.ApprovedBy = &userID
	task.ApprovedAt = &now

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "approve",
		EntityType: "task",
		EntityID:   &task.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}

// RejectTask - Manager/Admin rejects a task back to in_progress
func (s *TaskService) RejectTask(orgID, taskID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Only allow rejecting if status is "done" or "verified"
	if task.Status != "done" && task.Status != "verified" {
		return nil, fmt.Errorf("can only reject tasks that are done or verified")
	}

	task.Status = "in_progress"
	task.VerifiedBy = nil
	task.VerifiedAt = nil
	task.ApprovedBy = nil
	task.ApprovedAt = nil

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "reject",
		EntityType: "task",
		EntityID:   &task.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return task, nil
}
