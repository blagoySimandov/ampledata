package orchestrator

import (
	"context"
	"sync"

	"github.com/blagoySimandov/ampledata/go/internal/feedback"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/pipeline"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/rs/zerolog/log"
)

type RetryOrchestrator struct {
	pipeline       *pipeline.Pipeline
	stateManager   *state.StateManager
	evaluator      QualityEvaluator
	feedbackBuilder FeedbackBuilder
	policy         RetryPolicy
	attemptStore   feedback.AttemptStore
	pipelineConfig *pipeline.PipelineConfig
}

type OrchestratorConfig struct {
	Pipeline        *pipeline.Pipeline
	StateManager    *state.StateManager
	Evaluator       QualityEvaluator
	FeedbackBuilder FeedbackBuilder
	Policy          RetryPolicy
	AttemptStore    feedback.AttemptStore
	PipelineConfig  *pipeline.PipelineConfig
}

func NewRetryOrchestrator(cfg OrchestratorConfig) *RetryOrchestrator {
	evaluator := cfg.Evaluator
	if evaluator == nil {
		evaluator = NewDefaultQualityEvaluator()
	}

	builder := cfg.FeedbackBuilder
	if builder == nil {
		builder = NewDefaultFeedbackBuilder()
	}

	policy := cfg.Policy
	if policy == nil {
		policy = NewDefaultRetryPolicy(DefaultRetryPolicyConfig())
	}

	attemptStore := cfg.AttemptStore
	if attemptStore == nil {
		attemptStore = feedback.NewInMemoryAttemptStore()
	}

	return &RetryOrchestrator{
		pipeline:        cfg.Pipeline,
		stateManager:    cfg.StateManager,
		evaluator:       evaluator,
		feedbackBuilder: builder,
		policy:          policy,
		attemptStore:    attemptStore,
		pipelineConfig:  cfg.PipelineConfig,
	}
}

func (o *RetryOrchestrator) Run(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error {
	if err := o.stateManager.InitializeJob(ctx, jobID, rowKeys); err != nil {
		return err
	}

	pendingRows := make(map[string]*rowContext)
	for _, key := range rowKeys {
		pendingRows[key] = &rowContext{
			key:             key,
			columnsMetadata: columnsMetadata,
			history:         o.attemptStore.GetOrCreate(key),
		}
	}

	for attempt := 1; len(pendingRows) > 0 && attempt <= o.policy.MaxAttempts(); attempt++ {
		log.Info().
			Str("jobID", jobID).
			Int("attempt", attempt).
			Int("pendingRows", len(pendingRows)).
			Msg("Starting enrichment attempt")

		keysToProcess := make([]string, 0, len(pendingRows))
		for key := range pendingRows {
			keysToProcess = append(keysToProcess, key)
		}

		results := o.runPipelineForRows(ctx, jobID, keysToProcess, columnsMetadata, pendingRows)

		completedRows := make([]string, 0)
		for key, result := range results {
			rc := pendingRows[key]

			attemptRecord := feedback.Attempt{
				Number:        attempt,
				TargetColumns: rc.focusColumns,
				PatternsUsed:  result.patternsUsed,
				URLsCrawled:   result.urlsCrawled,
				Results:       result.extractedData,
				Confidences:   toConfidenceMap(result.confidence),
			}

			assessment := o.evaluator.Evaluate(&models.EnrichmentResult{
				Key:           key,
				ExtractedData: result.extractedData,
				Confidence:    result.confidence,
				Sources:       result.urlsCrawled,
			}, o.policy.ConfidenceThreshold())

			attemptRecord.Assessment = assessment
			rc.history.Record(attemptRecord)

			if result.err != nil {
				log.Error().
					Err(result.err).
					Str("jobID", jobID).
					Str("rowKey", key).
					Msg("Row processing failed")
				completedRows = append(completedRows, key)
				// Transition to completed even on error, so it's picked up by GetResults
				// The error is already preserved in the state
				o.stateManager.Transition(ctx, jobID, key, models.StageCompleted, nil)
				continue
			}

			if assessment.Passed {
				log.Debug().
					Str("jobID", jobID).
					Str("rowKey", key).
					Int("attempt", attempt).
					Msg("Row passed quality check")
				completedRows = append(completedRows, key)
				o.stateManager.Transition(ctx, jobID, key, models.StageCompleted, nil)
				continue
			}

			if !o.policy.ShouldRetry(attempt, assessment) {
				log.Debug().
					Str("jobID", jobID).
					Str("rowKey", key).
					Int("attempt", attempt).
					Msg("Row did not pass but policy says no more retries")
				completedRows = append(completedRows, key)
				o.stateManager.Transition(ctx, jobID, key, models.StageCompleted, nil)
				continue
			}

			weakColumns := assessment.GetWeakColumnNames()
			rc.focusColumns = weakColumns
			rc.feedback = o.feedbackBuilder.Build(rc.history, assessment.WeakColumns)

			log.Debug().
				Str("jobID", jobID).
				Str("rowKey", key).
				Int("attempt", attempt).
				Strs("weakColumns", weakColumns).
				Msg("Row needs retry")
		}

		for _, key := range completedRows {
			o.attemptStore.Delete(key)
			delete(pendingRows, key)
		}
	}

	for key := range pendingRows {
		log.Warn().
			Str("jobID", jobID).
			Str("rowKey", key).
			Int("maxAttempts", o.policy.MaxAttempts()).
			Msg("Row exhausted all retry attempts")
		o.stateManager.Transition(ctx, jobID, key, models.StageCompleted, nil)
		o.attemptStore.Delete(key)
	}

	return o.stateManager.Complete(ctx, jobID)
}

type rowContext struct {
	key             string
	columnsMetadata []*models.ColumnMetadata
	focusColumns    []string
	feedback        *feedback.EnrichmentFeedback
	history         *feedback.AttemptHistory
}

type rowResult struct {
	extractedData map[string]any
	confidence    map[string]*models.FieldConfidenceInfo
	patternsUsed  []string
	urlsCrawled   []string
	err           error
}

func (o *RetryOrchestrator) runPipelineForRows(
	ctx context.Context,
	jobID string,
	rowKeys []string,
	columnsMetadata []*models.ColumnMetadata,
	rowContexts map[string]*rowContext,
) map[string]*rowResult {
	results := make(map[string]*rowResult)
	var mu sync.Mutex

	channels := make([]chan pipeline.Message, len(o.pipeline.Stages())+1)
	for i := range channels {
		channels[i] = make(chan pipeline.Message, o.pipelineConfig.ChannelBufferSize)
	}

	var stagesWg sync.WaitGroup
	for i, stage := range o.pipeline.Stages() {
		stagesWg.Add(1)
		go func(s pipeline.Stage, in, out chan pipeline.Message) {
			defer stagesWg.Done()
			s.Run(ctx, in, out)
		}(stage, channels[i], channels[i+1])
	}

	var feedWg sync.WaitGroup
	feedWg.Add(1)
	go func() {
		defer feedWg.Done()
		for _, key := range rowKeys {
			rc := rowContexts[key]

			msg := pipeline.Message{
				JobID:           jobID,
				RowKey:          key,
				ColumnsMetadata: columnsMetadata,
				Feedback:        rc.feedback,
				State: &models.RowState{
					Key:   key,
					Stage: models.StagePending,
				},
			}

			select {
			case <-ctx.Done():
				return
			case channels[0] <- msg:
			}
		}
	}()

	var collectWg sync.WaitGroup
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for msg := range channels[len(channels)-1] {
			mu.Lock()
			result := &rowResult{
				err: msg.Error,
			}
			if msg.State != nil {
				result.extractedData = msg.State.ExtractedData
				result.confidence = msg.State.Confidence
				if msg.State.Decision != nil {
					result.urlsCrawled = msg.State.Decision.URLsToCrawl
				}
				if msg.State.CrawlResults != nil {
					result.urlsCrawled = msg.State.CrawlResults.Sources
				}
			}
			result.patternsUsed = msg.QueryPatterns
			results[msg.RowKey] = result
			mu.Unlock()
		}
	}()

	feedWg.Wait()
	close(channels[0])

	stagesWg.Wait()

	collectWg.Wait()

	return results
}

func toConfidenceMap(confInfo map[string]*models.FieldConfidenceInfo) map[string]float64 {
	if confInfo == nil {
		return nil
	}
	result := make(map[string]float64)
	for k, v := range confInfo {
		if v != nil {
			result[k] = v.Score
		}
	}
	return result
}
