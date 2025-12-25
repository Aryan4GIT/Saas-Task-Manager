package repository

import (
	"database/sql"
	"fmt"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		user.ID,
		user.OrgID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByEmail(orgID uuid.UUID, email string) (*models.User, error) {
	query := `
		SELECT id, org_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users
		WHERE org_id = $1 AND email = $2
	`
	user := &models.User{}
	err := r.db.QueryRow(query, orgID, email).Scan(
		&user.ID,
		&user.OrgID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) GetByEmailGlobal(email string) (*models.User, error) {
	query := `
		SELECT id, org_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
		LIMIT 1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.OrgID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) GetByID(orgID, userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, org_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users
		WHERE org_id = $1 AND id = $2
	`
	user := &models.User{}
	err := r.db.QueryRow(query, orgID, userID).Scan(
		&user.ID,
		&user.OrgID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) List(orgID uuid.UUID) ([]models.User, error) {
	query := `
		SELECT id, org_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users
		WHERE org_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.OrgID,
			&user.Email,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, role = $4, is_active = $5
		WHERE org_id = $6 AND id = $7
	`
	result, err := r.db.Exec(query, user.Email, user.FirstName, user.LastName, user.Role, user.IsActive, user.OrgID, user.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *UserRepository) Delete(orgID, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE org_id = $1 AND id = $2`
	result, err := r.db.Exec(query, orgID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
