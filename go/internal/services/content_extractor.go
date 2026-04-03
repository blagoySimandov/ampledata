package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type ExtractionResult struct {
	ExtractedData map[string]interface{}                 `json:"extracted_data"`
	Confidence    map[string]*models.FieldConfidenceInfo `json:"confidence"`
	Reasoning     string                                 `json:"reasoning"`
}

type IContentExtractor interface {
	Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription string) (*ExtractionResult, error)
}

type AIContentExtractor struct {
	client        IAIClient
	promptService IPromptService
	crawler       WebCrawler
}

type ContentExtractorOption func(*AIContentExtractor)

func WithCrawler(crawler WebCrawler) ContentExtractorOption {
	return func(e *AIContentExtractor) {
		e.crawler = crawler
	}
}

func NewAIContentExtractor(client IAIClient, promptService IPromptService, opts ...ContentExtractorOption) (*AIContentExtractor, error) {
	e := &AIContentExtractor{
		client:        client,
		promptService: promptService,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e, nil
}

const maxToolSteps = 3

func (g *AIContentExtractor) Extract(ctx context.Context, content string, entityKey string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription string) (*ExtractionResult, error) {
	prompt := g.promptService.ExtractionPrompt(entityKey, keyColumnDescription, columnsMetadata, content)

	var result string
	var err error

	toolClient, hasTools := g.client.(IToolAIClient)
	if g.crawler != nil && hasTools {
		result, err = g.extractWithTools(ctx, toolClient, prompt)
	} else {
		result, err = g.client.GenerateContent(ctx, prompt)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	er, err := parseResponse(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// TODO: make this cleaner in some way.
	coercedData := ValidateAndCoerceTypes(er.ExtractedData, columnsMetadata, er.Confidence)
	er.ExtractedData = coercedData
	return er, nil
}

func (g *AIContentExtractor) extractWithTools(ctx context.Context, client IToolAIClient, prompt string) (string, error) {
	tools := []ToolDefinition{
		{
			Name:        "fetch_page",
			Description: "Fetch the content of a web page by URL. Use this when you see a link in the content that likely contains additional information about the target entity (e.g. an about page, a detailed profile, or a data source referenced in the text).",
			Parameters: []ToolParameter{
				{Name: "url", Type: "STRING", Description: "The URL of the page to fetch"},
			},
			Required: []string{"url"},
		},
	}

	handler := func(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
		if name != "fetch_page" {
			return map[string]any{"error": fmt.Sprintf("unknown tool: %s", name)}, nil
		}
		url, ok := args["url"].(string)
		if !ok {
			return map[string]any{"error": "url parameter must be a string"}, nil
		}
		content, err := g.crawler.Crawl(ctx, []string{url}, "")
		if err != nil {
			return map[string]any{"error": fmt.Sprintf("failed to fetch page: %v", err)}, nil
		}
		return map[string]any{"content": content}, nil
	}

	return client.GenerateContentWithTools(ctx, prompt, tools, handler, maxToolSteps)
}

func parseResponse(content string) (*ExtractionResult, error) {
	content = cleanJSONMarkdown(content)

	var result ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return &ExtractionResult{
			ExtractedData: make(map[string]interface{}),
			Confidence:    make(map[string]*models.FieldConfidenceInfo),
			Reasoning:     fmt.Sprintf("Failed to parse LLM response: %s", content[:min(100, len(content))]),
		}, nil
	}

	if result.Confidence == nil {
		result.Confidence = make(map[string]*models.FieldConfidenceInfo)
	}

	return &result, nil
}
