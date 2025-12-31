package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"saas-backend/config"
	"saas-backend/internal/ai"
	"saas-backend/internal/models"
	"saas-backend/internal/rag"
	"saas-backend/internal/repository"
	"saas-backend/internal/utils"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type DocumentService struct {
	docRepo       *repository.DocumentRepository
	geminiService *GeminiService
	langChainSvc  *ai.LangChainService
	ragIndexer    *rag.Indexer
	cfg           *config.Config
}

func NewDocumentService(docRepo *repository.DocumentRepository, geminiService *GeminiService, langChainSvc *ai.LangChainService, ragIndexer *rag.Indexer, cfg *config.Config) *DocumentService {
	return &DocumentService{docRepo: docRepo, geminiService: geminiService, langChainSvc: langChainSvc, ragIndexer: ragIndexer, cfg: cfg}
}

func (s *DocumentService) Upload(ctx context.Context, orgID, userID uuid.UUID, taskID *uuid.UUID, fh *multipart.FileHeader, title *string) (*models.Document, error) {
	if fh == nil {
		return nil, fmt.Errorf("missing file")
	}

	if err := os.MkdirAll(s.cfg.Server.UploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open upload: %w", err)
	}
	defer src.Close()

	// Limit to 15MB to avoid runaway memory/disk usage.
	const maxBytes = int64(15 << 20)
	limited := &io.LimitedReader{R: src, N: maxBytes + 1}

	h := sha256.New()
	var buf bytes.Buffer
	written, err := io.Copy(io.MultiWriter(&buf, h), limited)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload: %w", err)
	}
	if written > maxBytes {
		return nil, fmt.Errorf("file too large (max 15MB)")
	}

	sum := hex.EncodeToString(h.Sum(nil))

	// Store file on disk
	docID := uuid.New()
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	storedName := fmt.Sprintf("%s%s", docID.String(), ext)
	storedPath := filepath.Join(s.cfg.Server.UploadDir, storedName)

	if err := os.WriteFile(storedPath, buf.Bytes(), 0o644); err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	mimeType := strings.TrimSpace(fh.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	extracted, err := s.extractText(storedPath, mimeType, buf.Bytes())
	if err != nil {
		return nil, err
	}

	doc := &models.Document{
		ID:            docID,
		OrgID:         orgID,
		TaskID:        taskID,
		UploadedBy:    &userID,
		Title:         title,
		Filename:      fh.Filename,
		MimeType:      &mimeType,
		FileSize:      int64(len(buf.Bytes())),
		SHA256:        sum,
		StoragePath:   storedPath,
		ExtractedText: &extracted,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	chunks := utils.ChunkTextRunes(extracted, 1200, 200, 200)
	var embeddings [][]float32
	if s.geminiService != nil {
		embeddings = make([][]float32, 0, len(chunks))
		for _, c := range chunks {
			emb, err := s.geminiService.EmbedText(c)
			if err != nil {
				// If embeddings fail, fall back to keyword retrieval.
				embeddings = nil
				break
			}
			embeddings = append(embeddings, emb)
		}
	}

	if err := s.docRepo.CreateChunks(ctx, orgID, docID, chunks, embeddings); err != nil {
		return nil, fmt.Errorf("failed to store chunks: %w", err)
	}

	// Index document into RAG for Ask AI queries
	if s.ragIndexer != nil {
		titleStr := fh.Filename
		if title != nil && *title != "" {
			titleStr = *title
		}
		content := fmt.Sprintf("Document: %s\n\nContent:\n%s", titleStr, extracted)
		go func() {
			_ = s.ragIndexer.IndexDocument(context.Background(), orgID, docID, titleStr, content)
		}()
	}

	// Do not expose extracted text/storage path in API; model struct already hides them.
	return doc, nil
}

func (s *DocumentService) List(ctx context.Context, orgID uuid.UUID, limit int) ([]models.Document, error) {
	return s.docRepo.List(ctx, orgID, limit)
}

func (s *DocumentService) ListByTask(ctx context.Context, orgID, taskID uuid.UUID, limit int) ([]models.Document, error) {
	return s.docRepo.ListByTask(ctx, orgID, taskID, limit)
}

func (s *DocumentService) ListPending(ctx context.Context, orgID uuid.UUID, limit int) ([]models.Document, error) {
	return s.docRepo.ListPending(ctx, orgID, limit)
}

func (s *DocumentService) GetByID(ctx context.Context, orgID, documentID uuid.UUID) (*models.Document, error) {
	return s.docRepo.GetByID(ctx, orgID, documentID)
}

func (s *DocumentService) UpdateStatus(ctx context.Context, orgID, documentID, verifiedBy uuid.UUID, status, notes string) error {
	// Validate status
	if status != "verified" && status != "rejected" {
		return fmt.Errorf("invalid status: must be 'verified' or 'rejected'")
	}
	return s.docRepo.UpdateStatus(ctx, orgID, documentID, verifiedBy, status, notes)
}

func (s *DocumentService) GenerateSummary(ctx context.Context, orgID, documentID uuid.UUID) (*models.DocumentSummaryResponse, error) {
	if s.langChainSvc == nil && s.geminiService == nil {
		return nil, fmt.Errorf("AI service not configured")
	}

	text, err := s.docRepo.GetExtractedText(ctx, orgID, documentID)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, fmt.Errorf("document has no extracted text")
	}

	// Truncate if too long
	if len(text) > 10000 {
		text = text[:10000] + "\n...[truncated]"
	}

	var out string
	if s.langChainSvc != nil {
		// Use LangChain with structured prompt
		prompt := fmt.Sprintf(`You are an AI assistant for document verification. Analyze the following document and provide a structured summary.

Return STRICT JSON with this shape:
{
  "summary": "A concise 2-3 sentence summary of the document's main content",
  "key_points": ["list of 3-5 key points or findings"],
  "document_type": "work_report | technical_doc | meeting_notes | proposal | other",
  "quality_assessment": "high | medium | low",
  "verification_recommendation": "approve | needs_review | reject",
  "notes": "Any additional notes or concerns for the reviewer"
}

Document content:
%s`, text)
		out, err = s.langChainSvc.GenerateText(ctx, prompt)
	} else {
		// Fallback to old Gemini service
		prompt := fmt.Sprintf(`You are an AI assistant for document verification. Analyze the following document and provide a structured summary.

Return STRICT JSON with this shape:
{
  "summary": "A concise 2-3 sentence summary of the document's main content",
  "key_points": ["list of 3-5 key points or findings"],
  "document_type": "work_report | technical_doc | meeting_notes | proposal | other",
  "quality_assessment": "high | medium | low",
  "verification_recommendation": "approve | needs_review | reject",
  "notes": "Any additional notes or concerns for the reviewer"
}

Document content:
%s`, text)
		out, err = s.geminiService.GenerateText(prompt)
	}

	if err != nil {
		return nil, err
	}

	var resp models.DocumentSummaryResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		// Fallback: return raw output
		resp.Summary = out
		resp.DocumentType = "other"
		resp.QualityAssessment = "medium"
		resp.VerificationRecommendation = "needs_review"
	}
	return &resp, nil
}

func (s *DocumentService) Verify(ctx context.Context, orgID uuid.UUID, documentID uuid.UUID, question string) (*models.VerifyDocumentResponse, error) {
	if s.geminiService == nil {
		return nil, fmt.Errorf("Gemini service not configured")
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

	// Retrieve top chunks
	topK := 5
	scored := make([]utils.ScoredIndex, 0, len(stored))

	queryEmb, embErr := s.geminiService.EmbedText(question)
	var q64 []float64
	if embErr == nil {
		q64 = make([]float64, 0, len(queryEmb))
		for _, v := range queryEmb {
			q64 = append(q64, float64(v))
		}
	}

	for i, ch := range stored {
		if q64 != nil && len(ch.Embedding) > 0 {
			scored = append(scored, utils.ScoredIndex{Index: i, Score: utils.CosineSimilarity(q64, ch.Embedding)})
			continue
		}
		// Fallback keyword score
		scored = append(scored, utils.ScoredIndex{Index: i, Score: utils.KeywordScore(question, ch.Chunk.Content)})
	}

	selected := utils.TopK(scored, topK)
	contexts := make([]string, 0, len(selected))
	citations := make([]models.VerifyDocumentCitation, 0, len(selected))
	for _, sidx := range selected {
		chunk := stored[sidx.Index].Chunk
		contexts = append(contexts, fmt.Sprintf("[Chunk %d]\n%s", chunk.ChunkIndex, chunk.Content))
		snippet := chunk.Content
		if len([]rune(snippet)) > 240 {
			snippet = string([]rune(snippet)[:240])
		}
		citations = append(citations, models.VerifyDocumentCitation{ChunkIndex: chunk.ChunkIndex, Snippet: snippet})
	}

	prompt := fmt.Sprintf(`You are a document verification assistant.

Use ONLY the provided context excerpts to answer the user's question. Do not use outside knowledge.
If the context does not contain enough evidence, say so.

Return STRICT JSON with this shape:
{
  "verdict": "verified" | "unverified" | "insufficient",
  "confidence": number,
  "answer": string,
  "citations": [{"chunk_index": number, "reason": string}]
}

Rules:
- "confidence" must be between 0 and 1.
- If you cannot support the answer from context, use verdict "insufficient".

User question:
%s

Context excerpts:
%s
`, question, strings.Join(contexts, "\n\n"))

	out, err := s.geminiService.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	// Best-effort JSON parse
	type modelCitation struct {
		ChunkIndex int    `json:"chunk_index"`
		Reason     string `json:"reason"`
	}
	type modelResp struct {
		Verdict    string          `json:"verdict"`
		Confidence float64         `json:"confidence"`
		Answer     string          `json:"answer"`
		Citations  []modelCitation `json:"citations"`
	}

	resp := &models.VerifyDocumentResponse{RawModelOut: out}
	var mr modelResp
	if err := json.Unmarshal([]byte(out), &mr); err == nil {
		resp.Verdict = strings.TrimSpace(mr.Verdict)
		resp.Confidence = mr.Confidence
		resp.Answer = strings.TrimSpace(mr.Answer)
		// Keep snippets from retrieved chunks (stable + safe); we still include the model's chunk indices.
		resp.Citations = citations
		resp.RawModelOut = ""
		return resp, nil
	}

	// Fallback: return raw output
	resp.Verdict = "insufficient"
	resp.Confidence = 0
	resp.Answer = strings.TrimSpace(out)
	resp.Citations = citations
	return resp, nil
}

func (s *DocumentService) extractText(path string, mimeType string, raw []byte) (string, error) {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	ext := strings.ToLower(filepath.Ext(path))

	if mimeType == "application/pdf" || ext == ".pdf" {
		f, r, err := pdf.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open PDF: %w", err)
		}
		defer f.Close()

		plain, err := r.GetPlainText()
		if err != nil {
			return "", fmt.Errorf("failed to extract PDF text: %w", err)
		}
		b, err := io.ReadAll(plain)
		if err != nil {
			return "", fmt.Errorf("failed to read PDF text: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	}

	// Treat common text-like files as UTF-8 text.
	if strings.HasPrefix(mimeType, "text/") || ext == ".txt" || ext == ".md" || ext == ".json" || ext == ".csv" {
		return strings.TrimSpace(string(raw)), nil
	}

	return "", fmt.Errorf("unsupported file type for extraction: %s", mimeType)
}
