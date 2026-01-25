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

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"google.golang.org/genai"
)

type CrawlDecision struct {
	URLsToCrawl    []string               `json:"urls_to_crawl"`
	ExtractedData  map[string]interface{} `json:"extracted_data"`
	Reasoning      string                 `json:"reasoning"`
	MissingColumns []string               `json:"-"`
}

type DecisionMaker interface {
	MakeDecision(ctx context.Context, serp *models.GoogleSearchResults, rowKey string, maxURLs int, columnsMetadata []*models.ColumnMetadata, entityType string) (*CrawlDecision, error)
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

func (g *GroqDecisionMaker) MakeDecision(ctx context.Context, serp *models.GoogleSearchResults, rowKey string, maxURLs int, columnsMetadata []*models.ColumnMetadata, entityType string) (*CrawlDecision, error) {
	prompt := g.buildPrompt(serp, rowKey, maxURLs, columnsMetadata, entityType)

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

func (g *GroqDecisionMaker) buildPrompt(serp *models.GoogleSearchResults, entity string, maxURLs int, columnsMetadata []*models.ColumnMetadata, entityType string) string {
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

	return fmt.Sprintf(`You are a data extraction assistant. Analyze these search results for the %s "%s" and decide how to proceed.

## CRITICAL: Entity Extraction Rules

You are analyzing results for TARGET ENTITY: %s "%s"

ALL extracted data must be about THIS SPECIFIC ENTITY - not about related or mentioned entities.

When search results contain information about MULTIPLE entities:
- ✓ Extract ONLY data clearly about the target entity "%s"
- ✗ Do NOT extract data about related/mentioned entities
- ✗ Do NOT select URLs primarily about different entities for crawling

## Columns We Need to Extract
%s

## Search Results
%s

## People Also Ask
%s

## Your Task

1. Extract data from snippets that is about "%s" (%s):
   - CRITICAL: Verify each piece of data is about the TARGET ENTITY, not related entities
   - IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata
   - For number types: use numeric values without quotes (e.g., 1000, 228000)
   - For string types: use quoted strings
   - For boolean types: use true/false without quotes
   - For date types: use ISO 8601 format (YYYY-MM-DD)
   - If unsure whether data applies to target entity, do NOT extract it

2. For columns you CANNOT extract from snippets:
   - ALWAYS select URLs to crawl that are likely to contain the missing data
   - Even if snippets don't contain the exact value, if they reference or link to where the data exists, SELECT THOSE URLs
   - Select up to %d URLs to crawl for missing data, prioritizing:
     * Official or authoritative sources about the TARGET ENTITY specifically
     * URLs whose titles/snippets clearly reference the target entity "%s"
     * Wikipedia pages specifically about the target entity
     * Reliable data sources (official sites, registries, databases)
     * Image hosting sites (Getty, Shutterstock, etc.) for image URLs
     * Avoid: URLs primarily about related entities, SEO aggregators

   ⚠️  CRITICAL:
     - If you cannot extract a column's data from snippets BUT the search results contain relevant URLs, YOU MUST SELECT URLs TO CRAWL
     - Do NOT return empty urls_to_crawl when relevant URLs exist in the results
     - Verify URL titles and snippets are about "%s" (%s), not related entities
     - Skip URLs that focus on different entities, even if they mention the target

## Entity Consistency Check

Before responding:
1. Review ALL extracted data - does it ALL refer to the same entity ("%s")?
2. Review ALL selected URLs - are they primarily about "%s" (%s)?
3. If you find mixed entity data, extract ONLY the data about the target entity
4. In your reasoning, note any entity ambiguity you encountered

## Examples

Example 1 - Data visible in snippets:
- Column needed: employee_count (number)
- Snippet: "Microsoft had 228,000 employees as of June 2025"
- Response: {"urls_to_crawl": [], "extracted_data": {"employee_count": 228000}, "reasoning": "Extracted employee count directly from snippet"}

Example 2 - Data not in snippets but URLs available:
- Column needed: founder_picture_url (string)
- Snippets: "Getty Images has photos of Microsoft founder", "Shutterstock Microsoft founder images"
- Response: {"urls_to_crawl": ["https://www.gettyimages.com/photos/microsoft-founder", "https://news.microsoft.com/..."], "extracted_data": null, "reasoning": "Cannot extract image URL from snippets, but Getty Images and official Microsoft site likely contain founder photos"}

Example 3 - Mixed scenario:
- Columns needed: employee_count, founder_picture_url
- Snippets show employee count but not picture URL
- Response: {"urls_to_crawl": ["url1", "url2"], "extracted_data": {"employee_count": 228000}, "reasoning": "Extracted employee count, selecting URLs to find founder picture"}

## Response Format (JSON only, no markdown)
{
    "urls_to_crawl": ["url1", "url2"] or [],
    "extracted_data": {"column_name": value_with_correct_type} or null,
    "reasoning": "Explanation of what was extracted, what needs crawling, and any entity disambiguation performed"
}`, entityType, entity, entityType, entity, entity, columnsText, organicResults, peopleAlsoAsk, entity, entityType, maxURLs, entity, entity, entityType, entity, entity, entityType)
}

func (g *GroqDecisionMaker) parseResponse(content string, serp *models.GoogleSearchResults, columnsMetadata []*models.ColumnMetadata) (*CrawlDecision, error) {
	content = cleanJSONMarkdown(content)

	cfg := config.Load()
	var decision CrawlDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		fallbackURLs := []string{}
		for i, result := range serp.Organic {
			if i >= cfg.MaxOrganicResults {
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

// GeminiDecisionMaker implements DecisionMaker using Google's Gemini API
type GeminiDecisionMaker struct {
	model  string
	client *genai.Client
}

func NewGeminiDecisionMaker(apiKey string) (*GeminiDecisionMaker, error) {
	ctx := context.Background() // ctx used for auth/initialization, not passed to later requests
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return &GeminiDecisionMaker{
		model:  "gemini-2.5-flash-preview-09-2025",
		client: client,
	}, nil
}

func (g *GeminiDecisionMaker) MakeDecision(ctx context.Context, serp *models.GoogleSearchResults, rowKey string, maxURLs int, columnsMetadata []*models.ColumnMetadata, entityType string) (*CrawlDecision, error) {
	prompt := g.buildPrompt(serp, rowKey, maxURLs, columnsMetadata, entityType)

	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return g.parseResponse(result.Text(), serp, columnsMetadata)
}

func (g *GeminiDecisionMaker) buildPrompt(serp *models.GoogleSearchResults, entity string, maxURLs int, columnsMetadata []*models.ColumnMetadata, entityType string) string {
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

	return fmt.Sprintf(`You are a data extraction assistant. Analyze these search results for the %s "%s" and decide how to proceed.

## CRITICAL: Entity Extraction Rules

You are analyzing results for TARGET ENTITY: %s "%s"

ALL extracted data must be about THIS SPECIFIC ENTITY - not about related or mentioned entities.

When search results contain information about MULTIPLE entities:
- ✓ Extract ONLY data clearly about the target entity "%s"
- ✗ Do NOT extract data about related/mentioned entities
- ✗ Do NOT select URLs primarily about different entities for crawling

## Columns We Need to Extract
%s

## Search Results
%s

## People Also Ask
%s

## Your Task

1. Extract data from snippets that is about "%s" (%s):
   - CRITICAL: Verify each piece of data is about the TARGET ENTITY, not related entities
   - IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata
   - For number types: use numeric values without quotes (e.g., 1000, 228000)
   - For string types: use quoted strings
   - For boolean types: use true/false without quotes
   - For date types: use ISO 8601 format (YYYY-MM-DD)
   - If unsure whether data applies to target entity, do NOT extract it

2. For columns you CANNOT extract from snippets:
   - ALWAYS select URLs to crawl that are likely to contain the missing data
   - Even if snippets don't contain the exact value, if they reference or link to where the data exists, SELECT THOSE URLs
   - Select up to %d URLs to crawl for missing data, prioritizing:
     * Official or authoritative sources about the TARGET ENTITY specifically
     * URLs whose titles/snippets clearly reference the target entity "%s"
     * Wikipedia pages specifically about the target entity
     * Reliable data sources (official sites, registries, databases)
     * Image hosting sites (Getty, Shutterstock, etc.) for image URLs
     * Avoid: URLs primarily about related entities, SEO aggregators

   ⚠️  CRITICAL:
     - If you cannot extract a column's data from snippets BUT the search results contain relevant URLs, YOU MUST SELECT URLs TO CRAWL
     - Do NOT return empty urls_to_crawl when relevant URLs exist in the results
     - Verify URL titles and snippets are about "%s" (%s), not related entities
     - Skip URLs that focus on different entities, even if they mention the target

## Entity Consistency Check

Before responding:
1. Review ALL extracted data - does it ALL refer to the same entity ("%s")?
2. Review ALL selected URLs - are they primarily about "%s" (%s)?
3. If you find mixed entity data, extract ONLY the data about the target entity
4. In your reasoning, note any entity ambiguity you encountered

## Examples

Example 1 - Data visible in snippets:
- Column needed: employee_count (number)
- Snippet: "Microsoft had 228,000 employees as of June 2025"
- Response: {"urls_to_crawl": [], "extracted_data": {"employee_count": 228000}, "reasoning": "Extracted employee count directly from snippet"}

Example 2 - Data not in snippets but URLs available:
- Column needed: founder_picture_url (string)
- Snippets: "Getty Images has photos of Microsoft founder", "Shutterstock Microsoft founder images"
- Response: {"urls_to_crawl": ["https://www.gettyimages.com/photos/microsoft-founder", "https://news.microsoft.com/..."], "extracted_data": null, "reasoning": "Cannot extract image URL from snippets, but Getty Images and official Microsoft site likely contain founder photos"}

Example 3 - Mixed scenario:
- Columns needed: employee_count, founder_picture_url
- Snippets show employee count but not picture URL
- Response: {"urls_to_crawl": ["url1", "url2"], "extracted_data": {"employee_count": 228000}, "reasoning": "Extracted employee count, selecting URLs to find founder picture"}

## Response Format (JSON only, no markdown)
{
    "urls_to_crawl": ["url1", "url2"] or [],
    "extracted_data": {"column_name": value_with_correct_type} or null,
    "reasoning": "Explanation of what was extracted, what needs crawling, and any entity disambiguation performed"
}`, entityType, entity, entityType, entity, entity, columnsText, organicResults, peopleAlsoAsk, entity, entityType, maxURLs, entity, entity, entityType, entity, entity, entityType)
}

func (g *GeminiDecisionMaker) parseResponse(content string, serp *models.GoogleSearchResults, columnsMetadata []*models.ColumnMetadata) (*CrawlDecision, error) {
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

func (g *GeminiDecisionMaker) getMissingColumns(extractedData map[string]interface{}, columnsMetadata []*models.ColumnMetadata) []string {
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
