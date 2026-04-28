package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type CrawlDecision struct {
	URLsToCrawl    []string                               `json:"urls_to_crawl"`
	ExtractedData  map[string]interface{}                 `json:"extracted_data"`
	Confidence     map[string]*models.FieldConfidenceInfo `json:"extracted_confidence"`
	Reasoning      string                                 `json:"reasoning"`
	SourceURLs     []string                               `json:"source_urls"`
	MissingColumns []string                               `json:"-"`
}

type AIDecisionMaker struct {
	client        IAIClient
	promptService IPromptService
}

func NewGeminiDecisionMaker(promptService IPromptService, client IAIClient) (*AIDecisionMaker, error) {
	return &AIDecisionMaker{
		client,
		promptService,
	}, nil
}

func (g *AIDecisionMaker) MakeDecision(ctx context.Context, serp *models.GoogleSearchResults, rowKey string, maxURLs int, columnsMetadata []*models.ColumnMetadata, keyColumnDescription string, previousAttempts []*models.EnrichmentAttempt) (*CrawlDecision, error) {
	prompt := g.promptService.DecisionMakerPrompt(rowKey, keyColumnDescription, columnsMetadata, serp, maxURLs, previousAttempts)
	result, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return g.parseResponse(result, serp, columnsMetadata)
}

func (g *AIDecisionMaker) parseResponse(content string, serp *models.GoogleSearchResults, columnsMetadata []*models.ColumnMetadata) (*CrawlDecision, error) {
	content = cleanJSONMarkdown(content)

	var decision CrawlDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		fallbackURLs := fallbackSerpURLs(serp)
		return &CrawlDecision{
			URLsToCrawl:   fallbackURLs,
			ExtractedData: nil,
			Reasoning:     fmt.Sprintf("Failed to parse LLM response: %s. Falling back to top URLs.", content[:min(100, len(content))]),
		}, nil
	}

	decision.MissingColumns = getMissingColumns(decision.ExtractedData, columnsMetadata)
	return &decision, nil
}

func fallbackSerpURLs(serp *models.GoogleSearchResults) []string {
	cfg := config.Load()
	var urls []string
	for i, result := range serp.Organic {
		if i >= cfg.MaxOrganicResults {
			break
		}
		if result.Link != nil {
			urls = append(urls, *result.Link)
		}
	}
	return urls
}

func getMissingColumns(extractedData map[string]interface{}, columnsMetadata []*models.ColumnMetadata) []string {
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
