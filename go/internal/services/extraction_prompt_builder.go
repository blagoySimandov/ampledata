package services

import (
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type IExtractionPromptBuilder interface {
	Build(content string, columnsMetadata []*models.ColumnMetadata, entity string, entityType string) string
}

type ExtractionPromptBuilder struct{}

func NewExtractionPromptBuilder() *ExtractionPromptBuilder {
	return &ExtractionPromptBuilder{}
}

func (e *ExtractionPromptBuilder) Build(content string, columnsMetadata []*models.ColumnMetadata, entity string, entityType string) string {
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

	return fmt.Sprintf(`You are a data extraction specialist. Extract the following fields about the %s named "%s" from the provided website content.

## CRITICAL: Entity Extraction Rules

You are extracting data about the TARGET ENTITY: %s "%s"

ALL extracted fields must be about THIS SPECIFIC ENTITY - not about related or mentioned entities.

Common entity disambiguation scenarios:
- If extracting about a company: data should be about the COMPANY, not its founders/executives/employees
- If extracting about a person: data should be about THIS PERSON, not their company/colleagues/family
- If extracting about a product: data should be about the PRODUCT, not the manufacturer or similar products
- If extracting about a location: data should be about THIS LOCATION, not nearby places or regions

When content mentions MULTIPLE entities:
- ✓ Extract ONLY data that clearly applies to the target entity "%s"
- ✗ Do NOT extract data about related/mentioned entities
- ✗ Do NOT mix attributes from different entities

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

## Entity Consistency Validation

Before finalizing your response, verify:
1. ALL extracted data refers to the TARGET ENTITY (%s: "%s"), not to:
   - Related or associated entities mentioned in the content
   - Similar or competing entities
   - Parent/subsidiary entities (unless explicitly requested)

2. If you find data about a DIFFERENT entity:
   - Do NOT include that field in extracted_data
   - REDUCE confidence score to 0.0-0.3 if uncertain which entity it applies to
   - Explain the ambiguity in your reasoning

3. Cross-field consistency check:
   - Verify all extracted fields logically apply to the SAME entity
   - If fields seem contradictory or from different entities, investigate before extracting
   - When in doubt, omit the ambiguous field rather than extract incorrect data

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

Base your confidence on BOTH information availability AND entity certainty:

- 1.0: Exact match, explicitly stated, AND clearly about the target entity
  * The information is unambiguous and directly attributed to "%s"

- 0.8-0.9: Clear statement, minor interpretation needed, target entity is clear
  * Strong attribution to the target entity with minimal ambiguity

- 0.6-0.7: Partial information, OR content mentions multiple entities
  * Information exists but requires context or interpretation
  * Multiple entities mentioned but target entity can be distinguished

- 0.4-0.5: Significant uncertainty, entity ambiguity, or approximation
  * Unclear which entity the information applies to
  * Information is approximate or estimated

- <0.4: High uncertainty, likely derived or estimated, or wrong entity suspected
  * Information may be about a different but related entity
  * Heavy inference required

⚠️  ALWAYS reduce confidence by at least 0.2 when:
- Information is about a related entity instead of the target entity "%s"
- Multiple entities are mentioned and target is unclear
- Source is indirect (third-party descriptions, not primary source)`, entityType, entity, entityType, entity, entityType, entityType, entityType, entityType, columnsText, truncatedContent, entity, entityType)
}
