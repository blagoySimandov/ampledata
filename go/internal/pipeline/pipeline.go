package pipeline

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type Pipeline struct {
	stateManager *state.StateManager
	stages       []Stage
	config       *PipelineConfig
}

type PipelineConfig struct {
	WorkersPerStage   int
	ChannelBufferSize int
}

func NewPipeline(manager *state.StateManager, stages []Stage, config *PipelineConfig) *Pipeline {
	return &Pipeline{
		stateManager: manager,
		stages:       stages,
		config:       config,
	}
}

func (p *Pipeline) Stages() []Stage {
	return p.stages
}

func (p *Pipeline) Run(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error {
	if err := p.stateManager.InitializeJob(ctx, jobID, rowKeys); err != nil {
		return err
	}

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
		p.feedInitialMessages(ctx, jobID, rowKeys, columnsMetadata, channels[0])
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

	return nil
}

func (p *Pipeline) feedInitialMessages(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, outChan chan<- Message) {
	for _, key := range rowKeys {
		select {
		case <-ctx.Done():
			return
		case outChan <- Message{
			JobID:           jobID,
			RowKey:          key,
			ColumnsMetadata: columnsMetadata,
			QueryPatterns:   nil, // Will be populated by PatternStage
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
					err := p.stateManager.Complete(ctx, jobID)
					// TODO: proper error state handling. should be persisted in the db... maybe return ?
					if err != nil {
						log.Printf("failed to mark job as complete: %s", err)
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
				// TODO: proper error state handling. should persist in the db...
				err := p.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageCompleted, nil)
				if err != nil {
					log.Printf("failed to transition row state: %s", err)
				}
			}
		}
	}
}
