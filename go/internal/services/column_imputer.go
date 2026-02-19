package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

// ColumnImputer derives missing target column values from existing source column data using LLM reasoning.
type ColumnImputer interface {
	ImputeColumns(ctx context.Context, rowKey string, sourceData map[string]string, targetColumns []*models.ColumnMetadata) (*ImputationResult, error)
}

// ImputationResult holds LLM-derived field values with confidence scores.
type ImputationResult struct {
	ImputedData map[string]interface{}
	Confidence  map[string]*models.FieldConfidenceInfo
}

// GeminiColumnImputer implements ColumnImputer using the Gemini AI client.
type GeminiColumnImputer struct {
	client IAIClient
}

// NewGeminiColumnImputer creates a new GeminiColumnImputer.
func NewGeminiColumnImputer(client IAIClient) *GeminiColumnImputer {
	return &GeminiColumnImputer{client: client}
}

// ImputeColumns uses an LLM to derive target column values from the provided source data.
// It only fills values that can be reliably inferred; uncertain fields are omitted.
func (g *GeminiColumnImputer) ImputeColumns(ctx context.Context, rowKey string, sourceData map[string]string, targetColumns []*models.ColumnMetadata) (*ImputationResult, error) {
	prompt := buildImputationPrompt(rowKey, sourceData, targetColumns)

	raw, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("imputation LLM call failed: %w", err)
	}

	er, err := parseResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse imputation response: %w", err)
	}

	coerced := ValidateAndCoerceTypes(er.ExtractedData, targetColumns, er.Confidence)

	return &ImputationResult{
		ImputedData: coerced,
		Confidence:  er.Confidence,
	}, nil
}

func buildImputationPrompt(rowKey string, sourceData map[string]string, targetColumns []*models.ColumnMetadata) string {
	var sourceLines []string
	for k, v := range sourceData {
		sourceLines = append(sourceLines, fmt.Sprintf("- %s: %s", k, v))
	}
	sourceText := strings.Join(sourceLines, "\n")

	var targetLines []string
	for _, col := range targetColumns {
		desc := ""
		if col.Description != nil {
			desc = fmt.Sprintf(" (%s)", *col.Description)
		}
		targetLines = append(targetLines, fmt.Sprintf("- %s [type: %s]%s", col.Name, col.Type, desc))
	}
	targetText := strings.Join(targetLines, "\n")

	return fmt.Sprintf(`You are a data imputation specialist. Your task is to derive missing field values for the entity "%s" using ONLY the source data provided below.

## Source Data (already known)
%s

## Target Fields to Derive
%s

## Rules
- ONLY derive a field if it can be reliably inferred from the source data with high confidence.
- Do NOT guess, estimate, or make up values. If a field cannot be derived from the source data, omit it entirely.
- Do NOT use any external knowledge beyond what is provided in the source data.
- Examples of valid derivation: splitting "Full Name" into "First Name"/"Last Name", deriving "Country Code" from "Country", formatting a date to ISO 8601.
- Examples of invalid derivation: guessing revenue from company name, inferring headcount from industry.

## Response Format (JSON only, no markdown)
{
    "extracted_data": {"field_name": value_with_correct_type},
    "confidence": {
        "field_name": {
            "score": 0.95,
            "reason": "Brief 1-sentence explanation of how this was derived"
        }
    },
    "reasoning": "Overall summary of what could and could not be derived"
}

## Confidence Scoring
- 1.0: Directly computable from source data (e.g., splitting a name field)
- 0.8-0.9: Clearly inferable with very high certainty from source data
- 0.6-0.7: Reasonably inferable but with some ambiguity
- < 0.6: Too uncertain — omit this field entirely rather than include it

Omit any field you cannot derive with at least 0.6 confidence.`, rowKey, sourceText, targetText)
}
