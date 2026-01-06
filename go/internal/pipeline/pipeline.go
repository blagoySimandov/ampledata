package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/logging"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type Pipeline struct {
	stateManager     *state.StateManager
	stages           []Stage
	config           *PipelineConfig
	patternGenerator services.QueryPatternGenerator
}

type PipelineConfig struct {
	WorkersPerStage   int
	ChannelBufferSize int
}

func NewPipeline(manager *state.StateManager, stages []Stage, config *PipelineConfig, patternGenerator services.QueryPatternGenerator) *Pipeline {
	return &Pipeline{
		stateManager:     manager,
		stages:           stages,
		config:           config,
		patternGenerator: patternGenerator,
	}
}

func (p *Pipeline) Run(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error {
	// Create a job-level WideEvent for this pipeline execution
	jobEvent := logging.NewWideEvent("job_processing")
	ctx = logging.WithContext(ctx, jobEvent)

	logging.EnrichJob(ctx, jobID, string(models.JobStatusRunning))
	logging.EnrichJobRows(ctx, len(rowKeys), 0, 0)
	logging.EnrichPipeline(ctx, "initializing", nil)

	if err := p.stateManager.InitializeJob(ctx, jobID, rowKeys); err != nil {
		logging.EnrichError(ctx, err, "initialize_job")
		logging.Emit(ctx)
		return err
	}

	patterns, err := p.patternGenerator.GeneratePatterns(ctx, columnsMetadata)
	if err != nil {
		logging.EnrichMetadata(ctx, "pattern_generation_warning", err.Error())
	}
	logging.EnrichPipeline(ctx, "pattern_generation", patterns)

	channels := make([]chan Message, len(p.stages)+1)
	for i := range channels {
		channels[i] = make(chan Message, p.config.ChannelBufferSize)
	}

	var stagesWg sync.WaitGroup
	var collectWg sync.WaitGroup

	for i, stage := range p.stages {
		stagesWg.Add(1)
		go func(s Stage, in, out chan Message) {
			defer stagesWg.Done()
			s.Run(ctx, in, out)
		}(stage, channels[i], channels[i+1])
	}

	var feedWg sync.WaitGroup
	feedWg.Add(1)
	go func() {
		defer feedWg.Done()
		p.feedInitialMessages(ctx, jobID, rowKeys, columnsMetadata, patterns, channels[0])
	}()

	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		p.collectResults(ctx, jobID, channels[len(channels)-1])
	}()

	feedWg.Wait()
	close(channels[0])

	stagesWg.Wait()

	collectWg.Wait()

	// Emit the final job event
	logging.Emit(ctx)

	return nil
}

func (p *Pipeline) feedInitialMessages(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, queryPatterns []string, outChan chan<- Message) {
	for _, key := range rowKeys {
		select {
		case <-ctx.Done():
			return
		case outChan <- Message{
			JobID:           jobID,
			RowKey:          key,
			ColumnsMetadata: columnsMetadata,
			QueryPatterns:   queryPatterns,
			State: &models.RowState{
				Key:       key,
				Stage:     models.StagePending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}:
		}
	}
}

func (p *Pipeline) collectResults(ctx context.Context, jobID string, inChan <-chan Message) {
	completedCount := 0
	failedCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-inChan:
			if !ok {
				if completedCount > 0 || failedCount > 0 {
					logging.EnrichJobRows(ctx, completedCount+failedCount, completedCount, failedCount)
					logging.EnrichJob(ctx, jobID, string(models.JobStatusCompleted))

					err := p.stateManager.Complete(ctx, jobID)
					if err != nil {
						logging.EnrichError(ctx, err, "mark_job_complete")
					}
				}
				return
			}

			if msg.Error != nil {
				failedCount++
			} else {
				completedCount++

				msg.State.Stage = models.StageCompleted
				msg.State.UpdatedAt = time.Now()
				err := p.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageCompleted, nil)
				if err != nil {
					logging.EnrichError(ctx, err, "transition_row_complete")
				}
			}
		}
	}
}
