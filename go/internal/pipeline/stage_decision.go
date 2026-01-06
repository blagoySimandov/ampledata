package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/logging"
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

func (s *DecisionStage) mergeSerpResults(results []*models.GoogleSearchResults) *models.GoogleSearchResults {
	if len(results) == 0 {
		return nil
	}
	if len(results) == 1 {
		return results[0]
	}

	merged := &models.GoogleSearchResults{
		SearchParameters: results[0].SearchParameters,
		Organic:          []models.OrganicResult{},
		PeopleAlsoAsk:    []models.PeopleAlsoAskItem{},
		RelatedSearches:  []models.RelatedSearch{},
	}

	seenURLs := make(map[string]bool)
	for _, result := range results {
		if result.KnowledgeGraph != nil && merged.KnowledgeGraph == nil {
			merged.KnowledgeGraph = result.KnowledgeGraph
		}

		for _, organic := range result.Organic {
			url := ""
			if organic.Link != nil {
				url = *organic.Link
			}
			if url != "" && !seenURLs[url] {
				merged.Organic = append(merged.Organic, organic)
				seenURLs[url] = true
			}
		}

		merged.PeopleAlsoAsk = append(merged.PeopleAlsoAsk, result.PeopleAlsoAsk...)
		merged.RelatedSearches = append(merged.RelatedSearches, result.RelatedSearches...)
	}

	return merged
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

			start := time.Now()

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
				logging.EmitRowEvent(ctx, "row_decision_failed", msg.JobID, msg.RowKey, string(models.StageDecisionMade), time.Since(start), msg.Error)
			} else if msg.State.SerpData.Results == nil || len(msg.State.SerpData.Results) == 0 {
				errStr := "SERP results not found in data"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
				logging.EmitRowEvent(ctx, "row_decision_failed", msg.JobID, msg.RowKey, string(models.StageDecisionMade), time.Since(start), msg.Error)
			} else {
				mergedResults := s.mergeSerpResults(msg.State.SerpData.Results)
				decision, err := s.decisionMaker.MakeDecision(ctx, mergedResults, msg.RowKey, 3, msg.ColumnsMetadata)
				if err != nil {
					errStr := err.Error()
					msg.Error = err
					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
						"error": errStr,
					})
					logging.EmitRowEvent(ctx, "row_decision_failed", msg.JobID, msg.RowKey, string(models.StageDecisionMade), time.Since(start), err)
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
					logging.EmitRowEvent(ctx, "row_decision_completed", msg.JobID, msg.RowKey, string(models.StageDecisionMade), time.Since(start), nil)
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
