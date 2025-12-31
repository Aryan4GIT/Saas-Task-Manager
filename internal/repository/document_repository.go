package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (
			id, org_id, task_id, uploaded_by, title, filename, mime_type, file_size, sha256, storage_path, extracted_text, status
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING created_at, updated_at
	`
	doc.Status = "submitted"
	return r.db.QueryRowContext(
		ctx,
		query,
		doc.ID,
		doc.OrgID,
		doc.TaskID,
		doc.UploadedBy,
		doc.Title,
		doc.Filename,
		doc.MimeType,
		doc.FileSize,
		doc.SHA256,
		doc.StoragePath,
		doc.ExtractedText,
		doc.Status,
	).Scan(&doc.CreatedAt, &doc.UpdatedAt)
}

func (r *DocumentRepository) List(ctx context.Context, orgID uuid.UUID, limit int) ([]models.Document, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT 
			d.id, d.org_id, d.task_id, d.uploaded_by, d.title, d.filename, d.mime_type, 
			d.file_size, d.sha256, d.status, d.verified_by, d.verified_at, d.verification_notes,
			d.created_at, d.updated_at,
			CASE WHEN u.id IS NULL THEN NULL ELSE CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) END AS uploaded_by_name,
			CASE WHEN v.id IS NULL THEN NULL ELSE CONCAT(COALESCE(v.first_name, ''), ' ', COALESCE(v.last_name, '')) END AS verified_by_name
		FROM documents d
		LEFT JOIN users u ON u.id = d.uploaded_by
		LEFT JOIN users v ON v.id = d.verified_by
		WHERE d.org_id = $1
		ORDER BY d.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Document{}
	for rows.Next() {
		var d models.Document
		err := rows.Scan(
			&d.ID,
			&d.OrgID,
			&d.TaskID,
			&d.UploadedBy,
			&d.Title,
			&d.Filename,
			&d.MimeType,
			&d.FileSize,
			&d.SHA256,
			&d.Status,
			&d.VerifiedBy,
			&d.VerifiedAt,
			&d.VerificationNotes,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.UploadedByName,
			&d.VerifiedByName,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentRepository) ListByTask(ctx context.Context, orgID, taskID uuid.UUID, limit int) ([]models.Document, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT 
			d.id, d.org_id, d.task_id, d.uploaded_by, d.title, d.filename, d.mime_type, 
			d.file_size, d.sha256, d.status, d.verified_by, d.verified_at, d.verification_notes,
			d.created_at, d.updated_at,
			CASE WHEN u.id IS NULL THEN NULL ELSE CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) END AS uploaded_by_name,
			CASE WHEN v.id IS NULL THEN NULL ELSE CONCAT(COALESCE(v.first_name, ''), ' ', COALESCE(v.last_name, '')) END AS verified_by_name
		FROM documents d
		LEFT JOIN users u ON u.id = d.uploaded_by
		LEFT JOIN users v ON v.id = d.verified_by
		WHERE d.org_id = $1 AND d.task_id = $2
		ORDER BY d.created_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, taskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Document{}
	for rows.Next() {
		var d models.Document
		err := rows.Scan(
			&d.ID,
			&d.OrgID,
			&d.TaskID,
			&d.UploadedBy,
			&d.Title,
			&d.Filename,
			&d.MimeType,
			&d.FileSize,
			&d.SHA256,
			&d.Status,
			&d.VerifiedBy,
			&d.VerifiedAt,
			&d.VerificationNotes,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.UploadedByName,
			&d.VerifiedByName,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentRepository) GetByID(ctx context.Context, orgID, documentID uuid.UUID) (*models.Document, error) {
	query := `
		SELECT 
			d.id, d.org_id, d.task_id, d.uploaded_by, d.title, d.filename, d.mime_type,
			d.file_size, d.sha256, d.storage_path, d.extracted_text, d.status,
			d.verified_by, d.verified_at, d.verification_notes, d.created_at, d.updated_at,
			CASE WHEN u.id IS NULL THEN NULL ELSE CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) END AS uploaded_by_name,
			CASE WHEN v.id IS NULL THEN NULL ELSE CONCAT(COALESCE(v.first_name, ''), ' ', COALESCE(v.last_name, '')) END AS verified_by_name
		FROM documents d
		LEFT JOIN users u ON u.id = d.uploaded_by
		LEFT JOIN users v ON v.id = d.verified_by
		WHERE d.org_id = $1 AND d.id = $2
	`
	var d models.Document
	err := r.db.QueryRowContext(ctx, query, orgID, documentID).Scan(
		&d.ID,
		&d.OrgID,
		&d.TaskID,
		&d.UploadedBy,
		&d.Title,
		&d.Filename,
		&d.MimeType,
		&d.FileSize,
		&d.SHA256,
		&d.StoragePath,
		&d.ExtractedText,
		&d.Status,
		&d.VerifiedBy,
		&d.VerifiedAt,
		&d.VerificationNotes,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.UploadedByName,
		&d.VerifiedByName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DocumentRepository) UpdateStatus(ctx context.Context, orgID, documentID, verifiedBy uuid.UUID, status, notes string) error {
	query := `
		UPDATE documents
		SET status = $1, verified_by = $2, verified_at = NOW(), verification_notes = $3, updated_at = NOW()
		WHERE org_id = $4 AND id = $5
	`
	result, err := r.db.ExecContext(ctx, query, status, verifiedBy, notes, orgID, documentID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *DocumentRepository) ListPending(ctx context.Context, orgID uuid.UUID, limit int) ([]models.Document, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT 
			d.id, d.org_id, d.task_id, d.uploaded_by, d.title, d.filename, d.mime_type, 
			d.file_size, d.sha256, d.status, d.verified_by, d.verified_at, d.verification_notes,
			d.created_at, d.updated_at,
			CASE WHEN u.id IS NULL THEN NULL ELSE CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) END AS uploaded_by_name,
			CASE WHEN v.id IS NULL THEN NULL ELSE CONCAT(COALESCE(v.first_name, ''), ' ', COALESCE(v.last_name, '')) END AS verified_by_name
		FROM documents d
		LEFT JOIN users u ON u.id = d.uploaded_by
		LEFT JOIN users v ON v.id = d.verified_by
		WHERE d.org_id = $1 AND d.status = 'submitted'
		ORDER BY d.created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Document{}
	for rows.Next() {
		var d models.Document
		err := rows.Scan(
			&d.ID,
			&d.OrgID,
			&d.TaskID,
			&d.UploadedBy,
			&d.Title,
			&d.Filename,
			&d.MimeType,
			&d.FileSize,
			&d.SHA256,
			&d.Status,
			&d.VerifiedBy,
			&d.VerifiedAt,
			&d.VerificationNotes,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.UploadedByName,
			&d.VerifiedByName,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentRepository) GetExtractedText(ctx context.Context, orgID, documentID uuid.UUID) (string, error) {
	query := `SELECT COALESCE(extracted_text, '') FROM documents WHERE org_id = $1 AND id = $2`
	var text string
	err := r.db.QueryRowContext(ctx, query, orgID, documentID).Scan(&text)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return text, err
}

type StoredChunk struct {
	Chunk models.DocumentChunk
	// Embedding may be nil if not computed.
	Embedding []float64
}

func (r *DocumentRepository) CreateChunks(ctx context.Context, orgID, documentID uuid.UUID, chunks []string, embeddings [][]float32) error {
	if len(chunks) == 0 {
		return nil
	}
	if embeddings != nil && len(embeddings) != len(chunks) {
		return fmt.Errorf("embeddings length mismatch")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO document_chunks (id, org_id, document_id, chunk_index, content, embedding, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for i, content := range chunks {
		var embJSON []byte
		if embeddings != nil && embeddings[i] != nil {
			embJSON, err = json.Marshal(embeddings[i])
			if err != nil {
				return err
			}
		}
		_, err = stmt.ExecContext(
			ctx,
			uuid.New(),
			orgID,
			documentID,
			i,
			content,
			embJSON,
			now,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *DocumentRepository) ListChunks(ctx context.Context, orgID, documentID uuid.UUID, limit int) ([]StoredChunk, error) {
	if limit <= 0 || limit > 2000 {
		limit = 1000
	}
	query := `
		SELECT id, org_id, document_id, chunk_index, content, embedding, created_at
		FROM document_chunks
		WHERE org_id = $1 AND document_id = $2
		ORDER BY chunk_index ASC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, documentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []StoredChunk{}
	for rows.Next() {
		var sc StoredChunk
		var embRaw []byte
		err := rows.Scan(
			&sc.Chunk.ID,
			&sc.Chunk.OrgID,
			&sc.Chunk.DocumentID,
			&sc.Chunk.ChunkIndex,
			&sc.Chunk.Content,
			&embRaw,
			&sc.Chunk.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(embRaw) > 0 {
			var emb32 []float32
			if err := json.Unmarshal(embRaw, &emb32); err == nil {
				emb := make([]float64, 0, len(emb32))
				for _, v := range emb32 {
					emb = append(emb, float64(v))
				}
				sc.Embedding = emb
			}
		}

		out = append(out, sc)
	}
	return out, rows.Err()
}
