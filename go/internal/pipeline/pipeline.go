package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type Pipeline struct {
	stateManager    *state.StateManager
	stages          []Stage
	config          *PipelineConfig
	columnsMetadata map[string][]*models.ColumnMetadata
	mu              sync.RWMutex
}

type PipelineConfig struct {
	WorkersPerStage   int
	ChannelBufferSize int
}

func NewPipeline(manager *state.StateManager, stages []Stage, config *PipelineConfig) *Pipeline {
	return &Pipeline{
		stateManager:    manager,
		stages:          stages,
		config:          config,
		columnsMetadata: make(map[string][]*models.ColumnMetadata),
	}
}

func (p *Pipeline) Run(ctx context.Context, jobID string, rowKeys []string) error {
	return p.RunWithMetadata(ctx, jobID, rowKeys, nil)
}

func (p *Pipeline) RunWithMetadata(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error {
	if columnsMetadata != nil {
		p.mu.Lock()
		p.columnsMetadata[jobID] = columnsMetadata
		p.mu.Unlock()
		defer func() {
			p.mu.Lock()
			delete(p.columnsMetadata, jobID)
			p.mu.Unlock()
		}()
	}

	if err := p.stateManager.InitializeJob(ctx, jobID, rowKeys); err != nil {
		return err
	}

	channels := make([]chan Message, len(p.stages)+1)
	for i := range channels {
		channels[i] = make(chan Message, p.config.ChannelBufferSize)
	}

	var wg sync.WaitGroup

	for i, stage := range p.stages {
		wg.Add(1)
		go func(s Stage, in, out chan Message) {
			defer wg.Done()
			s.Run(ctx, in, out)
		}(stage, channels[i], channels[i+1])
	}

	p.mu.RLock()
	colMeta := p.columnsMetadata[jobID]
	p.mu.RUnlock()

	go p.feedInitialMessages(ctx, jobID, rowKeys, colMeta, channels[0])

	go p.collectResults(ctx, jobID, channels[len(channels)-1])

	wg.Wait()

	for _, ch := range channels {
		close(ch)
	}

	return nil
}

func (p *Pipeline) feedInitialMessages(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, outChan chan<- Message) {
	defer close(outChan)

	for _, key := range rowKeys {
		select {
		case <-ctx.Done():
			return
		case outChan <- Message{
			JobID:           jobID,
			RowKey:          key,
			ColumnsMetadata: columnsMetadata,
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
					p.stateManager.Complete(ctx, jobID)
				}
				return
			}

			if msg.Error != nil {
				failedCount++
			} else {
				completedCount++

				msg.State.Stage = models.StageCompleted
				msg.State.UpdatedAt = time.Now()
				p.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageCompleted, nil)
			}
		}
	}
}
