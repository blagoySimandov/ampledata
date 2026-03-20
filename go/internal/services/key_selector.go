package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"google.golang.org/genai"
)

type KeySelectorResult struct {
	SelectedKey string
	AllKeys     []string
	Reasoning   string
}

type KeySelector interface {
	SelectBestKey(ctx context.Context, headers []string, columnsMetadata []*models.ColumnMetadata) (*KeySelectorResult, error)
}

type GeminiKeySelector struct {
	model         string
	client        *genai.Client
	promptService IPromptService
}

func NewGeminiKeySelector(promptService IPromptService) (*GeminiKeySelector, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return &GeminiKeySelector{
		model:         "gemini-2.5-flash-lite",
		client:        client,
		promptService: promptService,
	}, nil
}

func (g *GeminiKeySelector) SelectBestKey(ctx context.Context, headers []string, columnsMetadata []*models.ColumnMetadata) (*KeySelectorResult, error) {
	prompt := g.promptService.KeySelectorPrompt(headers, columnsMetadata)

	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return g.parseResponse(result.Text(), headers)
}

func (g *GeminiKeySelector) parseResponse(response string, headers []string) (*KeySelectorResult, error) {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var parsed struct {
		SelectedKey string `json:"selected_key"`
		Reasoning   string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w. Response: %s", err, response)
	}

	for _, header := range headers {
		if header == parsed.SelectedKey {
			return &KeySelectorResult{
				SelectedKey: parsed.SelectedKey,
				AllKeys:     headers,
				Reasoning:   parsed.Reasoning,
			}, nil
		}
	}

	return nil, fmt.Errorf("selected key '%s' not found in CSV headers: %v", parsed.SelectedKey, headers)
}
