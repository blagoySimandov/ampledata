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
	model  string
	client *genai.Client
}

func NewGeminiKeySelector(apiKey string) (*GeminiKeySelector, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return &GeminiKeySelector{
		model:  "gemini-2.5-flash-lite",
		client: client,
	}, nil
}

func (g *GeminiKeySelector) SelectBestKey(ctx context.Context, headers []string, columnsMetadata []*models.ColumnMetadata) (*KeySelectorResult, error) {
	prompt := g.buildPrompt(headers, columnsMetadata)

	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return g.parseResponse(result.Text(), headers)
}

func (g *GeminiKeySelector) buildPrompt(headers []string, columnsMetadata []*models.ColumnMetadata) string {
	var columnsInfo string
	if len(columnsMetadata) > 0 {
		var metadataLines []string
		for _, col := range columnsMetadata {
			desc := ""
			if col.Description != nil {
				desc = fmt.Sprintf(" - %s", *col.Description)
			}
			metadataLines = append(metadataLines, fmt.Sprintf("  - %s [type: %s]%s", col.Name, col.Type, desc))
		}
		columnsInfo = fmt.Sprintf("\n\nColumn Metadata (columns to be enriched):\n%s", strings.Join(metadataLines, "\n"))
	}

	return fmt.Sprintf(`You are analyzing a CSV file to select the best column to use as a unique identifier (key) for data enrichment.

Available CSV Headers:
%s
%s

Your task is to select the BEST column to use as the key for enrichment. Consider:
1. Uniqueness: The column should contain unique identifiers for each row
2. Descriptiveness: The column should be descriptive enough to use for web search and enrichment
3. Common key patterns: Look for columns like "id", "email", "company_name", "organization", "person_name", "domain", etc.
4. Avoid generic columns: Don't select columns like "date", "status", "score", "index", etc.
5. Context from metadata: If column metadata is provided, use it to understand which columns are being enriched (these are likely NOT the key)

Respond ONLY with a JSON object in the following format:
{
  "selected_key": "column_name_here",
  "reasoning": "Brief explanation of why this column was selected"
}

Do not include any text before or after the JSON object.`, strings.Join(headers, "\n"), columnsInfo)
}

func (g *GeminiKeySelector) parseResponse(response string, headers []string) (*KeySelectorResult, error) {
	// Clean up the response - remove markdown code blocks if present
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

	// Validate that the selected key exists in headers
	found := false
	for _, header := range headers {
		if header == parsed.SelectedKey {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("selected key '%s' not found in CSV headers: %v", parsed.SelectedKey, headers)
	}

	return &KeySelectorResult{
		SelectedKey: parsed.SelectedKey,
		AllKeys:     headers,
		Reasoning:   parsed.Reasoning,
	}, nil
}
