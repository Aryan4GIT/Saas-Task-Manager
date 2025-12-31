package ai

import (
	"context"
	"fmt"

	"saas-backend/config"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type LangChainService struct {
	llm       llms.Model
	embedder  embeddings.Embedder
	apiKey    string
	modelName string
}

func NewLangChainService(cfg *config.Config) (*LangChainService, error) {
	if cfg.Gemini.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key not configured")
	}

	modelName := cfg.Gemini.Model
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	llm, err := googleai.New(
		context.Background(),
		googleai.WithAPIKey(cfg.Gemini.APIKey),
		googleai.WithDefaultModel(modelName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LangChain LLM: %w", err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &LangChainService{
		llm:       llm,
		embedder:  embedder,
		apiKey:    cfg.Gemini.APIKey,
		modelName: modelName,
	}, nil
}

func (s *LangChainService) GenerateText(ctx context.Context, prompt string) (string, error) {
	result, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}
	return result, nil
}

func (s *LangChainService) EmbedText(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.embedder.EmbedQuery(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	result := make([]float32, len(embeddings))
	for i, v := range embeddings {
		result[i] = float32(v)
	}
	return result, nil
}

func (s *LangChainService) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("batch embedding failed: %w", err)
	}

	result := make([][]float32, len(embeddings))
	for i, emb := range embeddings {
		result[i] = make([]float32, len(emb))
		for j, v := range emb {
			result[i][j] = float32(v)
		}
	}
	return result, nil
}

func (s *LangChainService) Close() error {
	return nil
}
