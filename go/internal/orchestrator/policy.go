package orchestrator

import (
	"github.com/blagoySimandov/ampledata/go/internal/feedback"
)

type RetryPolicy interface {
	ShouldRetry(attemptCount int, assessment *feedback.QualityAssessment) bool
	MaxAttempts() int
	ConfidenceThreshold() float64
}

type RetryPolicyConfig struct {
	MaxRetries          int
	Threshold           float64
	RequireImprovement  bool
	MinWeakColumnsRetry int
}

func DefaultRetryPolicyConfig() RetryPolicyConfig {
	return RetryPolicyConfig{
		MaxRetries:          3,
		Threshold:           0.6,
		RequireImprovement:  true,
		MinWeakColumnsRetry: 1,
	}
}

type DefaultRetryPolicy struct {
	config              RetryPolicyConfig
	previousWeakCount   int
	previousBestScores  map[string]float64
}

func NewDefaultRetryPolicy(config RetryPolicyConfig) *DefaultRetryPolicy {
	return &DefaultRetryPolicy{
		config:             config,
		previousBestScores: make(map[string]float64),
	}
}

func (p *DefaultRetryPolicy) ShouldRetry(attemptCount int, assessment *feedback.QualityAssessment) bool {
	if assessment == nil || assessment.Passed {
		return false
	}

	if attemptCount >= p.config.MaxRetries {
		return false
	}

	if len(assessment.WeakColumns) < p.config.MinWeakColumnsRetry {
		return false
	}

	if p.config.RequireImprovement && attemptCount > 1 {
		currentWeakCount := len(assessment.WeakColumns)
		if currentWeakCount >= p.previousWeakCount && p.previousWeakCount > 0 {
			improved := false
			for _, wc := range assessment.WeakColumns {
				if prev, ok := p.previousBestScores[wc.Name]; ok {
					if wc.Confidence > prev {
						improved = true
						break
					}
				}
			}
			if !improved {
				return false
			}
		}
	}

	p.previousWeakCount = len(assessment.WeakColumns)
	for _, wc := range assessment.WeakColumns {
		if existing, ok := p.previousBestScores[wc.Name]; !ok || wc.Confidence > existing {
			p.previousBestScores[wc.Name] = wc.Confidence
		}
	}

	return true
}

func (p *DefaultRetryPolicy) MaxAttempts() int {
	return p.config.MaxRetries
}

func (p *DefaultRetryPolicy) ConfidenceThreshold() float64 {
	return p.config.Threshold
}

func (p *DefaultRetryPolicy) Reset() {
	p.previousWeakCount = 0
	p.previousBestScores = make(map[string]float64)
}

type AlwaysRetryPolicy struct {
	maxAttempts int
	threshold   float64
}

func NewAlwaysRetryPolicy(maxAttempts int, threshold float64) *AlwaysRetryPolicy {
	return &AlwaysRetryPolicy{
		maxAttempts: maxAttempts,
		threshold:   threshold,
	}
}

func (p *AlwaysRetryPolicy) ShouldRetry(attemptCount int, assessment *feedback.QualityAssessment) bool {
	if assessment == nil || assessment.Passed {
		return false
	}
	return attemptCount < p.maxAttempts
}

func (p *AlwaysRetryPolicy) MaxAttempts() int {
	return p.maxAttempts
}

func (p *AlwaysRetryPolicy) ConfidenceThreshold() float64 {
	return p.threshold
}

type NeverRetryPolicy struct {
	threshold float64
}

func NewNeverRetryPolicy(threshold float64) *NeverRetryPolicy {
	return &NeverRetryPolicy{threshold: threshold}
}

func (p *NeverRetryPolicy) ShouldRetry(attemptCount int, assessment *feedback.QualityAssessment) bool {
	return false
}

func (p *NeverRetryPolicy) MaxAttempts() int {
	return 1
}

func (p *NeverRetryPolicy) ConfidenceThreshold() float64 {
	return p.threshold
}
