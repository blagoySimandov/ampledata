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
	"google.golang.org/genai"
)

type QueryPatternGenerator interface {
	GeneratePatterns(ctx context.Context, columnsMetadata []*models.ColumnMetadata) ([]string, error)
}

type GeminiPatternGenerator struct {
	model  string
	client *genai.Client
}

func NewGeminiPatternGenerator(apiKey string) (*GeminiPatternGenerator, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &GeminiPatternGenerator{
		model:  "gemini-2.5-flash-lite",
		client: client,
	}, nil
}

func (g *GeminiPatternGenerator) GeneratePatterns(ctx context.Context, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	prompt := g.buildPrompt(columnsMetadata)

	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to generate content: %w", err)
	}

	return g.parseResponse(result.Text(), columnsMetadata)
}

func (g *GeminiPatternGenerator) buildPrompt(columnsMetadata []*models.ColumnMetadata) string {
	var columnsInfo []string
	for _, col := range columnsMetadata {
		desc := ""
		if col.Description != nil {
			desc = fmt.Sprintf(": %s", *col.Description)
		}
		columnsInfo = append(columnsInfo, fmt.Sprintf("- %s (type: %s)%s", col.Name, col.Type, desc))
	}
	columnsText := strings.Join(columnsInfo, "\n")

	return fmt.Sprintf(`You are a Google search query optimization expert. Your task is to create query PATTERNS
for finding information about entities and their attributes.

ENTITY FORMAT EXAMPLES:
- company:openai
- company:stripe
- company:anthropic

COLUMNS TO FIND (%d columns):
%s

YOUR TASK:
1. Analyze the columns and determine the optimal number of search queries (1-3 patterns)
2. Group related columns together in the same query pattern
3. Write queries in natural language that a human would type into Google
4. Use ONLY the %%entity placeholder - everything else should be natural language
5. Balance thoroughness vs SERP API costs:
   - Use 1 pattern if columns are closely related (1-5 columns)
   - Use 2-3 patterns if columns are diverse or many (6-15 columns)

CRITICAL RULES:
- Only use %%entity as a placeholder (replaced with entity value at runtime)
- Do NOT create placeholders from column names
- Write realistic Google search queries in natural language
- Return ONLY valid JSON array of pattern strings
- Each pattern max 150 characters
- Prioritize search accuracy over minimizing queries

CORRECT EXAMPLES:
Input: Columns=[founder, founder_picture_url, website, employee_count]
Output: ["%%entity founder picture", "%%entity company website employee count"]

Input: Columns=[CEO, Revenue, Founded, Employees]
Output: ["%%entity company CEO revenue founded employees"]

Input: Columns=[CEO, Industry, Revenue, Market_Cap, Headquarters]
Output: [
  "%%entity company CEO headquarters founded",
  "%%entity industry revenue market cap financials"
]

Respond with JSON only, no markdown:`, len(columnsMetadata), columnsText)
}

func (g *GeminiPatternGenerator) parseResponse(content string, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	content = cleanJSONMarkdown(content)
	content = strings.TrimSpace(content)

	var patterns []string
	if err := json.Unmarshal([]byte(content), &patterns); err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(patterns) == 0 || len(patterns) > 5 {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("invalid number of patterns: %d (expected 1-5)", len(patterns))
	}

	for _, pattern := range patterns {
		if len(pattern) > 150 {
			return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("pattern too long: %d chars (max 150)", len(pattern))
		}
	}

	return patterns, nil
}

func (g *GeminiPatternGenerator) getFallbackPatterns(columnsMetadata []*models.ColumnMetadata) []string {
	parts := []string{"%entity"}
	for _, col := range columnsMetadata {
		parts = append(parts, col.Name)
	}
	return []string{strings.Join(parts, " ")}
}

type GroqPatternGenerator struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewGroqPatternGenerator(apiKey string) *GroqPatternGenerator {
	return &GroqPatternGenerator{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		model: "openai/gpt-oss-20b",
	}
}

func (g *GroqPatternGenerator) GeneratePatterns(ctx context.Context, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	prompt := g.buildPrompt(columnsMetadata)

	reqBody := map[string]interface{}{
		"model": g.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature":           0,
		"max_completion_tokens": 1024,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("no response from LLM")
	}

	return g.parseResponse(groqResp.Choices[0].Message.Content, columnsMetadata)
}

func (g *GroqPatternGenerator) buildPrompt(columnsMetadata []*models.ColumnMetadata) string {
	var columnsInfo []string
	for _, col := range columnsMetadata {
		desc := ""
		if col.Description != nil {
			desc = fmt.Sprintf(": %s", *col.Description)
		}
		columnsInfo = append(columnsInfo, fmt.Sprintf("- %%%s (type: %s)%s", col.Name, col.Type, desc))
	}
	columnsText := strings.Join(columnsInfo, "\n")

	return fmt.Sprintf(`You are a Google search query optimization expert. Your task is to create query PATTERNS
(templates) for finding information about entities and their attributes.

COLUMNS TO FIND (%d columns):
%s

YOUR TASK:
1. Analyze the columns and determine the optimal number of search queries (1-3 patterns)
2. Group related columns together in the same query pattern
3. Create patterns using placeholders:
   - %%entity = the entity being searched (e.g., company name, person name)
   - %%ColumnName = column name placeholders (e.g., %%CEO, %%Revenue)
4. Balance thoroughness vs SERP API costs:
   - Use 1 pattern if columns are closely related (1-5 columns)
   - Use 2-3 patterns if columns are diverse or many (6-15 columns)
5. Include strategic keywords (e.g., "company", "financials", "founded")

RULES:
- Return ONLY valid JSON array of pattern strings
- Each pattern max 150 characters
- Use quotes strategically for exact phrases
- Prioritize search accuracy over minimizing queries

EXAMPLES:
Input: Columns=[CEO, Revenue, Founded, Employees]
Output: ["%%entity company %%CEO %%Revenue %%Founded %%Employees"]

Input: Columns=[CEO, Founded, Industry, Revenue, Market_Cap, Headquarters, Website, Description]
Output: [
  "%%entity company %%CEO %%Founded %%Headquarters",
  "%%entity %%Industry %%Revenue %%Market_Cap financials",
  "%%entity %%Website %%Description about"
]

Respond with JSON only, no markdown:`, len(columnsMetadata), columnsText)
}

func (g *GroqPatternGenerator) parseResponse(content string, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	content = cleanJSONMarkdown(content)
	content = strings.TrimSpace(content)

	var patterns []string
	if err := json.Unmarshal([]byte(content), &patterns); err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(patterns) == 0 || len(patterns) > 5 {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("invalid number of patterns: %d (expected 1-5)", len(patterns))
	}

	for _, pattern := range patterns {
		if len(pattern) > 150 {
			return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("pattern too long: %d chars (max 150)", len(pattern))
		}
	}

	return patterns, nil
}

func (g *GroqPatternGenerator) getFallbackPatterns(columnsMetadata []*models.ColumnMetadata) []string {
	parts := []string{"%entity"}
	for _, col := range columnsMetadata {
		parts = append(parts, col.Name)
	}
	return []string{strings.Join(parts, " ")}
}
