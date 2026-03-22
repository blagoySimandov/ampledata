package services

import (
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services/prompts"
)

type IPromptService interface {
	ExtractionPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, content string) string
	DecisionMakerPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, serp *models.GoogleSearchResults, maxURLs int, previouslyCrawledURLs []string) string
	QueryPatternPrompt(columns []*models.ColumnMetadata) string
	QueryPatternWithFeedbackPrompt(columns []*models.ColumnMetadata, previousAttempts []*models.EnrichmentAttempt) string
	KeySelectorPrompt(headers []string, columns []*models.ColumnMetadata) string
}

type PromptService struct{}

func NewPromptService() *PromptService {
	return &PromptService{}
}

func (p *PromptService) ExtractionPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, content string) string {
	return renderPrompt(prompts.Extraction, map[string]string{
		"entity_context": formatEntityContext(entity, keyDescription),
		"columns":        columnsText(columns),
		"content":        truncateContent(content),
	})
}

func (p *PromptService) DecisionMakerPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, serp *models.GoogleSearchResults, maxURLs int, previouslyCrawledURLs []string) string {
	return renderPrompt(prompts.DecisionMaker, map[string]string{
		"entity_context":         formatEntityContext(entity, keyDescription),
		"entity":                 entity,
		"columns":                columnsText(columns),
		"search_results":         searchResultsText(serp),
		"people_also_ask":        peopleAlsoAskText(serp),
		"max_urls":               fmt.Sprintf("%d", maxURLs),
		"previously_crawled_urls": previouslyCrawledURLsText(previouslyCrawledURLs),
	})
}

func previouslyCrawledURLsText(urls []string) string {
	if len(urls) == 0 {
		return "None"
	}
	var sb strings.Builder
	for _, url := range urls {
		fmt.Fprintf(&sb, "- %s\n", url)
	}
	return sb.String()
}

func (p *PromptService) QueryPatternPrompt(columns []*models.ColumnMetadata) string {
	return renderPrompt(prompts.QueryPattern, map[string]string{
		"column_count": fmt.Sprintf("%d", len(columns)),
		"columns":      columnsText(columns),
		"feedback":     "",
	})
}

func (p *PromptService) QueryPatternWithFeedbackPrompt(columns []*models.ColumnMetadata, previousAttempts []*models.EnrichmentAttempt) string {
	return renderPrompt(prompts.QueryPattern, map[string]string{
		"column_count": fmt.Sprintf("%d", len(columns)),
		"columns":      columnsText(columns),
		"feedback":     feedbackText(previousAttempts),
	})
}

func (p *PromptService) KeySelectorPrompt(headers []string, columns []*models.ColumnMetadata) string {
	return renderPrompt(prompts.KeySelector, map[string]string{
		"headers":      strings.Join(headers, "\n"),
		"columns_info": keySelectorColumnsInfo(columns),
	})
}

func formatEntityContext(entity, keyColumnDescription string) string {
	if keyColumnDescription == "" {
		return fmt.Sprintf(`"%s"`, entity)
	}
	return fmt.Sprintf(`"%s" (context: %s)`, entity, keyColumnDescription)
}

func renderPrompt(tmpl string, vars map[string]string) string {
	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, "{{"+k+"}}", v)
	}
	return strings.NewReplacer(pairs...).Replace(tmpl)
}

func columnsText(columns []*models.ColumnMetadata) string {
	lines := make([]string, 0, len(columns))
	for _, col := range columns {
		line := fmt.Sprintf("- %s [type: %s]", col.Name, col.Type)
		if col.Description != nil {
			line += fmt.Sprintf(" (%s)", *col.Description)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func searchResultsText(serp *models.GoogleSearchResults) string {
	var sb strings.Builder
	for i, r := range serp.Organic {
		if i >= 10 {
			break
		}
		pos := i + 1
		if r.Position != nil {
			pos = *r.Position
		}
		fmt.Fprintf(&sb, "\nPosition %d: %s\nURL: %s\nSnippet: %s\n---",
			pos, Deref(r.Title), Deref(r.Link), Deref(r.Snippet))
	}
	return sb.String()
}

func peopleAlsoAskText(serp *models.GoogleSearchResults) string {
	var sb strings.Builder
	for i, item := range serp.PeopleAlsoAsk {
		if i >= 3 {
			break
		}
		fmt.Fprintf(&sb, "Q: %s A: %s\n", item.Question, item.Snippet)
	}
	return sb.String()
}

func keySelectorColumnsInfo(columns []*models.ColumnMetadata) string {
	if len(columns) == 0 {
		return ""
	}
	lines := make([]string, 0, len(columns))
	for _, col := range columns {
		line := fmt.Sprintf("  - %s [type: %s]", col.Name, col.Type)
		if col.Description != nil {
			line += fmt.Sprintf(" - %s", *col.Description)
		}
		lines = append(lines, line)
	}
	return "\nColumn Metadata (columns to be enriched):\n" + strings.Join(lines, "\n")
}

func truncateContent(s string) string {
	if len(s) > 8000 {
		return s[:8000]
	}
	return s
}

func feedbackText(previousAttempts []*models.EnrichmentAttempt) string {
	if len(previousAttempts) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\nPREVIOUS ATTEMPTS THAT FAILED:\n")
	for _, attempt := range previousAttempts {
		fmt.Fprintf(&sb, "\nAttempt %d:\n", attempt.AttemptNumber)
		fmt.Fprintf(&sb, "  Patterns tried: %s\n", strings.Join(attempt.QueryPatterns, ", "))
		if len(attempt.MissingColumns) > 0 {
			fmt.Fprintf(&sb, "  Missing columns: %s\n", strings.Join(attempt.MissingColumns, ", "))
		}
		if len(attempt.LowConfidenceColumns) > 0 {
			fmt.Fprintf(&sb, "  Low confidence columns: %s\n", strings.Join(attempt.LowConfidenceColumns, ", "))
		}
	}
	sb.WriteString("\nYOUR TASK: Generate DIFFERENT query patterns that will help find the missing or low-confidence columns. Try alternative phrasings and search angles.\n")
	return sb.String()
}
