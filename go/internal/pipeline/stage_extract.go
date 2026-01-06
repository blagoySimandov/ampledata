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

type ExtractStage struct {
	extractor    services.ContentExtractor
	stateManager *state.StateManager
	workerCount  int
}

func NewExtractStage(
	extractor services.ContentExtractor,
	stateManager *state.StateManager,
	workerCount int,
) *ExtractStage {
	return &ExtractStage{
		extractor:    extractor,
		stateManager: stateManager,
		workerCount:  workerCount,
	}
}

func (s *ExtractStage) Name() string {
	return "Extract"
}

func (s *ExtractStage) Run(ctx context.Context, inChan <-chan Message, outChan chan<- Message) {
	var wg sync.WaitGroup

	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, inChan, outChan)
	}

	wg.Wait()
	close(outChan)
}

func (s *ExtractStage) worker(ctx context.Context, wg *sync.WaitGroup, in <-chan Message, out chan<- Message) {
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

			if msg.State.Decision == nil || msg.State.CrawlResults == nil {
				errStr := "Missing decision or crawl results"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
				logging.EmitRowEvent(ctx, "row_extract_failed", msg.JobID, msg.RowKey, string(models.StageEnriched), time.Since(start), msg.Error)
			} else {
				hasContent := msg.State.CrawlResults.Content != nil && *msg.State.CrawlResults.Content != ""
				var content string
				if hasContent {
					content = *msg.State.CrawlResults.Content
				}
				if !hasContent || content == "" {
					msg.State.ExtractedData = msg.State.Decision.ExtractedData
				} else {
					missingColumns := msg.State.Decision.MissingColumns
					missingColsMetadata := []*models.ColumnMetadata{}

					for _, colName := range missingColumns {
						for _, col := range msg.ColumnsMetadata {
							if col.Name == colName {
								missingColsMetadata = append(missingColsMetadata, col)
								break
							}
						}
					}

					if len(missingColsMetadata) > 0 {
						result, err := s.extractor.Extract(ctx, content, msg.RowKey, missingColsMetadata)
						if err != nil {
							errStr := err.Error()
							msg.Error = err
							s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
								"error": errStr,
							})
							logging.EmitRowEvent(ctx, "row_extract_failed", msg.JobID, msg.RowKey, string(models.StageEnriched), time.Since(start), err)
						} else {
							extractedFromDecision := msg.State.Decision.ExtractedData
							if extractedFromDecision == nil {
								extractedFromDecision = make(map[string]interface{})
							}

							merged := make(map[string]interface{})
							for k, v := range extractedFromDecision {
								merged[k] = v
							}
							for k, v := range result.ExtractedData {
								merged[k] = v
							}

							msg.State.ExtractedData = merged
						}
					} else {
						extractedFromDecision := msg.State.Decision.ExtractedData
						msg.State.ExtractedData = extractedFromDecision
					}
				}

				if msg.Error == nil {
					msg.State.Stage = models.StageEnriched
					msg.State.UpdatedAt = time.Now()

					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageEnriched, map[string]interface{}{
						"extracted_data": msg.State.ExtractedData,
					})

					fields := []string{}
					if msg.State.ExtractedData != nil {
						for k := range msg.State.ExtractedData {
							fields = append(fields, k)
						}
					}
					logging.EmitRowEvent(ctx, "row_extract_completed", msg.JobID, msg.RowKey, string(models.StageEnriched), time.Since(start), nil)
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
