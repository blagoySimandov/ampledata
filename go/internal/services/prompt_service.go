package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services/prompts"
)

type IPromptService interface {
	ExtractionPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, content string) string
	DecisionMakerPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, serp *models.GoogleSearchResults, maxURLs int, previousAttempts []*models.EnrichmentAttempt) string
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
		"content":        content, // TODO: maybe truncate the content and have a fetch tool for the agent ?
	})
}

func (p *PromptService) DecisionMakerPrompt(entity, keyDescription string, columns []*models.ColumnMetadata, serp *models.GoogleSearchResults, maxURLs int, previousAttempts []*models.EnrichmentAttempt) string {
	return renderPrompt(prompts.DecisionMaker, map[string]string{
		"entity_context":    formatEntityContext(entity, keyDescription),
		"entity":            entity,
		"columns":           columnsText(columns),
		"search_results":    searchResultsText(serp),
		"people_also_ask":   peopleAlsoAskText(serp),
		"max_urls":          fmt.Sprintf("%d", maxURLs),
		"previous_attempts": previousAttemptsMemory(previousAttempts),
	})
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

func (p *PromptService) GenerateSourceNamePrompt(_ context.Context, headers []string) string {
	return renderPrompt(prompts.SourceName, map[string]string{
		"headers": strings.Join(headers, "\n"),
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

func previousAttemptsMemory(attempts []*models.EnrichmentAttempt) string {
	if len(attempts) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n<memory>\nThe following URLs have already been crawled in previous attempts. Do NOT select any of these URLs again:\n")
	for _, attempt := range attempts {
		fmt.Fprintf(&sb, "\nAttempt %d:\n", attempt.AttemptNumber)
		if len(attempt.CrawledURLs) > 0 {
			sb.WriteString("  Crawled URLs:\n")
			for _, url := range attempt.CrawledURLs {
				fmt.Fprintf(&sb, "    - %s\n", url)
			}
		}
		if len(attempt.MissingColumns) > 0 {
			fmt.Fprintf(&sb, "  Still missing: %s\n", strings.Join(attempt.MissingColumns, ", "))
		}
		if len(attempt.LowConfidenceColumns) > 0 {
			fmt.Fprintf(&sb, "  Low confidence: %s\n", strings.Join(attempt.LowConfidenceColumns, ", "))
		}
	}
	sb.WriteString("\nFocus on finding NEW URLs that may contain the missing or low-confidence data.\n</memory>")
	return sb.String()
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
