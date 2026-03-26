package workflows

import "go.temporal.io/sdk/workflow"

type workflowSemaphore struct {
	ctx      workflow.Context
	limit    int
	inFlight int
	futures  []workflow.Future
}

func (s *workflowSemaphore) Acquire() {
	if s.inFlight < s.limit {
		return
	}
	sel := workflow.NewSelector(s.ctx)
	for _, f := range s.futures[len(s.futures)-s.inFlight:] {
		f := f
		sel.AddFuture(f, func(workflow.Future) { s.inFlight-- })
	}
	sel.Select(s.ctx)
}

func (s *workflowSemaphore) Add(f workflow.Future) {
	s.futures = append(s.futures, f)
	s.inFlight++
}

func (s *workflowSemaphore) Futures() []workflow.Future {
	return s.futures
}
