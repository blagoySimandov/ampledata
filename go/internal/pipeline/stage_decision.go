package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type DecisionStage struct {
	decisionMaker services.DecisionMaker
	stateManager  *state.StateManager
	workerCount   int
}

func NewDecisionStage(
	decisionMaker services.DecisionMaker,
	stateManager *state.StateManager,
	workerCount int,
) *DecisionStage {
	return &DecisionStage{
		decisionMaker: decisionMaker,
		stateManager:  stateManager,
		workerCount:   workerCount,
	}
}

func (s *DecisionStage) Name() string {
	return "Decision"
}

func (s *DecisionStage) Run(ctx context.Context, inChan <-chan Message, outChan chan<- Message) {
	var wg sync.WaitGroup

	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, inChan, outChan)
	}

	wg.Wait()
	close(outChan)
}

func (s *DecisionStage) worker(ctx context.Context, wg *sync.WaitGroup, in <-chan Message, out chan<- Message) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			cancelled, _ := s.stateManager.CheckCancelled(ctx, msg.JobID)
			if cancelled {
				return
			}

			if msg.State.SerpData == nil {
				errStr := "SERP data is missing"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
			} else if msg.State.SerpData.Results == nil {
				errStr := "SERP results not found in data"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
			} else {
				decision, err := s.decisionMaker.MakeDecision(ctx, msg.State.SerpData.Results, msg.RowKey, 3, msg.ColumnsMetadata)
				if err != nil {
					errStr := err.Error()
					msg.Error = err
					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
						"error": errStr,
					})
				} else {
					msg.State.Decision = &models.Decision{
						URLsToCrawl:    decision.URLsToCrawl,
						ExtractedData:  decision.ExtractedData,
						Reasoning:      decision.Reasoning,
						MissingColumns: decision.MissingColumns,
					}
					msg.State.Stage = models.StageDecisionMade
					msg.State.UpdatedAt = time.Now()

					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageDecisionMade, map[string]interface{}{
						"decision": msg.State.Decision,
					})
				}
			}

			select {
			case out <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}
