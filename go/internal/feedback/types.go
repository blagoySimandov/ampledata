package feedback

type EnrichmentFeedback struct {
	AttemptNumber    int
	FocusColumns     []string
	AvoidPatterns    []string
	AvoidURLs        []string
	PreviousAttempts []AttemptSummary
	Hints            []string
}

type AttemptSummary struct {
	Number       int
	Patterns     []string
	URLsCrawled  []string
	WeakColumns  []WeakColumn
	Improvements []string
}

type WeakColumn struct {
	Name       string
	Confidence float64
	Reason     string
}

type QualityAssessment struct {
	Passed      bool
	WeakColumns []WeakColumn
	Suggestions []string
}

func (f *EnrichmentFeedback) IsRetry() bool {
	return f != nil && f.AttemptNumber > 1
}

func (f *EnrichmentFeedback) GetFocusColumnNames() []string {
	if f == nil {
		return nil
	}
	return f.FocusColumns
}

func (a *QualityAssessment) HasWeakColumns() bool {
	return len(a.WeakColumns) > 0
}

func (a *QualityAssessment) GetWeakColumnNames() []string {
	names := make([]string, len(a.WeakColumns))
	for i, wc := range a.WeakColumns {
		names[i] = wc.Name
	}
	return names
}
