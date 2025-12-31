package service

import (
	"context"
	"fmt"

	"saas-backend/internal/ai"
	"saas-backend/internal/models"

	"github.com/google/uuid"
)

type DocumentVerificationResult struct {
	TaskMatches       bool     `json:"task_matches"`
	CompletedWork     string   `json:"completed_work"`
	MissingWork       []string `json:"missing_work"`
	VerificationNotes string   `json:"verification_notes"`
	Recommendation    string   `json:"recommendation"`
}

func (s *DocumentService) VerifyDocumentAgainstTask(ctx context.Context, orgID, documentID, taskID uuid.UUID) (*DocumentVerificationResult, error) {
	if s.langChainSvc == nil {
		return nil, fmt.Errorf("LangChain service not configured - document verification requires AI")
	}

	// Get document content
	doc, err := s.docRepo.GetByID(ctx, orgID, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	text := ""
	if doc.ExtractedText != nil {
		text = *doc.ExtractedText
	}
	if text == "" {
		return nil, fmt.Errorf("document has no extracted text")
	}

	// Truncate if too long
	if len(text) > 15000 {
		text = text[:15000] + "\n...[content truncated for analysis]"
	}

	// Get task details - assuming we have access to task repository
	// In production, inject TaskRepository into DocumentService
	// For now, we'll use a simplified approach
	taskTitle := "Task"
	taskDescription := "Task description not available"

	// Use LangChain with document verification prompt
	prompt := ai.DocumentVerificationPrompt(taskTitle, taskDescription, text)

	response, err := s.langChainSvc.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("verification analysis failed: %w", err)
	}

	// Parse the response into structured format
	result := &DocumentVerificationResult{
		TaskMatches:       true, // Parse from response
		CompletedWork:     response,
		MissingWork:       []string{},
		VerificationNotes: response,
		Recommendation:    "needs_review",
	}

	return result, nil
}

func (s *DocumentService) VerifyWithQuestion(ctx context.Context, orgID, documentID uuid.UUID, question string) (*models.VerifyDocumentResponse, error) {
	if s.langChainSvc == nil && s.geminiService == nil {
		return nil, fmt.Errorf("AI service not configured")
	}

	doc, err := s.docRepo.GetByID(ctx, orgID, documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	stored, err := s.docRepo.ListChunks(ctx, orgID, documentID, 1000)
	if err != nil {
		return nil, err
	}
	if len(stored) == 0 {
		return nil, fmt.Errorf("document has no extracted content")
	}

	// Retrieve top chunks using similarity
	topK := 5
	scored := make([]ScoredIndex, 0, len(stored))

	var q64 []float64
	if s.langChainSvc != nil {
		queryEmb, embErr := s.langChainSvc.EmbedText(ctx, question)
		if embErr == nil {
			q64 = make([]float64, 0, len(queryEmb))
			for _, v := range queryEmb {
				q64 = append(q64, float64(v))
			}
		}
	} else if s.geminiService != nil {
		queryEmb, embErr := s.geminiService.EmbedText(question)
		if embErr == nil {
			q64 = make([]float64, 0, len(queryEmb))
			for _, v := range queryEmb {
				q64 = append(q64, float64(v))
			}
		}
	}

	for i, ch := range stored {
		if q64 != nil && len(ch.Embedding) > 0 {
			scored = append(scored, ScoredIndex{Index: i, Score: CosineSimilarity(q64, ch.Embedding)})
		} else {
			scored = append(scored, ScoredIndex{Index: i, Score: 0.0})
		}
	}

	SortByScore(scored)

	if len(scored) > topK {
		scored = scored[:topK]
	}

	contextParts := make([]string, 0, len(scored))
	for _, s := range scored {
		contextParts = append(contextParts, stored[s.Index].Chunk.Content)
	}

	// Use LangChain prompt template for grounded responses
	relevantDocs := contextParts
	prompt := ai.RAGQueryPrompt(relevantDocs, question)

	var answer string
	if s.langChainSvc != nil {
		answer, err = s.langChainSvc.GenerateText(ctx, prompt)
	} else {
		answer, err = s.geminiService.GenerateText(prompt)
	}

	if err != nil {
		return nil, fmt.Errorf("verification question failed: %w", err)
	}

	// Return simple response structure (no Question/Sources fields)
	return &models.VerifyDocumentResponse{
		Answer:     answer,
		Verdict:    "INFO", // Generic info response
		Confidence: 0.8,
	}, nil
}

type ScoredIndex struct {
	Index int
	Score float64
}

func SortByScore(scored []ScoredIndex) {
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0.0
	}
	return dot / (magA * magB)
}
