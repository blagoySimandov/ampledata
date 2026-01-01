package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type ExtractionResult struct {
	ExtractedData map[string]interface{} `json:"extracted_data"`
	Reasoning     string                 `json:"reasoning"`
}

type ContentExtractor interface {
	Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata) (*ExtractionResult, error)
}

type GroqContentExtractor struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewGroqContentExtractor(apiKey string) *GroqContentExtractor {
	return &GroqContentExtractor{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "openai/gpt-oss-20b",
	}
}

func (g *GroqContentExtractor) Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata) (*ExtractionResult, error) {
	prompt := g.buildExtractionPrompt(content, columnsMetadata, entityKey)

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

	return g.parseResponse(groqResp.Choices[0].Message.Content)
}

func (g *GroqContentExtractor) buildExtractionPrompt(content string, columnsMetadata []*models.ColumnMetadata, entity string) string {
	var columnsInfo []string
	for _, col := range columnsMetadata {
		desc := ""
		if col.Description != nil {
			desc = fmt.Sprintf(" (%s)", *col.Description)
		}
		columnsInfo = append(columnsInfo, fmt.Sprintf("- %s [type: %s]%s", col.Name, col.Type, desc))
	}
	columnsText := strings.Join(columnsInfo, "\n")

	truncatedContent := content
	if len(content) > 8000 {
		truncatedContent = content[:8000]
	}

	return fmt.Sprintf(`You are a data extraction specialist. Extract the following fields from the provided website content about %s.

## Fields to Extract (ONLY extract these fields)
%s

## Website Content
%s

## Your Task

Extract ONLY the fields listed above from the website content. Do not extract any other fields.

IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata:
- For number types: use numeric values without quotes (e.g., 1000)
- For string types: use quoted strings
- For boolean types: use true/false without quotes
- For date types: use ISO 8601 format (YYYY-MM-DD)

If a field cannot be found in the content, omit it from the response.

## Response Format (JSON only, no markdown)
{
    "extracted_data": {"field_name": value_with_correct_type},
    "reasoning": "Explanation of what was extracted from the content and how you found each field"
}`, entity, columnsText, truncatedContent)
}

func (g *GroqContentExtractor) parseResponse(content string) (*ExtractionResult, error) {
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
