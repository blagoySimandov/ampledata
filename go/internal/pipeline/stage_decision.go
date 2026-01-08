package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/rs/zerolog/log"
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

			cancelled, _ := s.stateManager.CheckCancelled(ctx, msg.JobID)
			if cancelled {
				return
			}

			if msg.Error != nil {
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
				continue
			}

			if msg.State.SerpData == nil {
				errStr := "SERP data is missing"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
			} else if msg.State.SerpData.Results == nil || len(msg.State.SerpData.Results) == 0 {
				errStr := "SERP results not found in data"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
			} else {
				mergedResults := s.mergeSerpResults(msg.State.SerpData.Results)
				decision, err := s.decisionMaker.MakeDecision(ctx, mergedResults, msg.RowKey, 3, msg.ColumnsMetadata)
				if err != nil {
					errStr := err.Error()
					msg.Error = err
					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]any{
						"error": errStr,
					})
				} else {
					filteredURLs := s.filterAvoidedURLs(decision.URLsToCrawl, msg)

					msg.State.Decision = &models.Decision{
						URLsToCrawl:    filteredURLs,
						ExtractedData:  decision.ExtractedData,
						Reasoning:      decision.Reasoning,
						MissingColumns: decision.MissingColumns,
					}
					msg.State.Stage = models.StageDecisionMade
					msg.State.UpdatedAt = time.Now()

					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageDecisionMade, map[string]any{
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

func (s *DecisionStage) filterAvoidedURLs(urls []string, msg Message) []string {
	if msg.Feedback == nil || len(msg.Feedback.AvoidURLs) == 0 {
		return urls
	}

	avoidSet := make(map[string]struct{})
	for _, u := range msg.Feedback.AvoidURLs {
		avoidSet[u] = struct{}{}
	}

	filtered := make([]string, 0, len(urls))
	for _, u := range urls {
		if _, avoid := avoidSet[u]; !avoid {
			filtered = append(filtered, u)
		} else {
			log.Debug().
				Str("stage", s.Name()).
				Str("jobID", msg.JobID).
				Str("rowKey", msg.RowKey).
				Str("url", u).
				Msg("Filtering out previously crawled URL")
		}
	}

	return filtered
}
