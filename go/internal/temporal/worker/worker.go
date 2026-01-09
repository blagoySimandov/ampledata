package worker

import (
	activity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/workflows"
)

type Worker struct {
	temporalWorker worker.Worker
}

func NewWorker(temporalClient client.Client, taskQueue string, activities *activities.Activities) *Worker {
	w := worker.New(temporalClient, taskQueue, worker.Options{
		Logger: logger.NewTemporalLogger(),
	})

	// Register workflows
	w.RegisterWorkflow(workflows.JobWorkflow)
	w.RegisterWorkflow(workflows.EnrichmentWorkflow)

	// Register activities
	// We use string names for activities to make them more discoverable and maintainable
	w.RegisterActivityWithOptions(activities.GeneratePatterns, activity.RegisterOptions{
		Name: "GeneratePatterns",
	})
	w.RegisterActivityWithOptions(activities.SerpFetch, activity.RegisterOptions{
		Name: "SerpFetch",
	})
	w.RegisterActivityWithOptions(activities.MakeDecision, activity.RegisterOptions{
		Name: "MakeDecision",
	})
	w.RegisterActivityWithOptions(activities.Crawl, activity.RegisterOptions{
		Name: "Crawl",
	})
	w.RegisterActivityWithOptions(activities.Extract, activity.RegisterOptions{
		Name: "Extract",
	})
	w.RegisterActivityWithOptions(activities.UpdateState, activity.RegisterOptions{
		Name: "UpdateState",
	})
	w.RegisterActivityWithOptions(activities.AnalyzeFeedback, activity.RegisterOptions{
		Name: "AnalyzeFeedback",
	})
	w.RegisterActivityWithOptions(activities.InitializeJob, activity.RegisterOptions{
		Name: "InitializeJob",
	})
	w.RegisterActivityWithOptions(activities.CompleteJob, activity.RegisterOptions{
		Name: "CompleteJob",
	})

	return &Worker{
		temporalWorker: w,
	}
}

func (w *Worker) Start() error {
	return w.temporalWorker.Start()
}

func (w *Worker) Stop() {
	w.temporalWorker.Stop()
}
