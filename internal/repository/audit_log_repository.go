package repository

import (
	"database/sql"
	"encoding/json"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO audit_logs (id, org_id, user_id, action, entity_type, entity_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`
	return r.db.QueryRow(
		query,
		log.ID,
		log.OrgID,
		log.UserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		detailsJSON,
		log.IPAddress,
	).Scan(&log.CreatedAt)
}

func (r *AuditLogRepository) List(orgID uuid.UUID, limit int) ([]models.AuditLog, error) {
	query := `
		SELECT id, org_id, user_id, action, entity_type, entity_id, details, ip_address, created_at
		FROM audit_logs
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []models.AuditLog{}
	for rows.Next() {
		var log models.AuditLog
		var detailsJSON []byte
		err := rows.Scan(
			&log.ID,
			&log.OrgID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&detailsJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				return nil, err
			}
		}

		logs = append(logs, log)
	}
	return logs, rows.Err()
}
