package wideevent

import (
	"context"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"go.temporal.io/sdk/workflow"
)

type EventType string

const (
	EventJobStarted       EventType = "job.started"
	EventJobCompleted     EventType = "job.completed"
	EventJobFailed        EventType = "job.failed"
	EventEnrichmentStart  EventType = "enrichment.started"
	EventEnrichmentDone   EventType = "enrichment.completed"
	EventEnrichmentError  EventType = "enrichment.failed"
	EventActivityStart    EventType = "activity.started"
	EventActivityComplete EventType = "activity.completed"
	EventActivityError    EventType = "activity.failed"
)

type WorkflowEvent struct {
	EventType     EventType              `json:"event_type"`
	Timestamp     time.Time              `json:"timestamp"`
	WorkflowID    string                 `json:"workflow_id"`
	RunID         string                 `json:"run_id"`
	JobID         string                 `json:"job_id"`
	RowKey        *string                `json:"row_key,omitempty"`
	Stage         models.RowStage        `json:"stage,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	DurationMs    int64                  `json:"duration_ms,omitempty"`
	Status        string                 `json:"status"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	ErrorType     *string                `json:"error_type,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type JobEventContext struct {
	StartTime  time.Time
	JobID      string
	UserID     string
	WorkflowID string
	RunID      string
	TotalRows  int
	Completed  int
	Failed     int
	Metadata   map[string]interface{}
}

type EnrichmentEventContext struct {
	StartTime   time.Time
	JobID       string
	RowKey      string
	WorkflowID  string
	RunID       string
	UserID      string
	Stages      map[models.RowStage]StageMetrics
	Metadata    map[string]interface{}
}

type StageMetrics struct {
	StartTime  time.Time              `json:"start_time"`
	DurationMs int64                  `json:"duration_ms"`
	Status     string                 `json:"status"`
	Error      *string                `json:"error,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func NewJobEventContext(jobID, userID string) *JobEventContext {
	return &JobEventContext{
		StartTime: time.Now(),
		JobID:     jobID,
		UserID:    userID,
		Metadata:  make(map[string]interface{}),
	}
}

func NewEnrichmentEventContext(jobID, rowKey, userID string) *EnrichmentEventContext {
	return &EnrichmentEventContext{
		StartTime: time.Now(),
		JobID:     jobID,
		RowKey:    rowKey,
		UserID:    userID,
		Stages:    make(map[models.RowStage]StageMetrics),
		Metadata:  make(map[string]interface{}),
	}
}

func (ctx *JobEventContext) SetWorkflowInfo(workflowID, runID string) {
	ctx.WorkflowID = workflowID
	ctx.RunID = runID
}

func (ctx *EnrichmentEventContext) SetWorkflowInfo(workflowID, runID string) {
	ctx.WorkflowID = workflowID
	ctx.RunID = runID
}

func (ctx *JobEventContext) EmitSuccess(wfCtx workflow.Context) {
	logger := workflow.GetLogger(wfCtx)
	duration := time.Since(ctx.StartTime).Milliseconds()

	logger.Info("Job workflow completed",
		"event_type", string(EventJobCompleted),
		"job_id", ctx.JobID,
		"user_id", ctx.UserID,
		"workflow_id", ctx.WorkflowID,
		"run_id", ctx.RunID,
		"duration_ms", duration,
		"total_rows", ctx.TotalRows,
		"completed", ctx.Completed,
		"failed", ctx.Failed,
		"status", "success",
	)
}

func (ctx *JobEventContext) EmitError(wfCtx workflow.Context, err error) {
	logger := workflow.GetLogger(wfCtx)
	duration := time.Since(ctx.StartTime).Milliseconds()
	errMsg := err.Error()

	logger.Error("Job workflow failed",
		"event_type", string(EventJobFailed),
		"job_id", ctx.JobID,
		"user_id", ctx.UserID,
		"workflow_id", ctx.WorkflowID,
		"run_id", ctx.RunID,
		"duration_ms", duration,
		"total_rows", ctx.TotalRows,
		"completed", ctx.Completed,
		"failed", ctx.Failed,
		"status", "error",
		"error_message", errMsg,
	)
}

func (ctx *EnrichmentEventContext) StartStage(stage models.RowStage) {
	ctx.Stages[stage] = StageMetrics{
		StartTime: time.Now(),
		Status:    "in_progress",
	}
}

func (ctx *EnrichmentEventContext) CompleteStage(stage models.RowStage, data map[string]interface{}) {
	metrics := ctx.Stages[stage]
	metrics.DurationMs = time.Since(metrics.StartTime).Milliseconds()
	metrics.Status = "success"
	metrics.Data = data
	ctx.Stages[stage] = metrics
}

func (ctx *EnrichmentEventContext) FailStage(stage models.RowStage, err error) {
	metrics := ctx.Stages[stage]
	metrics.DurationMs = time.Since(metrics.StartTime).Milliseconds()
	metrics.Status = "error"
	errMsg := err.Error()
	metrics.Error = &errMsg
	ctx.Stages[stage] = metrics
}

func (ctx *EnrichmentEventContext) EmitSuccess(wfCtx workflow.Context, finalStage models.RowStage) {
	logger := workflow.GetLogger(wfCtx)
	duration := time.Since(ctx.StartTime).Milliseconds()

	logger.Info("Enrichment workflow completed",
		"event_type", string(EventEnrichmentDone),
		"job_id", ctx.JobID,
		"row_key", ctx.RowKey,
		"user_id", ctx.UserID,
		"workflow_id", ctx.WorkflowID,
		"run_id", ctx.RunID,
		"duration_ms", duration,
		"final_stage", string(finalStage),
		"status", "success",
		"stages", ctx.Stages,
	)
}

func (ctx *EnrichmentEventContext) EmitError(wfCtx workflow.Context, currentStage models.RowStage, err error) {
	logger := workflow.GetLogger(wfCtx)
	duration := time.Since(ctx.StartTime).Milliseconds()
	errMsg := err.Error()

	logger.Error("Enrichment workflow failed",
		"event_type", string(EventEnrichmentError),
		"job_id", ctx.JobID,
		"row_key", ctx.RowKey,
		"user_id", ctx.UserID,
		"workflow_id", ctx.WorkflowID,
		"run_id", ctx.RunID,
		"duration_ms", duration,
		"current_stage", string(currentStage),
		"status", "error",
		"error_message", errMsg,
		"stages", ctx.Stages,
	)
}

type contextKey string

const eventContextKey contextKey = "wide_event_context"

func WithContext(ctx context.Context, event interface{}) context.Context {
	return context.WithValue(ctx, eventContextKey, event)
}

func FromContext(ctx context.Context) interface{} {
	return ctx.Value(eventContextKey)
}
