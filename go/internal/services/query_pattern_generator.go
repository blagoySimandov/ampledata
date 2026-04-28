package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type PatternGenerator struct {
	client        IAIClient
	promptService IPromptService
}

func NewPatternGenerator(aiClient IAIClient, promptService IPromptService) (*PatternGenerator, error) {
	return &PatternGenerator{
		client:        aiClient,
		promptService: promptService,
	}, nil
}

func (g *PatternGenerator) GeneratePatterns(ctx context.Context, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	prompt := g.promptService.QueryPatternPrompt(columnsMetadata)
	result, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to generate content: %w", err)
	}
	return g.parseResponse(result, columnsMetadata)
}

func (g *PatternGenerator) GeneratePatternsWithFeedback(ctx context.Context, columnsMetadata []*models.ColumnMetadata, previousAttempts []*models.EnrichmentAttempt) ([]string, error) {
	prompt := g.promptService.QueryPatternWithFeedbackPrompt(columnsMetadata, previousAttempts)
	result, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return g.getFallbackPatterns(columnsMetadata), fmt.Errorf("failed to generate content: %w", err)
	}
	return g.parseResponse(result, columnsMetadata)
}

func (g *PatternGenerator) parseResponse(content string, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
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

func (g *PatternGenerator) getFallbackPatterns(columnsMetadata []*models.ColumnMetadata) []string {
	parts := []string{"%entity"}
	for _, col := range columnsMetadata {
		parts = append(parts, col.Name)
	}
	return []string{strings.Join(parts, " ")}
}
