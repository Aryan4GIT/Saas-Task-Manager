package rag

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Embedder struct {
	client *genai.Client
	model  *genai.EmbeddingModel
}

func NewEmbedder(apiKey string, modelName string) (*Embedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	if modelName == "" {
		modelName = "text-embedding-004"
	}

	model := client.EmbeddingModel(modelName)

	return &Embedder{
		client: client,
		model:  model,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e == nil || e.model == nil {
		return nil, fmt.Errorf("embedder not initialized")
	}

	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	resp, err := e.model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("failed to embed content: %w", err)
	}

	if resp == nil || resp.Embedding == nil || len(resp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("no embedding returned from Gemini")
	}

	return resp.Embedding.Values, nil
}

func (e *Embedder) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}
