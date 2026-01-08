package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/rs/zerolog/log"
)

// PatternStage generates query patterns based on column metadata
type PatternStage struct {
	patternGenerator services.QueryPatternGenerator
	stateManager     *state.StateManager
	numWorkers       int
}

// NewPatternStage creates a new pattern generation stage
func NewPatternStage(
	patternGenerator services.QueryPatternGenerator,
	stateManager *state.StateManager,
	numWorkers int,
) *PatternStage {
	return &PatternStage{
		patternGenerator: patternGenerator,
		stateManager:     stateManager,
		numWorkers:       numWorkers,
	}
}

// Name returns the stage name
func (s *PatternStage) Name() string {
	return "Pattern"
}

// Run executes the pattern generation stage
func (s *PatternStage) Run(ctx context.Context, inChan <-chan Message, outChan chan<- Message) {
	var wg sync.WaitGroup

	for i := 0; i < s.numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.worker(ctx, workerID, inChan, outChan)
		}(i)
	}

	wg.Wait()
	close(outChan)
}

func (s *PatternStage) worker(ctx context.Context, workerID int, inChan <-chan Message, outChan chan<- Message) {
	for msg := range inChan {
		select {
		case <-ctx.Done():
			log.Info().
				Str("stage", s.Name()).
				Int("worker", workerID).
				Msg("Worker cancelled")
			return
		default:
			s.processMessage(ctx, workerID, msg, outChan)
		}
	}
}

func (s *PatternStage) processMessage(ctx context.Context, workerID int, msg Message, outChan chan<- Message) {
	log.Debug().
		Str("stage", s.Name()).
		Int("worker", workerID).
		Str("jobID", msg.JobID).
		Str("rowKey", msg.RowKey).
		Msg("Processing message")

	isRetry := msg.Feedback != nil && msg.Feedback.IsRetry()
	if isRetry {
		log.Debug().
			Str("stage", s.Name()).
			Str("jobID", msg.JobID).
			Str("rowKey", msg.RowKey).
			Int("attemptNumber", msg.Feedback.AttemptNumber).
			Strs("focusColumns", msg.Feedback.FocusColumns).
			Msg("Generating patterns with feedback for retry")
	}

	patterns, err := s.patternGenerator.GeneratePatternsWithFeedback(ctx, msg.ColumnsMetadata, msg.Feedback)
	if err != nil {
		log.Error().
			Err(err).
			Str("stage", s.Name()).
			Str("jobID", msg.JobID).
			Str("rowKey", msg.RowKey).
			Msg("Failed to generate patterns")

		msg.Error = fmt.Errorf("pattern generation failed: %w", err)
		outChan <- msg
		return
	}

	msg.QueryPatterns = patterns

	log.Debug().
		Str("stage", s.Name()).
		Int("worker", workerID).
		Str("jobID", msg.JobID).
		Str("rowKey", msg.RowKey).
		Int("numPatterns", len(patterns)).
		Strs("patterns", patterns).
		Bool("isRetry", isRetry).
		Msg("Generated query patterns")

	outChan <- msg
}
