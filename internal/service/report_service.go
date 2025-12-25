package service

import (
	"fmt"
	"strings"
	"time"

	"saas-backend/internal/models"
	"saas-backend/internal/repository"

	"github.com/google/uuid"
)

type ReportService struct {
	taskRepo      *repository.TaskRepository
	issueRepo     *repository.IssueRepository
	auditLogRepo  *repository.AuditLogRepository
	geminiService *GeminiService
}

func NewReportService(
	taskRepo *repository.TaskRepository,
	issueRepo *repository.IssueRepository,
	auditLogRepo *repository.AuditLogRepository,
	geminiService *GeminiService,
) *ReportService {
	return &ReportService{
		taskRepo:      taskRepo,
		issueRepo:     issueRepo,
		auditLogRepo:  auditLogRepo,
		geminiService: geminiService,
	}
}

func (s *ReportService) GenerateWeeklySummary(orgID uuid.UUID, generatedBy uuid.UUID) (string, error) {
	if s.geminiService == nil {
		return "", fmt.Errorf("Gemini service not configured")
	}

	now := time.Now()
	tasks, err := s.taskRepo.List(orgID, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}
	issues, err := s.issueRepo.List(orgID, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to list issues: %w", err)
	}

	delayed := make([]models.Task, 0)
	blockedCount := 0
	for _, t := range tasks {
		if t.Status == "blocked" {
			blockedCount++
		}
		if t.DueDate != nil && t.DueDate.Before(now) && t.Status != "approved" {
			delayed = append(delayed, t)
		}
	}

	highRisk := make([]models.Issue, 0)
	for _, i := range issues {
		sev := strings.ToLower(strings.TrimSpace(i.Severity))
		st := strings.ToLower(strings.TrimSpace(i.Status))
		if (sev == "high" || sev == "critical") && (st == "open" || st == "in_progress") {
			highRisk = append(highRisk, i)
		}
	}

	maxTasks := 20
	if len(delayed) < maxTasks {
		maxTasks = len(delayed)
	}
	delayedLines := make([]string, 0, maxTasks)
	for idx := 0; idx < maxTasks; idx++ {
		t := delayed[idx]
		assignee := "unassigned"
		if t.AssignedToName != nil && strings.TrimSpace(*t.AssignedToName) != "" {
			assignee = strings.TrimSpace(*t.AssignedToName)
		}
		due := "none"
		if t.DueDate != nil {
			due = t.DueDate.Format(time.RFC3339)
		}
		delayedLines = append(delayedLines, fmt.Sprintf("%d. [%s/%s] %s (assignee: %s, due: %s)", idx+1, t.Status, t.Priority, t.Title, assignee, due))
	}

	maxIssues := 20
	if len(highRisk) < maxIssues {
		maxIssues = len(highRisk)
	}
	highRiskLines := make([]string, 0, maxIssues)
	for idx := 0; idx < maxIssues; idx++ {
		i := highRisk[idx]
		assignee := "unassigned"
		if i.AssignedToName != nil && strings.TrimSpace(*i.AssignedToName) != "" {
			assignee = strings.TrimSpace(*i.AssignedToName)
		}
		highRiskLines = append(highRiskLines, fmt.Sprintf("%d. [%s/%s] %s (assignee: %s)", idx+1, i.Status, i.Severity, i.Title, assignee))
	}

	prompt := fmt.Sprintf(`You are an operations assistant helping a manager produce a weekly summary.

Write a concise report in Markdown format using only headings (##, ###) and numbered lists (1., 2., 3., ...).
Do not use asterisks, dashes, or bullet points anywhere.
Use minimal punctuation. Avoid exclamation marks, question marks, semicolons, ellipses, emojis, and decorative symbols.
Use only the provided data. Do not invent tasks, issues, owners, or numbers.

Required sections
## Weekly Summary
2 to 4 short sentences

## Delayed Tasks
Numbered list up to 10 items

## High Risk Issues
Numbered list up to 10 items

## Bottlenecks
Numbered list up to 5 items based on the counts

## Recommended Actions
Numbered list up to 5 items

Data
Delayed tasks count: %d
High risk issues count: %d
Blocked tasks count: %d

Delayed tasks list
%s

High risk issues list
%s
`, len(delayed), len(highRisk), blockedCount, strings.Join(delayedLines, "\n"), strings.Join(highRiskLines, "\n"))

	summary, err := s.geminiService.GenerateText(prompt)
	if err != nil {
		return "", err
	}
	summary = sanitizeAIReport(summary)

	// Audit log
	details := map[string]interface{}{
		"delayed_tasks":    len(delayed),
		"high_risk_issues": len(highRisk),
		"blocked_tasks":    blockedCount,
		"summary_preview":  preview(summary, 800),
		"generated_at":     now.Format(time.RFC3339),
		"report_type":      "weekly_summary",
	}
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &generatedBy,
		Action:     "generate",
		EntityType: "report",
		Details:    details,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return summary, nil
}

func preview(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}
