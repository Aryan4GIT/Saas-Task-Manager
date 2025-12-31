package service

import (
	"context"
	"fmt"
	"strings"

	"saas-backend/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client         *genai.Client
	model          *genai.GenerativeModel
	embeddingModel *genai.EmbeddingModel
}

func NewGeminiService(cfg *config.Config) (*GeminiService, error) {
	if cfg.Gemini.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key not configured")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Gemini.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	modelName := strings.TrimSpace(cfg.Gemini.Model)
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	model := client.GenerativeModel(modelName)
	model.SetTemperature(0.7)
	model.SetTopP(0.8)
	model.SetTopK(40)

	embedModelName := strings.TrimSpace(cfg.Gemini.EmbeddingModel)
	if embedModelName == "" {
		embedModelName = "text-embedding-004"
	}
	embeddingModel := client.EmbeddingModel(embedModelName)

	return &GeminiService{
		client:         client,
		model:          model,
		embeddingModel: embeddingModel,
	}, nil
}

func (s *GeminiService) EmbedText(text string) ([]float32, error) {
	if s == nil || s.embeddingModel == nil {
		return nil, fmt.Errorf("Gemini embedding model not configured")
	}

	ctx := context.Background()
	resp, err := s.embeddingModel.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("failed to embed content: %w", err)
	}
	if resp == nil || resp.Embedding == nil {
		return nil, fmt.Errorf("no embedding returned")
	}
	return resp.Embedding.Values, nil
}

func (s *GeminiService) GenerateText(prompt string) (string, error) {
	ctx := context.Background()

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return strings.TrimSpace(text), nil
}

func (s *GeminiService) GenerateIssueSummary(title, description string) (string, error) {
	prompt := fmt.Sprintf(`You are a technical support assistant. Analyze the following issue and provide a concise summary (2-3 sentences) focusing on:
1. The core problem
2. Potential impact
3. Suggested priority

Issue Title: %s
Issue Description: %s

Summary:`, title, description)

	return s.GenerateText(prompt)
}

func (s *GeminiService) Close() error {
	return s.client.Close()
}
