package pipeline

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/logging"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type CrawlStage struct {
	crawler      services.WebCrawler
	stateManager *state.StateManager
	workerCount  int
}

func NewCrawlStage(
	crawler services.WebCrawler,
	stateManager *state.StateManager,
	workerCount int,
) *CrawlStage {
	return &CrawlStage{
		crawler:      crawler,
		stateManager: stateManager,
		workerCount:  workerCount,
	}
}

func (s *CrawlStage) Name() string {
	return "Crawl"
}

func (s *CrawlStage) Run(ctx context.Context, inChan <-chan Message, outChan chan<- Message) {
	var wg sync.WaitGroup

	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, inChan, outChan)
	}

	wg.Wait()
	close(outChan)
}

func (s *CrawlStage) worker(ctx context.Context, wg *sync.WaitGroup, in <-chan Message, out chan<- Message) {
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

			if msg.State.Decision == nil || msg.State.SerpData == nil {
				errStr := "Missing decision or SERP data"
				msg.Error = fmt.Errorf(errStr)
				s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
					"error": errStr,
				})
				logging.EmitRowEvent(ctx, "row_crawl_failed", msg.JobID, msg.RowKey, string(models.StageCrawled), time.Since(start), msg.Error)
			} else {
				urlsToCrawl := msg.State.Decision.URLsToCrawl
				if len(urlsToCrawl) == 0 {
					msg.State.CrawlResults = &models.CrawlResults{
						Content: nil,
						Sources: []string{},
					}
					msg.State.Stage = models.StageCrawled
					msg.State.UpdatedAt = time.Now()

					s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageCrawled, map[string]interface{}{
						"crawl_results": msg.State.CrawlResults,
					})
					logging.EmitRowEvent(ctx, "row_crawl_skipped", msg.JobID, msg.RowKey, string(models.StageCrawled), time.Since(start), nil)
				} else {
					query := strings.Join(msg.State.SerpData.Queries, " ")

					content, err := s.crawler.Crawl(ctx, urlsToCrawl, query)
					if err != nil {
						errStr := err.Error()
						msg.Error = err
						s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageFailed, map[string]interface{}{
							"error": errStr,
						})
						logging.EmitRowEvent(ctx, "row_crawl_failed", msg.JobID, msg.RowKey, string(models.StageCrawled), time.Since(start), err)
					} else {
						msg.State.CrawlResults = &models.CrawlResults{
							Content: &content,
							Sources: urlsToCrawl,
						}
						msg.State.Stage = models.StageCrawled
						msg.State.UpdatedAt = time.Now()

						s.stateManager.Transition(ctx, msg.JobID, msg.RowKey, models.StageCrawled, map[string]interface{}{
							"crawl_results": msg.State.CrawlResults,
						})
						logging.EmitRowEvent(ctx, "row_crawl_completed", msg.JobID, msg.RowKey, string(models.StageCrawled), time.Since(start), nil)
					}
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
