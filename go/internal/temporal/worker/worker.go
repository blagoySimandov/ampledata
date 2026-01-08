package worker

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/workflows"
)

// Worker wraps a Temporal worker
type Worker struct {
	temporalWorker worker.Worker
}

// NewWorker creates and configures a new Temporal worker
func NewWorker(temporalClient client.Client, taskQueue string, activities *activities.Activities) *Worker {
	// Create worker
	w := worker.New(temporalClient, taskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.JobWorkflow)
	w.RegisterWorkflow(workflows.EnrichmentWorkflow)

	// Register activities
	// We use string names for activities to make them more discoverable and maintainable
	w.RegisterActivityWithOptions(activities.GeneratePatterns, worker.ActivityRegistrationOptions{
		Name: "GeneratePatterns",
	})
	w.RegisterActivityWithOptions(activities.SerpFetch, worker.ActivityRegistrationOptions{
		Name: "SerpFetch",
	})
	w.RegisterActivityWithOptions(activities.MakeDecision, worker.ActivityRegistrationOptions{
		Name: "MakeDecision",
	})
	w.RegisterActivityWithOptions(activities.Crawl, worker.ActivityRegistrationOptions{
		Name: "Crawl",
	})
	w.RegisterActivityWithOptions(activities.Extract, worker.ActivityRegistrationOptions{
		Name: "Extract",
	})
	w.RegisterActivityWithOptions(activities.UpdateState, worker.ActivityRegistrationOptions{
		Name: "UpdateState",
	})
	w.RegisterActivityWithOptions(activities.AnalyzeFeedback, worker.ActivityRegistrationOptions{
		Name: "AnalyzeFeedback",
	})
	w.RegisterActivityWithOptions(activities.InitializeJob, worker.ActivityRegistrationOptions{
		Name: "InitializeJob",
	})
	w.RegisterActivityWithOptions(activities.CompleteJob, worker.ActivityRegistrationOptions{
		Name: "CompleteJob",
	})

	log.Printf("Temporal worker created with task queue: %s", taskQueue)

	return &Worker{
		temporalWorker: w,
	}
}

// Start starts the worker
func (w *Worker) Start() error {
	log.Println("Starting Temporal worker...")
	return w.temporalWorker.Start()
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	log.Println("Stopping Temporal worker...")
	w.temporalWorker.Stop()
}
