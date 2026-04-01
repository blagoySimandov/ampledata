package services

import (
	"google.golang.org/genai"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

// BuildQueryPatternSchema returns a schema that constrains the pattern-generator
// response to a plain JSON array of strings, eliminating parse failures.
func BuildQueryPatternSchema() *genai.Schema {
	return &genai.Schema{
		Type:  genai.TypeArray,
		Items: &genai.Schema{Type: genai.TypeString},
	}
}

// BuildDecisionMakerSchema builds a response schema for the decision-maker LLM
// call. The extracted_data and extracted_confidence properties are keyed by the
// actual column names so the model cannot invent or omit fields.
func BuildDecisionMakerSchema(columns []*models.ColumnMetadata) *genai.Schema {
	extractedDataProps := make(map[string]*genai.Schema, len(columns))
	confidenceProps := make(map[string]*genai.Schema, len(columns))
	for _, col := range columns {
		extractedDataProps[col.Name] = columnTypeToSchema(col.Type)
		confidenceProps[col.Name] = confidenceFieldSchema()
	}

	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"urls_to_crawl": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"extracted_data": {
				Type:       genai.TypeObject,
				Properties: extractedDataProps,
				Nullable:   genai.Ptr(true),
			},
			"extracted_confidence": {
				Type:       genai.TypeObject,
				Properties: confidenceProps,
			},
			"source_urls": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"reasoning": {Type: genai.TypeString},
		},
		Required: []string{"urls_to_crawl", "source_urls", "reasoning"},
	}
}

// BuildExtractionSchema builds a response schema for the content-extractor LLM
// call. Every column that was requested must appear in both extracted_data and
// confidence, even when the value is null.
func BuildExtractionSchema(columns []*models.ColumnMetadata) *genai.Schema {
	extractedDataProps := make(map[string]*genai.Schema, len(columns))
	confidenceProps := make(map[string]*genai.Schema, len(columns))
	for _, col := range columns {
		extractedDataProps[col.Name] = columnTypeToSchema(col.Type)
		confidenceProps[col.Name] = confidenceFieldSchema()
	}

	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"extracted_data": {
				Type:       genai.TypeObject,
				Properties: extractedDataProps,
			},
			"confidence": {
				Type:       genai.TypeObject,
				Properties: confidenceProps,
			},
			"reasoning": {Type: genai.TypeString},
		},
		Required: []string{"extracted_data", "confidence", "reasoning"},
	}
}

// columnTypeToSchema maps a ColumnType to the matching genai scalar schema.
// All fields are nullable so the model can return null when data is absent.
func columnTypeToSchema(t models.ColumnType) *genai.Schema {
	switch t {
	case models.ColumnTypeNumber:
		return &genai.Schema{Type: genai.TypeNumber, Nullable: true}
	case models.ColumnTypeBoolean:
		return &genai.Schema{Type: genai.TypeBoolean, Nullable: true}
	default: // string and date both serialise as strings
		return &genai.Schema{Type: genai.TypeString, Nullable: true}
	}
}

// confidenceFieldSchema returns the schema for a single confidence entry:
// {"score": 0.9, "reason": "..."}.
func confidenceFieldSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"score":  {Type: genai.TypeNumber},
			"reason": {Type: genai.TypeString},
		},
		Required: []string{"score", "reason"},
	}
}
