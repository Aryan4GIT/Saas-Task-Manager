package service

import (
	"context"
	"fmt"
	"time"

	"saas-backend/internal/models"
	"saas-backend/internal/rag"
	"saas-backend/internal/repository"

	"github.com/google/uuid"
)

type IssueService struct {
	issueRepo     *repository.IssueRepository
	auditLogRepo  *repository.AuditLogRepository
	geminiService *GeminiService
	ragIndexer    *rag.Indexer
}

func NewIssueService(
	issueRepo *repository.IssueRepository,
	auditLogRepo *repository.AuditLogRepository,
	geminiService *GeminiService,
	ragIndexer *rag.Indexer,
) *IssueService {
	return &IssueService{
		issueRepo:     issueRepo,
		auditLogRepo:  auditLogRepo,
		geminiService: geminiService,
		ragIndexer:    ragIndexer,
	}
}

func (s *IssueService) CreateIssue(orgID, reportedBy uuid.UUID, req *models.CreateIssueRequest) (*models.Issue, error) {
	issue := &models.Issue{
		ID:          uuid.New(),
		OrgID:       orgID,
		Title:       req.Title,
		Description: req.Description,
		Severity:    req.Severity,
		Status:      "open",
		ReportedBy:  reportedBy,
	}

	if req.AssignedTo != nil && *req.AssignedTo != "" {
		assignedID, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			return nil, fmt.Errorf("invalid assigned_to UUID: %w", err)
		}
		issue.AssignedTo = &assignedID
	}

	// Generate AI summary if Gemini service is available
	if s.geminiService != nil {
		summary, err := s.geminiService.GenerateIssueSummary(req.Title, req.Description)
		if err == nil && summary != "" {
			issue.AISummary = &summary
		}
		// Continue even if AI summary fails
	}

	if err := s.issueRepo.Create(issue); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	// Index issue for RAG
	if s.ragIndexer != nil {
		s.ragIndexer.IndexIssue(context.Background(), issue.OrgID, issue.ID, issue.Title, issue.Description)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &reportedBy,
		Action:     "create",
		EntityType: "issue",
		EntityID:   &issue.ID,
		Details: map[string]interface{}{
			"title":    issue.Title,
			"severity": issue.Severity,
		},
	}
	_ = s.auditLogRepo.Create(auditLog)

	return issue, nil
}

func (s *IssueService) GetIssue(orgID, issueID uuid.UUID) (*models.Issue, error) {
	issue, err := s.issueRepo.GetByID(orgID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue not found")
	}
	return issue, nil
}

func (s *IssueService) GetIssueForRole(orgID, issueID, userID uuid.UUID, role string) (*models.Issue, error) {
	issue, err := s.GetIssue(orgID, issueID)
	if err != nil {
		return nil, err
	}

	if role == "member" {
		isReporter := issue.ReportedBy == userID
		isAssignee := issue.AssignedTo != nil && *issue.AssignedTo == userID
		if !isReporter && !isAssignee {
			return nil, fmt.Errorf("insufficient permissions")
		}
	}

	return issue, nil
}

func (s *IssueService) ListIssuesForRole(orgID, userID uuid.UUID, role string, status string, severity string) ([]models.Issue, error) {
	if role == "member" {
		issues, err := s.issueRepo.ListForUser(orgID, userID, status, severity)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}
		return issues, nil
	}

	issues, err := s.issueRepo.List(orgID, status, severity)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	return issues, nil
}

func (s *IssueService) UpdateIssueForRole(orgID, issueID, userID uuid.UUID, role string, req *models.UpdateIssueRequest) (*models.Issue, error) {
	issue, err := s.issueRepo.GetByID(orgID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue not found")
	}

	if role == "member" {
		// Members can only update their own reported issues and cannot change status or assignment.
		if issue.ReportedBy != userID {
			return nil, fmt.Errorf("insufficient permissions")
		}
		if req.Status != nil || req.AssignedTo != nil {
			return nil, fmt.Errorf("insufficient permissions")
		}
	}

	if role != "admin" && role != "manager" && role != "member" {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Update fields if provided
	if req.Title != nil {
		issue.Title = *req.Title
	}
	if req.Description != nil {
		issue.Description = *req.Description
	}
	if req.Severity != nil {
		issue.Severity = *req.Severity
	}
	if req.Status != nil {
		issue.Status = *req.Status
		// Set resolved_at when status changes to resolved or closed
		if (*req.Status == "resolved" || *req.Status == "closed") && issue.ResolvedAt == nil {
			now := time.Now()
			issue.ResolvedAt = &now
		}
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			issue.AssignedTo = nil
		} else {
			assignedID, err := uuid.Parse(*req.AssignedTo)
			if err != nil {
				return nil, fmt.Errorf("invalid assigned_to UUID: %w", err)
			}
			issue.AssignedTo = &assignedID
		}
	}

	if err := s.issueRepo.Update(issue); err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	// Re-index issue for RAG
	if s.ragIndexer != nil {
		s.ragIndexer.IndexIssue(context.Background(), issue.OrgID, issue.ID, issue.Title, issue.Description)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "update",
		EntityType: "issue",
		EntityID:   &issue.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return issue, nil
}

func (s *IssueService) DeleteIssueForRole(orgID, issueID, userID uuid.UUID, role string) error {
	if role != "admin" && role != "manager" {
		return fmt.Errorf("insufficient permissions")
	}
	if err := s.issueRepo.Delete(orgID, issueID); err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &userID,
		Action:     "delete",
		EntityType: "issue",
		EntityID:   &issueID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	// Delete from RAG index
	if s.ragIndexer != nil {
		s.ragIndexer.DeleteIssue(context.Background(), orgID, issueID)
	}

	return nil
}
