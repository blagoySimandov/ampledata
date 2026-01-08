package orchestrator

import (
	"github.com/blagoySimandov/ampledata/go/internal/feedback"
	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type QualityEvaluator interface {
	Evaluate(result *models.EnrichmentResult, threshold float64) *feedback.QualityAssessment
}

type DefaultQualityEvaluator struct{}

func NewDefaultQualityEvaluator() *DefaultQualityEvaluator {
	return &DefaultQualityEvaluator{}
}

func (e *DefaultQualityEvaluator) Evaluate(result *models.EnrichmentResult, threshold float64) *feedback.QualityAssessment {
	assessment := &feedback.QualityAssessment{
		Passed:      true,
		WeakColumns: make([]feedback.WeakColumn, 0),
		Suggestions: make([]string, 0),
	}

	if result == nil {
		assessment.Passed = false
		assessment.Suggestions = append(assessment.Suggestions, "No result provided")
		return assessment
	}

	if result.Error != nil {
		assessment.Passed = false
		assessment.Suggestions = append(assessment.Suggestions, "Result contains error: "+*result.Error)
		return assessment
	}

	for colName, confInfo := range result.Confidence {
		if confInfo == nil {
			assessment.WeakColumns = append(assessment.WeakColumns, feedback.WeakColumn{
				Name:       colName,
				Confidence: 0,
				Reason:     "No confidence information",
			})
			continue
		}

		if confInfo.Score < threshold {
			assessment.WeakColumns = append(assessment.WeakColumns, feedback.WeakColumn{
				Name:       colName,
				Confidence: confInfo.Score,
				Reason:     confInfo.Reason,
			})
		}
	}

	if len(assessment.WeakColumns) > 0 {
		assessment.Passed = false
		for _, wc := range assessment.WeakColumns {
			assessment.Suggestions = append(assessment.Suggestions,
				"Column '"+wc.Name+"' has low confidence ("+formatFloat(wc.Confidence)+"): "+wc.Reason)
		}
	}

	return assessment
}

func formatFloat(f float64) string {
	return string(rune('0'+int(f*10)/10)) + "." + string(rune('0'+int(f*10)%10))
}

type ColumnTargetedEvaluator struct {
	targetColumns map[string]struct{}
}

func NewColumnTargetedEvaluator(columns []string) *ColumnTargetedEvaluator {
	targets := make(map[string]struct{})
	for _, col := range columns {
		targets[col] = struct{}{}
	}
	return &ColumnTargetedEvaluator{targetColumns: targets}
}

func (e *ColumnTargetedEvaluator) Evaluate(result *models.EnrichmentResult, threshold float64) *feedback.QualityAssessment {
	assessment := &feedback.QualityAssessment{
		Passed:      true,
		WeakColumns: make([]feedback.WeakColumn, 0),
		Suggestions: make([]string, 0),
	}

	if result == nil {
		assessment.Passed = false
		assessment.Suggestions = append(assessment.Suggestions, "No result provided")
		return assessment
	}

	for colName := range e.targetColumns {
		confInfo, exists := result.Confidence[colName]
		if !exists || confInfo == nil {
			assessment.WeakColumns = append(assessment.WeakColumns, feedback.WeakColumn{
				Name:       colName,
				Confidence: 0,
				Reason:     "No confidence information for targeted column",
			})
			continue
		}

		if confInfo.Score < threshold {
			assessment.WeakColumns = append(assessment.WeakColumns, feedback.WeakColumn{
				Name:       colName,
				Confidence: confInfo.Score,
				Reason:     confInfo.Reason,
			})
		}
	}

	if len(assessment.WeakColumns) > 0 {
		assessment.Passed = false
	}

	return assessment
}
