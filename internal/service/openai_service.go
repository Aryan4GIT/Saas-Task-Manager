package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIService struct {
	APIKey string
	Model  string
}

func NewOpenAIService(apiKey, model string) *OpenAIService {
	return &OpenAIService{APIKey: apiKey, Model: model}
}

func (s *OpenAIService) GenerateText(prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	body := map[string]interface{}{
		"model":    s.Model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}
	return result.Choices[0].Message.Content, nil
}
