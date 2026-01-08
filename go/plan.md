Implementation Plan: Caching & Feedback Loop for Pipeline

Decisions Made

- Retry scope: Per-row (individual rows that don't meet threshold)
- Retry granularity: Per-column (only re-enrich columns with low confidence)
- Retry starting point: Pattern generation (generate new targeted patterns)
- Approach: Option B - Quality Gate + Retry Orchestrator

---

Part 1: Pattern Generation Caching

Design

Create a PatternCache interface that wraps pattern generation:

type PatternCache interface {
Get(ctx context.Context, key string) ([]string, bool)
Set(ctx context.Context, key string, patterns []string) error
}

type CachedPatternGenerator struct {
underlying PatternGenerator
cache PatternCache
}

Cache key: SHA256 hash of sorted column metadata (name + type + description)

Files to Create/Modify

1. internal/cache/pattern_cache.go - Interface + in-memory implementation
2. internal/services/query_pattern_generator.go - Add caching wrapper

---

Part 2: Feedback Loop Architecture

Core Concept

┌──────────────────────────────────────────────────────────────────┐
│ RetryOrchestrator │
│ │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Pipeline: Pattern → SERP → Decision → Crawl → Extract │ │
│ └─────────────────────────────────────────────────────────┘ │
│ │ │
│ QualityGate │
│ / \ │
│ [pass] [retry needed] │
│ │ │ │
│ Complete Identify weak columns │
│ │ │
│ Build targeted feedback │
│ │ │
│ Re-run with modified input │
│ (focus on weak columns only) │
└──────────────────────────────────────────────────────────────────┘

Key Interfaces

type QualityEvaluator interface {
Evaluate(result *EnrichmentResult, threshold float64)*QualityAssessment
}

type QualityAssessment struct {
Passed bool
WeakColumns []WeakColumn
Suggestions []string
}

type WeakColumn struct {
Name string
Confidence float64
Reason string
}

type FeedbackBuilder interface {
Build(history *AttemptHistory, weakColumns []WeakColumn)*EnrichmentFeedback
}

type EnrichmentFeedback struct {
FocusColumns []string
AvoidPatterns []string
AvoidURLs []string
PreviousAttempts []AttemptSummary
Hints []string
}

type RetryPolicy interface {
ShouldRetry(attemptCount int, assessment \*QualityAssessment) bool
}

type RetryPolicyConfig struct {
MaxAttempts int
ConfidenceThreshold float64
RequireImprovement bool
}

Retry Orchestrator

type RetryOrchestrator struct {
pipeline *Pipeline
evaluator QualityEvaluator
builder FeedbackBuilder
policy RetryPolicy
history*AttemptStore
}

func (o *RetryOrchestrator) ProcessRow(ctx context.Context, input*RowInput) \*EnrichmentResult {
history := o.history.GetOrCreate(input.RowKey)

     for {
         feedback := o.builder.Build(history, history.LastWeakColumns())

         result := o.pipeline.Run(ctx, input.WithFeedback(feedback))

         assessment := o.evaluator.Evaluate(result, o.policy.Threshold())
         history.Record(result, assessment)

         if assessment.Passed || !o.policy.ShouldRetry(history.Count(), assessment) {
             return result
         }
     }

}

Prompt Modification Strategy

When retrying, the PatternStage receives feedback that modifies its prompt:

Original prompt (attempt 1):
Generate search patterns for columns: [founder, employee_count, website]

Modified prompt (attempt 2, after low confidence on employee_count):
[CONTEXT: RETRY ATTEMPT 2]

Previous attempt used patterns: ["%entity founder CEO", "%entity company info"]
Result: employee_count had low confidence (0.25) - pattern did not surface employee data

[TASK]
Generate NEW patterns specifically targeting: employee_count
The entity is a company. Focus on patterns likely to find employee/headcount data.
Avoid patterns similar to previous attempts.

Data Flow for Per-Column Retry

1. Initial run: All columns processed normally
2. Quality check: Identify columns with confidence < threshold
3. Retry input: Create modified input with:

- TargetColumns: Only weak columns
- Feedback: History of previous attempts

4. Pattern generation: Generates patterns focused on weak columns
5. SERP/Decision/Crawl/Extract: Run targeting weak columns
6. Merge: Combine new results with previous good results

Attempt History

type AttemptHistory struct {
RowKey string
Attempts []Attempt
}

type Attempt struct {
Number int
TargetColumns []string
PatternsUsed []string
URLsCrawled []string
Results map[string]interface{}
Confidences map[string]float64
Assessment \*QualityAssessment
}

type AttemptStore interface {
GetOrCreate(rowKey string) *AttemptHistory
Save(history*AttemptHistory) error
}

---

File Structure

internal/
├── pipeline/
│ ├── pipeline.go # Minor: accept feedback in messages
│ ├── stage_pattern.go # Modify: use feedback in prompt
│ ├── stage_decision.go # Modify: use feedback for URL selection
│ └── message.go # Add: Feedback field to Message
├── orchestrator/
│ ├── orchestrator.go # NEW: RetryOrchestrator
│ ├── evaluator.go # NEW: QualityEvaluator impl
│ ├── builder.go # NEW: FeedbackBuilder impl
│ └── policy.go # NEW: RetryPolicy impl
├── feedback/
│ ├── types.go # NEW: Feedback types
│ └── history.go # NEW: AttemptHistory, AttemptStore
├── cache/
│ └── pattern_cache.go # NEW: PatternCache interface + impl
└── config/
└── config.go # Add retry config fields

---

Implementation Steps

Phase 1: Caching (Low risk, immediate value)

1. Create internal/cache/pattern_cache.go with interface and in-memory impl
2. Wrap GeminiPatternGenerator with caching
3. Add cache to PatternStage initialization in main.go

Phase 2: Core Feedback Types

1. Create internal/feedback/types.go with all feedback structs
2. Create internal/feedback/history.go for attempt tracking
3. Add Feedback field to pipeline Message struct

Phase 3: Quality Evaluation

1. Create internal/orchestrator/evaluator.go
2. Implement confidence threshold checking per column
3. Generate weak column list and suggestions

Phase 4: Feedback Builder

1. Create internal/orchestrator/builder.go
2. Implement logic to create targeted feedback from history
3. Format previous attempts for prompt injection

Phase 5: Stage Modifications

1. Modify stage_pattern.go to read feedback and adjust prompt
2. Modify stage_decision.go to avoid previously crawled URLs
3. Ensure stages handle partial column targeting

Phase 6: Retry Orchestrator

1. Create internal/orchestrator/orchestrator.go
2. Implement retry loop with policy
3. Handle result merging across attempts

Phase 7: Integration

1. Add retry config to Config struct
2. Wire orchestrator into enricher.go
3. Add attempt history persistence (optional)

---

Configuration

type RetryConfig struct {
Enabled bool
MaxAttempts int
ConfidenceThreshold float64
RequireImprovement bool
}

Defaults:

- MaxAttempts: 3
- ConfidenceThreshold: 0.6
- RequireImprovement: true (stop if no improvement between attempts)

---

Verification Plan

1. Unit tests: Test evaluator, builder, policy independently
2. Integration test: Mock pipeline, verify retry loop behavior
3. Manual test: Run with real data, verify:

- Cache hits reduce API calls
- Low-confidence columns trigger retry
- Feedback appears in subsequent prompts
- Results improve after retry

---

Critical Files to Modify

- internal/pipeline/stage.go - Add Feedback field to Message
- internal/pipeline/stage_pattern.go - Use feedback in prompt building
- internal/pipeline/stage_decision.go - Avoid previously tried URLs
- internal/services/query_pattern_generator.go - Accept feedback parameter
- internal/enricher/enricher.go - Use orchestrator instead of raw pipeline
- cmd/server/main.go - Wire up cache and orchestrator
