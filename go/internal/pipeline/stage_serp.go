package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type SerpStage struct {
	webSearcher  services.WebSearcher
	stateManager *state.StateManager
	workerCount  int
}

func NewSerpStage(
	webSearcher services.WebSearcher,
	stateManager *state.StateManager,
	workerCount int,
) *SerpStage {
	return &SerpStage{
		webSearcher:  webSearcher,
		stateManager: stateManager,
		workerCount:  workerCount,
	}
}

func (s *SerpStage) Name() string {
	return "SERP"
}

func (s *SerpStage) Run(ctx context.Context, inChan <-chan Message, outChan chan<- Message) {
	var wg sync.WaitGroup

	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, inChan, outChan)
	}

	wg.Wait()
	close(outChan)
}

func (s *SerpStage) worker(ctx context.Context, wg *sync.WaitGroup, in <-chan Message, out chan<- Message) {
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

			queryBuilder := services.NewPatternQueryBuilder(msg.QueryPatterns, msg.ColumnsMetadata)
			queries := queryBuilder.Build(msg.RowKey)

			allResults := []*models.GoogleSearchResults{}
			var lastErr error

			for _, query := range queries {
				serp, err := s.webSearcher.Search(ctx, query)
				if err != nil {
					lastErr = err
					continue
				}
				allResults = append(allResults, serp)
			}

			if len(allResults) == 0 {
				msg.Error = lastErr
				errStr := lastErr.Error()
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
			} else {
				msg.State.SerpData = &models.SerpData{
					Queries: queries,
					Results: allResults,
				}
				msg.State.Stage = models.StageSerpFetched
				msg.State.UpdatedAt = time.Now()

				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageSerpFetched, map[string]interface{}{
					"serp_data": msg.State.SerpData,
				})
			}

			select {
			case out <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}
