package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"google.golang.org/genai"
)

type ExtractionResult struct {
	ExtractedData map[string]interface{} `json:"extracted_data"`
	Reasoning     string                 `json:"reasoning"`
}

type ContentExtractor interface {
	Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata) (*ExtractionResult, error)
}

type GroqContentExtractor struct {
	apiKey                  string
	httpClient              *http.Client
	model                   string
	extractionPromptBuilder IExtractionPromptBuilder
}

type GeminiContentExtractor struct {
	model                   string
	client                  *genai.Client
	extractionPromptBuilder IExtractionPromptBuilder
}

func NewGeminiContentExtractor(apiKey string) (*GeminiContentExtractor, error) {
	ctx := context.Background() // ctx used for auth/initilization, not passed to later requests
	client, err := genai.NewClient(ctx, nil)
	model := "gemini-2.0-flash-lite"
	if err != nil {
		return nil, err
	}
	return &GeminiContentExtractor{
		model:                   model,
		client:                  client,
		extractionPromptBuilder: NewExtractionPromptBuilder(),
	}, nil
}

func (g *GeminiContentExtractor) Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata) (*ExtractionResult, error) {
	prompt := g.extractionPromptBuilder.Build(content, columnsMetadata, entityKey)
	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	er, err := parseResponse(result.Text())
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return er, nil
}

func NewGroqContentExtractor(apiKey string) *GroqContentExtractor {
	return &GroqContentExtractor{
		apiKey:                  apiKey,
		extractionPromptBuilder: NewExtractionPromptBuilder(),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "openai/gpt-oss-20b",
	}
}

func (g *GroqContentExtractor) Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata) (*ExtractionResult, error) {
	prompt := g.extractionPromptBuilder.Build(content, columnsMetadata, entityKey)

	reqBody := map[string]interface{}{
		"model": g.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature":           0,
		"max_completion_tokens": 2048,
		"reasoning_effort":      "medium",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	return parseResponse(groqResp.Choices[0].Message.Content)
}

func parseResponse(content string) (*ExtractionResult, error) {
	content = cleanJSONMarkdown(content)

	var result ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return &ExtractionResult{
			ExtractedData: make(map[string]interface{}),
			Reasoning:     fmt.Sprintf("Failed to parse LLM response: %s", content[:min(100, len(content))]),
		}, nil
	}

	return &result, nil
}
