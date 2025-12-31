package rag

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type RAGDocument struct {
	ID         uuid.UUID `json:"id"`
	OrgID      uuid.UUID `json:"org_id"`
	SourceType string    `json:"source_type"`
	SourceID   uuid.UUID `json:"source_id"`
	Content    string    `json:"content"`
	Embedding  []float32 `json:"-"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(doc *RAGDocument) error {
	if len(doc.Embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}

	embeddingStr := floatsToVectorString(doc.Embedding)

	query := `
		INSERT INTO rag_documents (id, org_id, source_type, source_id, content, embedding, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::vector, NOW())
		ON CONFLICT (org_id, source_type, source_id)
		DO UPDATE SET 
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding,
			created_at = NOW()
	`

	_, err := r.db.Exec(query,
		doc.ID,
		doc.OrgID,
		doc.SourceType,
		doc.SourceID,
		doc.Content,
		embeddingStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert rag document: %w", err)
	}

	return nil
}

func (r *Repository) DeleteBySource(orgID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	query := `
		DELETE FROM rag_documents 
		WHERE org_id = $1 AND source_type = $2 AND source_id = $3
	`

	_, err := r.db.Exec(query, orgID, sourceType, sourceID)
	if err != nil {
		return fmt.Errorf("failed to delete rag document: %w", err)
	}

	return nil
}

type SimilarDocument struct {
	ID         uuid.UUID `json:"id"`
	SourceType string    `json:"source_type"`
	SourceID   uuid.UUID `json:"source_id"`
	Content    string    `json:"content"`
	Similarity float64   `json:"similarity"`
}

func (r *Repository) FindSimilar(
	orgID uuid.UUID,
	queryEmbedding []float32,
	allowedSourceTypes []string,
	limit int,
) ([]SimilarDocument, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding cannot be empty")
	}

	if limit <= 0 || limit > 10 {
		limit = 5
	}

	embeddingStr := floatsToVectorString(queryEmbedding)

	var args []interface{}
	args = append(args, orgID, embeddingStr, limit)

	sourceTypeFilter := ""
	if len(allowedSourceTypes) > 0 {
		placeholders := make([]string, len(allowedSourceTypes))
		for i, st := range allowedSourceTypes {
			args = append(args, st)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		sourceTypeFilter = fmt.Sprintf("AND source_type IN (%s)", strings.Join(placeholders, ", "))
	}

	query := fmt.Sprintf(`
		SELECT 
			id,
			source_type,
			source_id,
			content,
			1 - (embedding <=> $2::vector) AS similarity
		FROM rag_documents
		WHERE org_id = $1 %s
		ORDER BY embedding <=> $2::vector
		LIMIT $3
	`, sourceTypeFilter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar documents: %w", err)
	}
	defer rows.Close()

	var results []SimilarDocument
	for rows.Next() {
		var doc SimilarDocument
		if err := rows.Scan(&doc.ID, &doc.SourceType, &doc.SourceID, &doc.Content, &doc.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan similar document: %w", err)
		}
		results = append(results, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating similar documents: %w", err)
	}

	return results, nil
}

func (r *Repository) FindSimilarForMember(
	orgID uuid.UUID,
	userID uuid.UUID,
	queryEmbedding []float32,
	limit int,
) ([]SimilarDocument, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding cannot be empty")
	}

	if limit <= 0 || limit > 10 {
		limit = 5
	}

	embeddingStr := floatsToVectorString(queryEmbedding)

	query := `
		SELECT 
			rd.id,
			rd.source_type,
			rd.source_id,
			rd.content,
			1 - (rd.embedding <=> $2::vector) AS similarity
		FROM rag_documents rd
		WHERE rd.org_id = $1
		AND (
			(rd.source_type = 'task' AND EXISTS (
				SELECT 1 FROM tasks t 
				WHERE t.id = rd.source_id 
				AND t.org_id = rd.org_id
				AND (t.assigned_to = $4 OR t.created_by = $4)
			))
			OR
			(rd.source_type = 'issue' AND EXISTS (
				SELECT 1 FROM issues i 
				WHERE i.id = rd.source_id 
				AND i.org_id = rd.org_id
				AND (i.assigned_to = $4 OR i.reported_by = $4)
			))
			OR
			(rd.source_type = 'comment')
		)
		ORDER BY rd.embedding <=> $2::vector
		LIMIT $3
	`

	rows, err := r.db.Query(query, orgID, embeddingStr, limit, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar documents for member: %w", err)
	}
	defer rows.Close()

	var results []SimilarDocument
	for rows.Next() {
		var doc SimilarDocument
		if err := rows.Scan(&doc.ID, &doc.SourceType, &doc.SourceID, &doc.Content, &doc.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan similar document: %w", err)
		}
		results = append(results, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating similar documents: %w", err)
	}

	return results, nil
}

func floatsToVectorString(embedding []float32) string {
	strs := make([]string, len(embedding))
	for i, v := range embedding {
		strs[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(strs, ",") + "]"
}
