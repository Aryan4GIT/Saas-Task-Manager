package repository

import (
	"database/sql"
	"fmt"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type IssueRepository struct {
	db *sql.DB
}

func NewIssueRepository(db *sql.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

func (r *IssueRepository) Create(issue *models.Issue) error {
	query := `
		INSERT INTO issues (id, org_id, title, description, severity, status, reported_by, assigned_to, ai_summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		issue.ID,
		issue.OrgID,
		issue.Title,
		issue.Description,
		issue.Severity,
		issue.Status,
		issue.ReportedBy,
		issue.AssignedTo,
		issue.AISummary,
	).Scan(&issue.CreatedAt, &issue.UpdatedAt)
}

func (r *IssueRepository) GetByID(orgID, issueID uuid.UUID) (*models.Issue, error) {
	query := `
		SELECT
			i.id, i.org_id, i.title, i.description, i.severity, i.status, i.reported_by, i.assigned_to, i.ai_summary, i.created_at, i.updated_at, i.resolved_at,
			CONCAT(COALESCE(ru.first_name, ''), ' ', COALESCE(ru.last_name, '')) AS reported_by_name,
			CASE
				WHEN au.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(au.first_name, ''), ' ', COALESCE(au.last_name, ''))
			END AS assigned_to_name
		FROM issues i
		LEFT JOIN users ru ON ru.id = i.reported_by
		LEFT JOIN users au ON au.id = i.assigned_to
		WHERE i.org_id = $1 AND i.id = $2
	`
	issue := &models.Issue{}
	err := r.db.QueryRow(query, orgID, issueID).Scan(
		&issue.ID,
		&issue.OrgID,
		&issue.Title,
		&issue.Description,
		&issue.Severity,
		&issue.Status,
		&issue.ReportedBy,
		&issue.AssignedTo,
		&issue.AISummary,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.ResolvedAt,
		&issue.ReportedByName,
		&issue.AssignedToName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return issue, err
}

func (r *IssueRepository) List(orgID uuid.UUID, status string, severity string) ([]models.Issue, error) {
	base := `
		SELECT
			i.id, i.org_id, i.title, i.description, i.severity, i.status, i.reported_by, i.assigned_to, i.ai_summary, i.created_at, i.updated_at, i.resolved_at,
			CONCAT(COALESCE(ru.first_name, ''), ' ', COALESCE(ru.last_name, '')) AS reported_by_name,
			CASE
				WHEN au.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(au.first_name, ''), ' ', COALESCE(au.last_name, ''))
			END AS assigned_to_name
		FROM issues i
		LEFT JOIN users ru ON ru.id = i.reported_by
		LEFT JOIN users au ON au.id = i.assigned_to
		WHERE i.org_id = $1
	`

	args := []interface{}{orgID}
	argIdx := 2
	if status != "" {
		base += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if severity != "" {
		base += fmt.Sprintf(" AND i.severity = $%d", argIdx)
		args = append(args, severity)
		argIdx++
	}
	base += " ORDER BY i.created_at DESC"

	rows, err := r.db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := []models.Issue{}
	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(
			&issue.ID,
			&issue.OrgID,
			&issue.Title,
			&issue.Description,
			&issue.Severity,
			&issue.Status,
			&issue.ReportedBy,
			&issue.AssignedTo,
			&issue.AISummary,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.ResolvedAt,
			&issue.ReportedByName,
			&issue.AssignedToName,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}

func (r *IssueRepository) ListForUser(orgID uuid.UUID, userID uuid.UUID, status string, severity string) ([]models.Issue, error) {
	base := `
		SELECT
			i.id, i.org_id, i.title, i.description, i.severity, i.status, i.reported_by, i.assigned_to, i.ai_summary, i.created_at, i.updated_at, i.resolved_at,
			CONCAT(COALESCE(ru.first_name, ''), ' ', COALESCE(ru.last_name, '')) AS reported_by_name,
			CASE
				WHEN au.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(au.first_name, ''), ' ', COALESCE(au.last_name, ''))
			END AS assigned_to_name
		FROM issues i
		LEFT JOIN users ru ON ru.id = i.reported_by
		LEFT JOIN users au ON au.id = i.assigned_to
		WHERE i.org_id = $1 AND (i.reported_by = $2 OR i.assigned_to = $2)
	`

	args := []interface{}{orgID, userID}
	argIdx := 3
	if status != "" {
		base += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if severity != "" {
		base += fmt.Sprintf(" AND i.severity = $%d", argIdx)
		args = append(args, severity)
		argIdx++
	}
	base += " ORDER BY i.created_at DESC"

	rows, err := r.db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := []models.Issue{}
	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(
			&issue.ID,
			&issue.OrgID,
			&issue.Title,
			&issue.Description,
			&issue.Severity,
			&issue.Status,
			&issue.ReportedBy,
			&issue.AssignedTo,
			&issue.AISummary,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.ResolvedAt,
			&issue.ReportedByName,
			&issue.AssignedToName,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}

func (r *IssueRepository) Update(issue *models.Issue) error {
	query := `
		UPDATE issues
		SET title = $1, description = $2, severity = $3, status = $4, assigned_to = $5, ai_summary = $6, resolved_at = $7
		WHERE org_id = $8 AND id = $9
	`
	result, err := r.db.Exec(
		query,
		issue.Title,
		issue.Description,
		issue.Severity,
		issue.Status,
		issue.AssignedTo,
		issue.AISummary,
		issue.ResolvedAt,
		issue.OrgID,
		issue.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("issue not found")
	}
	return nil
}

func (r *IssueRepository) Delete(orgID, issueID uuid.UUID) error {
	query := `DELETE FROM issues WHERE org_id = $1 AND id = $2`
	result, err := r.db.Exec(query, orgID, issueID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("issue not found")
	}
	return nil
}
