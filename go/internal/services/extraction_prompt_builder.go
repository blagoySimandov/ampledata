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
