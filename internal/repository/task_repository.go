package repository

import (
	"database/sql"
	"fmt"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *models.Task) error {
	query := `
		INSERT INTO tasks (id, org_id, title, description, status, priority, assigned_to, created_by, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		task.ID,
		task.OrgID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.AssignedTo,
		task.CreatedBy,
		task.DueDate,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
}

func (r *TaskRepository) GetByID(orgID, taskID uuid.UUID) (*models.Task, error) {
	query := `
		SELECT
			t.id, t.org_id, t.title, t.description, t.status, t.priority, t.assigned_to, t.created_by, t.due_date, t.created_at, t.updated_at,
			t.verified_by, t.verified_at, t.approved_by, t.approved_at,
			CASE
				WHEN au.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(au.first_name, ''), ' ', COALESCE(au.last_name, ''))
			END AS assigned_to_name,
			CASE
				WHEN cu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(cu.first_name, ''), ' ', COALESCE(cu.last_name, ''))
			END AS created_by_name,
			CASE
				WHEN vu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(vu.first_name, ''), ' ', COALESCE(vu.last_name, ''))
			END AS verified_by_name,
			CASE
				WHEN apu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(apu.first_name, ''), ' ', COALESCE(apu.last_name, ''))
			END AS approved_by_name
		FROM tasks t
		LEFT JOIN users au ON au.id = t.assigned_to
		LEFT JOIN users cu ON cu.id = t.created_by
		LEFT JOIN users vu ON vu.id = t.verified_by
		LEFT JOIN users apu ON apu.id = t.approved_by
		WHERE t.org_id = $1 AND t.id = $2
	`
	task := &models.Task{}
	err := r.db.QueryRow(query, orgID, taskID).Scan(
		&task.ID,
		&task.OrgID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssignedTo,
		&task.CreatedBy,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.VerifiedBy,
		&task.VerifiedAt,
		&task.ApprovedBy,
		&task.ApprovedAt,
		&task.AssignedToName,
		&task.CreatedByName,
		&task.VerifiedByName,
		&task.ApprovedByName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return task, err
}

func (r *TaskRepository) List(orgID uuid.UUID, status string, priority string) ([]models.Task, error) {
	base := `
		SELECT
			t.id, t.org_id, t.title, t.description, t.status, t.priority, t.assigned_to, t.created_by, t.due_date, t.created_at, t.updated_at,
			t.verified_by, t.verified_at, t.approved_by, t.approved_at,
			CASE
				WHEN au.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(au.first_name, ''), ' ', COALESCE(au.last_name, ''))
			END AS assigned_to_name,
			CASE
				WHEN cu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(cu.first_name, ''), ' ', COALESCE(cu.last_name, ''))
			END AS created_by_name,
			CASE
				WHEN vu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(vu.first_name, ''), ' ', COALESCE(vu.last_name, ''))
			END AS verified_by_name,
			CASE
				WHEN apu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(apu.first_name, ''), ' ', COALESCE(apu.last_name, ''))
			END AS approved_by_name
		FROM tasks t
		LEFT JOIN users au ON au.id = t.assigned_to
		LEFT JOIN users cu ON cu.id = t.created_by
		LEFT JOIN users vu ON vu.id = t.verified_by
		LEFT JOIN users apu ON apu.id = t.approved_by
		WHERE t.org_id = $1
	`

	args := []interface{}{orgID}
	argIdx := 2

	if status != "" {
		if status == "completed" {
			// Aggregate completed statuses
			base += " AND t.status IN ('done','verified','approved')"
		} else {
			base += fmt.Sprintf(" AND t.status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}
	}
	if priority != "" {
		base += fmt.Sprintf(" AND t.priority = $%d", argIdx)
		args = append(args, priority)
		argIdx++
	}
	base += " ORDER BY t.created_at DESC"

	rows, err := r.db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.OrgID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.AssignedTo,
			&task.CreatedBy,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.VerifiedBy,
			&task.VerifiedAt,
			&task.ApprovedBy,
			&task.ApprovedAt,
			&task.AssignedToName,
			&task.CreatedByName,
			&task.VerifiedByName,
			&task.ApprovedByName,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) Update(task *models.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4, assigned_to = $5, due_date = $6,
			verified_by = $7, verified_at = $8, approved_by = $9, approved_at = $10, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = $11 AND id = $12
	`
	result, err := r.db.Exec(
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.AssignedTo,
		task.DueDate,
		task.VerifiedBy,
		task.VerifiedAt,
		task.ApprovedBy,
		task.ApprovedAt,
		task.OrgID,
		task.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *TaskRepository) Delete(orgID, taskID uuid.UUID) error {
	query := `DELETE FROM tasks WHERE org_id = $1 AND id = $2`
	result, err := r.db.Exec(query, orgID, taskID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *TaskRepository) ListByAssignee(orgID, userID uuid.UUID) ([]models.Task, error) {
	query := `
		SELECT
			t.id, t.org_id, t.title, t.description, t.status, t.priority, t.assigned_to, t.created_by, t.due_date, t.created_at, t.updated_at,
			t.verified_by, t.verified_at, t.approved_by, t.approved_at,
			CASE
				WHEN au.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(au.first_name, ''), ' ', COALESCE(au.last_name, ''))
			END AS assigned_to_name,
			CASE
				WHEN cu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(cu.first_name, ''), ' ', COALESCE(cu.last_name, ''))
			END AS created_by_name,
			CASE
				WHEN vu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(vu.first_name, ''), ' ', COALESCE(vu.last_name, ''))
			END AS verified_by_name,
			CASE
				WHEN apu.id IS NULL THEN NULL
				ELSE CONCAT(COALESCE(apu.first_name, ''), ' ', COALESCE(apu.last_name, ''))
			END AS approved_by_name
		FROM tasks t
		LEFT JOIN users au ON au.id = t.assigned_to
		LEFT JOIN users cu ON cu.id = t.created_by
		LEFT JOIN users vu ON vu.id = t.verified_by
		LEFT JOIN users apu ON apu.id = t.approved_by
		WHERE t.org_id = $1 AND t.assigned_to = $2
		ORDER BY t.created_at DESC
	`
	rows, err := r.db.Query(query, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.OrgID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.AssignedTo,
			&task.CreatedBy,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.VerifiedBy,
			&task.VerifiedAt,
			&task.ApprovedBy,
			&task.ApprovedAt,
			&task.AssignedToName,
			&task.CreatedByName,
			&task.VerifiedByName,
			&task.ApprovedByName,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}
