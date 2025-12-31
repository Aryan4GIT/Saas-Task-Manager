package rag

import (
	"context"
	"fmt"
	"strings"

	"saas-backend/internal/ai"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type Service struct {
	repo         *Repository
	embedder     *Embedder
	retriever    *Retriever
	genClient    *genai.Client
	genModel     *genai.GenerativeModel
	langChainSvc *ai.LangChainService // Optional - for improved prompts
}

func NewService(
	repo *Repository,
	embedder *Embedder,
	apiKey string,
	generationModel string,
	langChainSvc *ai.LangChainService, // Optional
) (*Service, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	if generationModel == "" {
		generationModel = "gemini-2.5-flash"
	}

	model := client.GenerativeModel(generationModel)
	model.SetTemperature(0.1)
	model.SetTopP(0.8)
	model.SetTopK(40)

	retriever := NewRetriever(repo, embedder)

	return &Service{
		repo:         repo,
		embedder:     embedder,
		retriever:    retriever,
		genClient:    client,
		genModel:     model,
		langChainSvc: langChainSvc,
	}, nil
}

type IndexRequest struct {
	OrgID      uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Content    string
}

func (s *Service) IndexDocument(ctx context.Context, req IndexRequest) error {
	if req.Content == "" {
		return nil
	}

	validSourceTypes := map[string]bool{"task": true, "issue": true, "comment": true, "document": true}
	if !validSourceTypes[req.SourceType] {
		return fmt.Errorf("invalid source type: %s", req.SourceType)
	}

	embedding, err := s.embedder.Embed(ctx, req.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	doc := &RAGDocument{
		ID:         uuid.New(),
		OrgID:      req.OrgID,
		SourceType: req.SourceType,
		SourceID:   req.SourceID,
		Content:    req.Content,
		Embedding:  embedding,
	}

	if err := s.repo.Upsert(doc); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	return nil
}

func (s *Service) DeleteDocument(ctx context.Context, orgID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	return s.repo.DeleteBySource(orgID, sourceType, sourceID)
}

type QueryRequest struct {
	OrgID    uuid.UUID
	UserID   uuid.UUID
	Role     string
	Question string
}

type QueryResponse struct {
	Answer  string            `json:"answer"`
	Sources []SimilarDocument `json:"sources"`
}

func (s *Service) Query(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	if req.Question == "" {
		return nil, fmt.Errorf("question cannot be empty")
	}

	validRoles := map[string]bool{"admin": true, "manager": true, "member": true}
	if !validRoles[req.Role] {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	// Check if this is a general/conversational question
	if isGeneralQuestion(req.Question) {
		answer, err := s.handleGeneralQuestion(ctx, req.Question)
		if err != nil {
			return nil, fmt.Errorf("failed to handle general question: %w", err)
		}
		return &QueryResponse{
			Answer:  answer,
			Sources: []SimilarDocument{},
		}, nil
	}

	result, err := s.retriever.Retrieve(ctx, RetrievalParams{
		OrgID:  req.OrgID,
		UserID: req.UserID,
		Role:   req.Role,
		Query:  req.Question,
		Limit:  5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	if len(result.Documents) == 0 {
		// No documents found - try to give a helpful response
		answer, err := s.handleNoDocumentsFound(ctx, req.Question)
		if err != nil {
			return &QueryResponse{
				Answer:  "I couldn't find any relevant tasks or issues matching your query. Try asking about specific tasks, issues, or project details.",
				Sources: []SimilarDocument{},
			}, nil
		}
		return &QueryResponse{
			Answer:  answer,
			Sources: []SimilarDocument{},
		}, nil
	}

	contextBlock := buildContextBlock(result.Documents)

	answer, err := s.generateAnswer(ctx, req.Question, contextBlock, result.Documents)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	return &QueryResponse{
		Answer:  answer,
		Sources: result.Documents,
	}, nil
}

func buildContextBlock(docs []SimilarDocument) string {
	var sb strings.Builder
	for i, doc := range docs {
		sb.WriteString(fmt.Sprintf("--- Document %d [%s] ---\n", i+1, doc.SourceType))
		sb.WriteString(doc.Content)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func (s *Service) generateAnswer(ctx context.Context, question string, contextBlock string, documents []SimilarDocument) (string, error) {
	// Try LangChain first for improved prompts and grounding
	if s.langChainSvc != nil {
		// Convert documents to string slice for LangChain prompt
		docStrings := make([]string, len(documents))
		for i, doc := range documents {
			docStrings[i] = fmt.Sprintf("[%s] %s", doc.SourceType, doc.Content)
		}
		prompt := ai.RAGQueryPrompt(docStrings, question)
		answer, err := s.langChainSvc.GenerateText(ctx, prompt)
		if err == nil && answer != "" {
			return strings.TrimSpace(answer), nil
		}
		// Log but don't fail - fall back to direct Gemini
		fmt.Printf("LangChain generation failed: %v, falling back to Gemini\n", err)
	}

	// Fallback to direct Gemini API
	prompt := fmt.Sprintf(`You are a helpful assistant that answers questions using ONLY the provided context.

STRICT RULES:
1. Answer ONLY using information from the context below
2. If the context does not contain enough information to answer the question, respond with: "Insufficient data to answer."
3. Do NOT make up or infer information not present in the context
4. Do NOT use any external knowledge
5. Be concise and direct in your response
6. Reference the source type (task, issue, comment) when relevant

CONTEXT:
%s

USER QUESTION:
%s

ANSWER:`, contextBlock, question)

	resp, err := s.genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "Insufficient data to answer.", nil
	}

	answer := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return strings.TrimSpace(answer), nil
}

func (s *Service) Close() error {
	var errs []error
	if s.genClient != nil {
		if err := s.genClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.embedder != nil {
		if err := s.embedder.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing RAG service: %v", errs)
	}
	return nil
}

// isGeneralQuestion checks if the question is a general/conversational query
func isGeneralQuestion(question string) bool {
	q := strings.ToLower(strings.TrimSpace(question))

	// Greetings and conversational phrases
	greetings := []string{
		"hello", "hi", "hey", "good morning", "good afternoon", "good evening",
		"howdy", "greetings", "what's up", "whats up", "sup",
	}
	for _, g := range greetings {
		if q == g || strings.HasPrefix(q, g+" ") || strings.HasPrefix(q, g+"!") || strings.HasPrefix(q, g+",") {
			return true
		}
	}

	// Questions about the AI itself
	aboutAI := []string{
		"who are you", "what are you", "what can you do", "how do i use you",
		"how to use you", "help", "what is this", "how does this work",
		"what can i ask", "what should i ask", "give me examples",
		"how can you help", "what do you do", "introduce yourself",
		"tell me about yourself", "your capabilities", "your features",
	}
	for _, phrase := range aboutAI {
		if strings.Contains(q, phrase) {
			return true
		}
	}

	// Thank you messages
	thanks := []string{"thank you", "thanks", "thx", "ty", "appreciate"}
	for _, t := range thanks {
		if strings.Contains(q, t) {
			return true
		}
	}

	return false
}

// handleGeneralQuestion generates responses for general/conversational queries
func (s *Service) handleGeneralQuestion(ctx context.Context, question string) (string, error) {
	prompt := `You are a helpful AI assistant integrated into a task and issue management system. 
You help users query their company's tasks, issues, and project data.

The user has asked a general question. Respond naturally and helpfully.
If they're greeting you, greet them back warmly.
If they're asking what you can do, explain that you can:
- Answer questions about their tasks and issues
- Help them find specific tasks by status, priority, or description
- Summarize project activity
- Search through their company data

Keep your response concise (2-4 sentences) and friendly.

User message: ` + question + `

Response:`

	resp, err := s.genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "Hello! I'm here to help you explore your tasks and issues. Just ask me anything about your project data!", nil
	}

	answer := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return strings.TrimSpace(answer), nil
}

// handleNoDocumentsFound provides a helpful response when no relevant documents are found
func (s *Service) handleNoDocumentsFound(ctx context.Context, question string) (string, error) {
	prompt := `You are a helpful AI assistant for a task and issue management system.
The user asked a question, but no relevant tasks or issues were found in their data.

Provide a brief, helpful response that:
1. Acknowledges you couldn't find relevant data
2. Suggests what they could try instead (be specific to their question if possible)
3. Reminds them you can help with tasks, issues, and project data

Keep it to 2-3 sentences. Be helpful, not apologetic.

User question: ` + question + `

Response:`

	resp, err := s.genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "I couldn't find any matching tasks or issues. Try asking about specific project details, task statuses, or recent activity.", nil
	}

	answer := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return strings.TrimSpace(answer), nil
}
