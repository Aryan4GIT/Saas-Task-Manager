package repository

import (
	"database/sql"
	"time"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, org_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`
	return r.db.QueryRow(
		query,
		token.ID,
		token.UserID,
		token.OrgID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)
}

func (r *RefreshTokenRepository) GetByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, org_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > $2
	`
	token := &models.RefreshToken{}
	err := r.db.QueryRow(query, tokenHash, time.Now()).Scan(
		&token.ID,
		&token.UserID,
		&token.OrgID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return token, err
}

func (r *RefreshTokenRepository) RevokeToken(tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE token_hash = $2`
	_, err := r.db.Exec(query, time.Now(), tokenHash)
	return err
}

func (r *RefreshTokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}
