package services

import "sync"

type ICostTracker interface {
	AddTokenCost(tknIn, tknOut int)
	AddSearchQueryCost(searchQueryCost int)
	CostDollars() int
	CostCredits() int
}

// All prices are 1/10 of a cent
type CostTracker struct {
	mu              sync.RWMutex
	cost            int
	tknInCost       int
	tknOutCost      int
	searchQueryCost int
	creditExchange  int
}

func NewCostTracker(tknInCost, tknOutCost int, searchQueryCost int, creditExchange int) *CostTracker {
	return &CostTracker{
		tknInCost:       tknInCost,
		tknOutCost:      tknOutCost,
		searchQueryCost: searchQueryCost,
		creditExchange:  creditExchange,
	}
}

func (c *CostTracker) AddTokenCost(tknIn, tknOut int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	totalCost := tknIn*c.tknInCost + tknOut*c.tknOutCost
	c.cost += totalCost
}

func (c *CostTracker) AddSearchQueryCost(searchQueryCost int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.searchQueryCost += searchQueryCost
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
