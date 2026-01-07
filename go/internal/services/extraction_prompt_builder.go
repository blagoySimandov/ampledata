package services

import (
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type IExtractionPromptBuilder interface {
	Build(content string, columnsMetadata []*models.ColumnMetadata, entity string) string
}

type ExtractionPromptBuilder struct{}

func NewExtractionPromptBuilder() *ExtractionPromptBuilder {
	return &ExtractionPromptBuilder{}
}

func (e *ExtractionPromptBuilder) Build(content string, columnsMetadata []*models.ColumnMetadata, entity string) string {
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

CRITICAL: NEVER infer, estimate, or make up data. Only extract information that is explicitly stated in the content.
- If you see "10000+" do NOT convert it to "10001" - use the exact value or omit if uncertain
- If information is approximate, partial, or unclear, reflect that uncertainty in the confidence score

IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata:
- For number types: use numeric values without quotes (e.g., 1000)
- For string types: use quoted strings
- For boolean types: use true/false without quotes
- For date types: use ISO 8601 format (YYYY-MM-DD)

If a field cannot be found in the content, omit it from the response.

## Response Format (JSON only, no markdown)
{
    "extracted_data": {"field_name": value_with_correct_type},
    "confidence": {
        "field_name": {
            "score": 0.95,
            "reason": "Brief 1-sentence explanation"
        }
    },
    "reasoning": "Overall extraction summary"
}

## Confidence Scoring Guidelines
- 1.0: Exact match, explicitly stated
- 0.8-0.9: Clear statement, minor interpretation needed
- 0.6-0.7: Partial information or context-based inference
- 0.4-0.5: Significant uncertainty or approximation
- <0.4: High uncertainty, possibly derived or estimated`, entity, columnsText, truncatedContent)
}
