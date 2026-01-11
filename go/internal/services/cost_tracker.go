package services

// All prices are in 1/10 cents
type CostTracker struct {
	cost            int
	tknInCost       int
	tknOutCost      int
	searchQueryCost int
	creditExchange  int
}

func NewCostTracker(tknInCost, tknOutCost int) *CostTracker {
	return &CostTracker{
		tknInCost:  tknInCost,
		tknOutCost: tknOutCost,
	}
}

func (c *CostTracker) AddTokenCost(tknIn, tknOut int) {
	totalCost := tknIn*c.tknInCost + tknOut*c.tknOutCost
	c.cost += totalCost
}

func (c *CostTracker) AddSearchQueryCost(searchQueryCost int) {
	c.searchQueryCost += searchQueryCost
}

func (c *CostTracker) CostDollars() int {
	return c.cost
}

func (c *CostTracker) CostCredits() int {
	return c.cost * c.creditExchange
}
