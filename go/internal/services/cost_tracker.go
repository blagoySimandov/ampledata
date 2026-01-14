package services

import (
	"context"
	"sync"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
)

type contextKey string

const jobIDContextKey contextKey = "jobID"

func ContextWithJobID(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, jobIDContextKey, jobID)
}

func JobIDFromContext(ctx context.Context) string {
	if v := ctx.Value(jobIDContextKey); v != nil {
		if jobID, ok := v.(string); ok {
			return jobID
		}
	}
	return ""
}

type CostStore interface {
	IncrementJobCost(ctx context.Context, jobID string, costDollars, costCredits int) error
}

type ICostTracker interface {
	AddTokenCost(ctx context.Context, tknIn, tknOut int)
	AddSearchQueryCost(ctx context.Context, count int)
	CostDollars() int
	CostCredits() int
}

type CostTracker struct {
	mu              sync.RWMutex
	cost            int
	tknInCost       int
	tknOutCost      int
	searchQueryCost int
	creditExchange  int
	store           CostStore
}

type CostTrackerOption func(*CostTracker) error

func NewCostTracker(tknInCost, tknOutCost int, searchQueryCost int, creditExchange int, opts ...CostTrackerOption) (*CostTracker, error) {
	ct := &CostTracker{
		tknInCost:       tknInCost,
		tknOutCost:      tknOutCost,
		searchQueryCost: searchQueryCost,
		creditExchange:  creditExchange,
	}

	for _, opt := range opts {
		if err := opt(ct); err != nil {
			return nil, err
		}
	}

	return ct, nil
}

func WithStore(store CostStore) CostTrackerOption {
	return func(ct *CostTracker) error {
		ct.store = store
		return nil
	}
}

func (c *CostTracker) AddTokenCost(ctx context.Context, tknIn, tknOut int) {
	c.mu.Lock()
	totalCost := tknIn*c.tknInCost + tknOut*c.tknOutCost
	credits := totalCost * c.creditExchange
	c.cost += totalCost
	c.mu.Unlock()

	jobID := JobIDFromContext(ctx)
	if jobID != "" && c.store != nil {
		go func() {
			if err := c.store.IncrementJobCost(context.Background(), jobID, totalCost, credits); err != nil {
				logger.Log.Error("failed to increment job cost", "error", err, "job_id", jobID)
			}
		}()
	}
}

func (c *CostTracker) AddSearchQueryCost(ctx context.Context, count int) {
	c.mu.Lock()
	cost := count * c.searchQueryCost
	credits := cost * c.creditExchange
	c.cost += cost
	c.mu.Unlock()

	jobID := JobIDFromContext(ctx)
	if jobID != "" && c.store != nil {
		go func() {
			if err := c.store.IncrementJobCost(context.Background(), jobID, cost, credits); err != nil {
				logger.Log.Error("failed to increment job cost", "error", err, "job_id", jobID)
			}
		}()
	}
}

func (c *CostTracker) CostDollars() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cost
}

func (c *CostTracker) CostCredits() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cost * c.creditExchange
}
