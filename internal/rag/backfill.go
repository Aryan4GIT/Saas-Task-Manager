package rag

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type BackfillService struct {
	db      *sql.DB
	service *Service
}

func NewBackfillService(db *sql.DB, service *Service) *BackfillService {
	if service == nil {
		return nil
	}
	return &BackfillService{
		db:      db,
		service: service,
	}
}

type BackfillResult struct {
	TasksIndexed  int `json:"tasks_indexed"`
	IssuesIndexed int `json:"issues_indexed"`
	Errors        int `json:"errors"`
}

func (b *BackfillService) BackfillOrganization(ctx context.Context, orgID uuid.UUID) (*BackfillResult, error) {
	if b == nil || b.service == nil {
		return nil, fmt.Errorf("backfill service not initialized")
	}

	result := &BackfillResult{}

	// Index all tasks for the organization
	tasksIndexed, taskErrors := b.indexTasks(ctx, orgID)
	result.TasksIndexed = tasksIndexed
	result.Errors += taskErrors

	// Index all issues for the organization
	issuesIndexed, issueErrors := b.indexIssues(ctx, orgID)
	result.IssuesIndexed = issuesIndexed
	result.Errors += issueErrors

	return result, nil
}

func (b *BackfillService) indexTasks(ctx context.Context, orgID uuid.UUID) (int, int) {
	query := `
		SELECT id, title, description 
		FROM tasks 
		WHERE org_id = $1
	`

	rows, err := b.db.QueryContext(ctx, query, orgID)
	if err != nil {
		log.Printf("Failed to query tasks for backfill: %v", err)
		return 0, 1
	}
	defer rows.Close()

	indexed := 0
	errors := 0

	for rows.Next() {
		var id uuid.UUID
		var title string
		var description string

		if err := rows.Scan(&id, &title, &description); err != nil {
			log.Printf("Failed to scan task: %v", err)
			errors++
			continue
		}

		content := fmt.Sprintf("Task: %s\n\n%s", title, description)
		err := b.service.IndexDocument(ctx, IndexRequest{
			OrgID:      orgID,
			SourceType: "task",
			SourceID:   id,
			Content:    content,
		})
		if err != nil {
			log.Printf("Failed to index task %s: %v", id, err)
			errors++
			continue
		}
		indexed++
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating tasks: %v", err)
		errors++
	}

	return indexed, errors
}

func (b *BackfillService) indexIssues(ctx context.Context, orgID uuid.UUID) (int, int) {
	query := `
		SELECT id, title, description 
		FROM issues 
		WHERE org_id = $1
	`

	rows, err := b.db.QueryContext(ctx, query, orgID)
	if err != nil {
		log.Printf("Failed to query issues for backfill: %v", err)
		return 0, 1
	}
	defer rows.Close()

	indexed := 0
	errors := 0

	for rows.Next() {
		var id uuid.UUID
		var title string
		var description string

		if err := rows.Scan(&id, &title, &description); err != nil {
			log.Printf("Failed to scan issue: %v", err)
			errors++
			continue
		}

		content := fmt.Sprintf("Issue: %s\n\n%s", title, description)
		err := b.service.IndexDocument(ctx, IndexRequest{
			OrgID:      orgID,
			SourceType: "issue",
			SourceID:   id,
			Content:    content,
		})
		if err != nil {
			log.Printf("Failed to index issue %s: %v", id, err)
			errors++
			continue
		}
		indexed++
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating issues: %v", err)
		errors++
	}

	return indexed, errors
}
