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

type CrawlDecision struct {
	URLsToCrawl    []string               `json:"urls_to_crawl"`
	ExtractedData  map[string]interface{} `json:"extracted_data"`
	Reasoning      string                 `json:"reasoning"`
	MissingColumns []string               `json:"-"`
}

type DecisionMaker interface {
	MakeDecision(ctx context.Context, serp *models.GoogleSearchResults, rowKey string, maxURLs int, columnsMetadata []*models.ColumnMetadata) (*CrawlDecision, error)
}

type GroqDecisionMaker struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewGroqDecisionMaker(apiKey string) *GroqDecisionMaker {
	return &GroqDecisionMaker{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "openai/gpt-oss-20b",
	}
}

func (g *GroqDecisionMaker) MakeDecision(ctx context.Context, serp *models.GoogleSearchResults, rowKey string, maxURLs int, columnsMetadata []*models.ColumnMetadata) (*CrawlDecision, error) {
	prompt := g.buildPrompt(serp, rowKey, maxURLs, columnsMetadata)

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

	return g.parseResponse(groqResp.Choices[0].Message.Content, serp, columnsMetadata)
}

func (g *GroqDecisionMaker) buildPrompt(serp *models.GoogleSearchResults, entity string, maxURLs int, columnsMetadata []*models.ColumnMetadata) string {
	var columnsInfo []string
	for _, col := range columnsMetadata {
		desc := ""
		if col.Description != nil {
			desc = fmt.Sprintf(" (%s)", *col.Description)
		}
		columnsInfo = append(columnsInfo, fmt.Sprintf("- %s [type: %s]%s", col.Name, col.Type, desc))
	}
	columnsText := strings.Join(columnsInfo, "\n")

	organicResults := ""
	for i, result := range serp.Organic {
		if i >= 10 {
			break
		}
		title := ""
		if result.Title != nil {
			title = *result.Title
		}
		link := ""
		if result.Link != nil {
			link = *result.Link
		}
		snippet := ""
		if result.Snippet != nil {
			snippet = *result.Snippet
		}
		pos := i + 1
		if result.Position != nil {
			pos = *result.Position
		}
		organicResults += fmt.Sprintf("\nPosition %d: %s\nURL: %s\nSnippet: %s\n---", pos, title, link, snippet)
	}

	peopleAlsoAsk := ""
	for i, item := range serp.PeopleAlsoAsk {
		if i >= 3 {
			break
		}
		peopleAlsoAsk += fmt.Sprintf("Q: %s A: %s\n", item.Question, item.Snippet)
	}

	return fmt.Sprintf(`You are a data extraction assistant. Analyze these search results for "%s" and decide how to proceed.

## Columns We Need to Extract
%s

## Search Results
%s

## People Also Ask
%s

## Your Task

1. Extract ALL data you can see in the answer box and snippets, even if partial
   - IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata
   - For number types: use numeric values without quotes (e.g., 1000)
   - For string types: use quoted strings
   - For boolean types: use true/false without quotes
   - For date types: use ISO 8601 format (YYYY-MM-DD)

2. Check if you extracted ALL the columns we need:
   - If YES: Return empty urls_to_crawl array
   - If NO: Select up to %d URLs to crawl for missing data, prioritizing:
     * Wikipedia
     * Reliable data sources (SEC filings, financial sites)
     * Avoid SEO aggregator sites when primary sources are available

## Response Format (JSON only, no markdown)
{
    "urls_to_crawl": ["url1", "url2"] or [],
    "extracted_data": {"column_name": value_with_correct_type} or null,
    "reasoning": "Explanation of what was extracted and what needs crawling"
}`, entity, columnsText, organicResults, peopleAlsoAsk, maxURLs)
}

func (g *GroqDecisionMaker) parseResponse(content string, serp *models.GoogleSearchResults, columnsMetadata []*models.ColumnMetadata) (*CrawlDecision, error) {
	content = cleanJSONMarkdown(content)

	var decision CrawlDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		fallbackURLs := []string{}
		for i, result := range serp.Organic {
			if i >= 3 {
				break
			}
			if result.Link != nil {
				fallbackURLs = append(fallbackURLs, *result.Link)
			}
		}

		return &CrawlDecision{
			URLsToCrawl:   fallbackURLs,
			ExtractedData: nil,
			Reasoning:     fmt.Sprintf("Failed to parse LLM response: %s. Falling back to top URLs.", content[:min(100, len(content))]),
		}, nil
	}

	decision.MissingColumns = g.getMissingColumns(decision.ExtractedData, columnsMetadata)

	return &decision, nil
}

func (g *GroqDecisionMaker) getMissingColumns(extractedData map[string]interface{}, columnsMetadata []*models.ColumnMetadata) []string {
	if extractedData == nil {
		missing := make([]string, len(columnsMetadata))
		for i, col := range columnsMetadata {
			missing[i] = col.Name
		}
		return missing
	}

	var missing []string
	for _, col := range columnsMetadata {
		if val, ok := extractedData[col.Name]; !ok || val == nil {
			missing = append(missing, col.Name)
		}
	}

	return missing
}

func cleanJSONMarkdown(content string) string {
	content = strings.TrimSpace(content)

	if !strings.Contains(content, "```") {
		return content
	}

	parts := strings.Split(content, "```")
	if len(parts) < 2 {
		return content
	}

	inner := parts[1]

	if strings.HasPrefix(inner, "json") {
		inner = inner[4:]
	} else if strings.HasPrefix(inner, "JSON") {
		inner = inner[4:]
	}

	return strings.TrimSpace(inner)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
