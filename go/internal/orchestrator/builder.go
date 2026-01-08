package orchestrator

import (
	"fmt"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/feedback"
)

type FeedbackBuilder interface {
	Build(history *feedback.AttemptHistory, weakColumns []feedback.WeakColumn) *feedback.EnrichmentFeedback
}

type DefaultFeedbackBuilder struct{}

func NewDefaultFeedbackBuilder() *DefaultFeedbackBuilder {
	return &DefaultFeedbackBuilder{}
}

func (b *DefaultFeedbackBuilder) Build(history *feedback.AttemptHistory, weakColumns []feedback.WeakColumn) *feedback.EnrichmentFeedback {
	if history == nil || history.Count() == 0 {
		return &feedback.EnrichmentFeedback{
			AttemptNumber: 1,
		}
	}

	fb := &feedback.EnrichmentFeedback{
		AttemptNumber:    history.Count() + 1,
		FocusColumns:     make([]string, 0, len(weakColumns)),
		AvoidPatterns:    history.AllPatternsUsed(),
		AvoidURLs:        history.AllURLsCrawled(),
		PreviousAttempts: make([]feedback.AttemptSummary, 0, len(history.Attempts)),
		Hints:            make([]string, 0),
	}

	for _, wc := range weakColumns {
		fb.FocusColumns = append(fb.FocusColumns, wc.Name)
	}

	for _, attempt := range history.Attempts {
		summary := feedback.AttemptSummary{
			Number:       attempt.Number,
			Patterns:     attempt.PatternsUsed,
			URLsCrawled:  attempt.URLsCrawled,
			WeakColumns:  make([]feedback.WeakColumn, 0),
			Improvements: make([]string, 0),
		}

		if attempt.Assessment != nil {
			summary.WeakColumns = attempt.Assessment.WeakColumns
		}

		fb.PreviousAttempts = append(fb.PreviousAttempts, summary)
	}

	fb.Hints = b.generateHints(history, weakColumns)

	return fb
}

func (b *DefaultFeedbackBuilder) generateHints(history *feedback.AttemptHistory, weakColumns []feedback.WeakColumn) []string {
	hints := make([]string, 0)

	if len(weakColumns) > 0 {
		colNames := make([]string, len(weakColumns))
		for i, wc := range weakColumns {
			colNames[i] = wc.Name
		}
		hints = append(hints, fmt.Sprintf("Focus specifically on finding data for: %s", strings.Join(colNames, ", ")))
	}

	if len(history.AllPatternsUsed()) > 0 {
		hints = append(hints, "Previous search patterns did not yield sufficient results. Try different keyword combinations.")
	}

	if len(history.AllURLsCrawled()) > 0 {
		hints = append(hints, "Previously crawled URLs did not contain the needed information. Look for alternative sources.")
	}

	return hints
}

func (b *DefaultFeedbackBuilder) FormatForPrompt(fb *feedback.EnrichmentFeedback) string {
	if fb == nil || fb.AttemptNumber <= 1 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[CONTEXT: RETRY ATTEMPT %d]\n\n", fb.AttemptNumber))

	if len(fb.PreviousAttempts) > 0 {
		sb.WriteString("PREVIOUS ATTEMPTS:\n")
		for _, attempt := range fb.PreviousAttempts {
			sb.WriteString(fmt.Sprintf("- Attempt %d:\n", attempt.Number))
			if len(attempt.Patterns) > 0 {
				sb.WriteString(fmt.Sprintf("  Patterns: %v\n", attempt.Patterns))
			}
			if len(attempt.WeakColumns) > 0 {
				weakNames := make([]string, len(attempt.WeakColumns))
				for i, wc := range attempt.WeakColumns {
					weakNames[i] = fmt.Sprintf("%s (%.2f)", wc.Name, wc.Confidence)
				}
				sb.WriteString(fmt.Sprintf("  Low confidence columns: %s\n", strings.Join(weakNames, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	if len(fb.FocusColumns) > 0 {
		sb.WriteString(fmt.Sprintf("FOCUS ON THESE COLUMNS: %s\n", strings.Join(fb.FocusColumns, ", ")))
	}

	if len(fb.AvoidPatterns) > 0 {
		sb.WriteString(fmt.Sprintf("AVOID SIMILAR PATTERNS TO: %v\n", fb.AvoidPatterns))
	}

	if len(fb.Hints) > 0 {
		sb.WriteString("\nHINTS:\n")
		for _, hint := range fb.Hints {
			sb.WriteString(fmt.Sprintf("- %s\n", hint))
		}
	}

	return sb.String()
}
