package rag

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Retriever struct {
	repo     *Repository
	embedder *Embedder
}

func NewRetriever(repo *Repository, embedder *Embedder) *Retriever {
	return &Retriever{
		repo:     repo,
		embedder: embedder,
	}
}

type RetrievalParams struct {
	OrgID              uuid.UUID
	UserID             uuid.UUID
	Role               string
	Query              string
	Limit              int
	AllowedSourceTypes []string
}

type RetrievalResult struct {
	Documents []SimilarDocument
}

func (r *Retriever) Retrieve(ctx context.Context, params RetrievalParams) (*RetrievalResult, error) {
	if params.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if params.Limit <= 0 || params.Limit > 5 {
		params.Limit = 5
	}

	queryEmbedding, err := r.embedder.Embed(ctx, params.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	var docs []SimilarDocument

	switch params.Role {
	case "admin", "manager":
		docs, err = r.repo.FindSimilar(
			params.OrgID,
			queryEmbedding,
			params.AllowedSourceTypes,
			params.Limit,
		)
	case "member":
		docs, err = r.repo.FindSimilarForMember(
			params.OrgID,
			params.UserID,
			queryEmbedding,
			params.Limit,
		)
	default:
		return nil, fmt.Errorf("invalid role: %s", params.Role)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	return &RetrievalResult{Documents: docs}, nil
}
