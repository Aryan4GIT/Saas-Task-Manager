package repository

import (
	"database/sql"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(org *models.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(query, org.ID, org.Name, org.Slug).Scan(&org.CreatedAt, &org.UpdatedAt)
}

func (r *OrganizationRepository) GetByID(orgID uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	org := &models.Organization{}
	err := r.db.QueryRow(query, orgID).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return org, err
}

func (r *OrganizationRepository) GetBySlug(slug string) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`
	org := &models.Organization{}
	err := r.db.QueryRow(query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return org, err
}
