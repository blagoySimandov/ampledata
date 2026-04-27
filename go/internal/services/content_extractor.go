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

	var opts []GenerateOption
	if g.crawler != nil {
		opts = append(opts, WithTools([]Tool{NewFetchPageTool(g.crawler)}, maxToolSteps))
	}
	result, err := g.client.GenerateContent(ctx, prompt, opts...)

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
