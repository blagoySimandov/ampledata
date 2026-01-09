package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"go.temporal.io/sdk/workflow"
)

var (
	Log     *slog.Logger
	sampler Sampler
)

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
	sampler = NewTailSampler(0.05) // 5% sampling for successful requests
}

// WideEvent represents a comprehensive log event with all context
// Following the wide event pattern: one event per request/workflow with high cardinality and dimensionality
type WideEvent struct {
	// Timestamp
	Timestamp time.Time `json:"timestamp"`

	// Service identification
	ServiceName    string `json:"service_name,omitempty"`
	ServiceVersion string `json:"service_version,omitempty"`
	Region         string `json:"region,omitempty"`

	// Request/Workflow identification (high cardinality)
	WorkflowID string `json:"workflow_id,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	TraceID    string `json:"trace_id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`

	// Business context (high cardinality)
	JobID  string `json:"job_id,omitempty"`
	RowKey string `json:"row_key,omitempty"`
	UserID string `json:"user_id,omitempty"`

	// Operation details
	EventType  string `json:"event_type"`
	Stage      string `json:"stage,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Status     string `json:"status"`

	// Error context
	Error *ErrorInfo `json:"error,omitempty"`

	// High-dimensionality metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Stage tracking for multi-stage operations
	Stages map[string]*StageInfo `json:"stages,omitempty"`

	// Job-specific fields
	TotalRows int `json:"total_rows,omitempty"`
	Completed int `json:"completed,omitempty"`
	Failed    int `json:"failed,omitempty"`

	// Internal tracking
	startTime time.Time
}

// StageInfo tracks metrics for individual stages
type StageInfo struct {
	StartTime  time.Time              `json:"start_time"`
	DurationMs int64                  `json:"duration_ms,omitempty"`
	Status     string                 `json:"status"`
	Error      *string                `json:"error,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// ErrorInfo provides structured error information
type ErrorInfo struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
	Retriable bool   `json:"retriable"`
}

// Sampler decides whether to keep an event based on tail-based sampling
type Sampler interface {
	ShouldSample(event *WideEvent) bool
}

// TailSampler implements tail-based sampling:
// - Always keep errors
// - Always keep slow requests
// - Sample successful fast requests at a low rate
type TailSampler struct {
	successRate      float64
	slowThresholdMs  int64
	errorSampleRate  float64
}

// NewTailSampler creates a tail sampler with the given success rate
func NewTailSampler(successRate float64) *TailSampler {
	return &TailSampler{
		successRate:     successRate,
		slowThresholdMs: 2000, // 2 seconds
		errorSampleRate: 1.0,  // Always keep errors
	}
}

// ShouldSample implements tail-based sampling logic
func (s *TailSampler) ShouldSample(event *WideEvent) bool {
	// Always keep errors
	if event.Status == "error" || event.Status == "failed" || event.Error != nil {
		return rand.Float64() < s.errorSampleRate
	}

	// Always keep slow requests
	if event.DurationMs > s.slowThresholdMs {
		return true
	}

	// Sample successful fast requests at low rate
	return rand.Float64() < s.successRate
}

// NewJobEvent creates a new wide event for a job workflow
func NewJobEvent(jobID, userID string) *WideEvent {
	return &WideEvent{
		Timestamp: time.Now(),
		EventType: "job.workflow",
		JobID:     jobID,
		UserID:    userID,
		Status:    "in_progress",
		Metadata:  make(map[string]interface{}),
		Stages:    make(map[string]*StageInfo),
		startTime: time.Now(),
	}
}

// NewEnrichmentEvent creates a new wide event for an enrichment workflow
func NewEnrichmentEvent(jobID, rowKey, userID string) *WideEvent {
	return &WideEvent{
		Timestamp: time.Now(),
		EventType: "enrichment.workflow",
		JobID:     jobID,
		RowKey:    rowKey,
		UserID:    userID,
		Status:    "in_progress",
		Metadata:  make(map[string]interface{}),
		Stages:    make(map[string]*StageInfo),
		startTime: time.Now(),
	}
}

// NewActivityEvent creates a new wide event for an activity
func NewActivityEvent(activityName, jobID string) *WideEvent {
	return &WideEvent{
		Timestamp: time.Now(),
		EventType: "activity." + activityName,
		JobID:     jobID,
		Status:    "in_progress",
		Metadata:  make(map[string]interface{}),
		startTime: time.Now(),
	}
}

// SetWorkflowInfo adds workflow context to the event
func (e *WideEvent) SetWorkflowInfo(workflowID, runID string) {
	e.WorkflowID = workflowID
	e.RunID = runID
}

// SetServiceInfo adds service identification
func (e *WideEvent) SetServiceInfo(name, version, region string) {
	e.ServiceName = name
	e.ServiceVersion = version
	e.Region = region
}

// StartStage begins tracking a new stage
func (e *WideEvent) StartStage(stage models.RowStage) {
	if e.Stages == nil {
		e.Stages = make(map[string]*StageInfo)
	}
	e.Stages[string(stage)] = &StageInfo{
		StartTime: time.Now(),
		Status:    "in_progress",
	}
}

// CompleteStage marks a stage as successful with optional data
func (e *WideEvent) CompleteStage(stage models.RowStage, data map[string]interface{}) {
	if info, exists := e.Stages[string(stage)]; exists {
		info.DurationMs = time.Since(info.StartTime).Milliseconds()
		info.Status = "success"
		info.Data = data
	}
}

// FailStage marks a stage as failed with error information
func (e *WideEvent) FailStage(stage models.RowStage, err error) {
	if info, exists := e.Stages[string(stage)]; exists {
		info.DurationMs = time.Since(info.StartTime).Milliseconds()
		info.Status = "error"
		errMsg := err.Error()
		info.Error = &errMsg
	}
}

// SetMetadata adds arbitrary metadata to the event
func (e *WideEvent) SetMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// EmitSuccess emits the event as successful
func (e *WideEvent) EmitSuccess(wfCtx workflow.Context) {
	e.DurationMs = time.Since(e.startTime).Milliseconds()
	e.Status = "success"
	e.emit(wfCtx)
}

// EmitError emits the event as failed
func (e *WideEvent) EmitError(wfCtx workflow.Context, err error) {
	e.DurationMs = time.Since(e.startTime).Milliseconds()
	e.Status = "error"
	e.Error = &ErrorInfo{
		Type:      "WorkflowError",
		Message:   err.Error(),
		Retriable: false,
	}
	e.emit(wfCtx)
}

// emit writes the event to logs if it passes sampling
func (e *WideEvent) emit(wfCtx workflow.Context) {
	if !sampler.ShouldSample(e) {
		return
	}

	logger := workflow.GetLogger(wfCtx)

	// Convert to map for structured logging
	eventMap := e.toMap()

	// Emit as structured log
	if e.Status == "error" {
		logger.Error("workflow event", eventMap...)
	} else {
		logger.Info("workflow event", eventMap...)
	}
}

// EmitActivitySuccess emits an activity event as successful
func (e *WideEvent) EmitActivitySuccess(ctx context.Context, data map[string]interface{}) {
	e.DurationMs = time.Since(e.startTime).Milliseconds()
	e.Status = "success"
	if data != nil {
		if e.Metadata == nil {
			e.Metadata = make(map[string]interface{})
		}
		for k, v := range data {
			e.Metadata[k] = v
		}
	}
	e.emitActivity(ctx)
}

// EmitActivityError emits an activity event as failed
func (e *WideEvent) EmitActivityError(ctx context.Context, err error) {
	e.DurationMs = time.Since(e.startTime).Milliseconds()
	e.Status = "error"
	e.Error = &ErrorInfo{
		Type:      "ActivityError",
		Message:   err.Error(),
		Retriable: false,
	}
	e.emitActivity(ctx)
}

// emitActivity writes the activity event to logs
func (e *WideEvent) emitActivity(ctx context.Context) {
	if !sampler.ShouldSample(e) {
		return
	}

	// Convert to structured fields
	fields := e.toSlogAttrs()

	if e.Status == "error" {
		Log.ErrorContext(ctx, "activity event", fields...)
	} else {
		Log.InfoContext(ctx, "activity event", fields...)
	}
}

// toMap converts the event to key-value pairs for workflow logger
func (e *WideEvent) toMap() []interface{} {
	result := []interface{}{}

	add := func(key string, value interface{}) {
		if value != nil {
			switch v := value.(type) {
			case string:
				if v != "" {
					result = append(result, key, v)
				}
			case int:
				if v != 0 {
					result = append(result, key, v)
				}
			case int64:
				if v != 0 {
					result = append(result, key, v)
				}
			default:
				result = append(result, key, v)
			}
		}
	}

	add("timestamp", e.Timestamp.Format(time.RFC3339))
	add("event_type", e.EventType)
	add("status", e.Status)
	add("duration_ms", e.DurationMs)

	add("service_name", e.ServiceName)
	add("service_version", e.ServiceVersion)
	add("region", e.Region)

	add("workflow_id", e.WorkflowID)
	add("run_id", e.RunID)
	add("trace_id", e.TraceID)
	add("request_id", e.RequestID)

	add("job_id", e.JobID)
	add("row_key", e.RowKey)
	add("user_id", e.UserID)
	add("stage", e.Stage)

	add("total_rows", e.TotalRows)
	add("completed", e.Completed)
	add("failed", e.Failed)

	if e.Error != nil {
		add("error_type", e.Error.Type)
		add("error_message", e.Error.Message)
		add("error_code", e.Error.Code)
		add("error_retriable", e.Error.Retriable)
	}

	if len(e.Metadata) > 0 {
		// Add metadata fields directly
		for k, v := range e.Metadata {
			add("meta_"+k, v)
		}
	}

	if len(e.Stages) > 0 {
		// Serialize stages as JSON
		if stagesJSON, err := json.Marshal(e.Stages); err == nil {
			add("stages", string(stagesJSON))
		}
	}

	return result
}

// toSlogAttrs converts the event to slog attributes
func (e *WideEvent) toSlogAttrs() []any {
	attrs := []any{}

	add := func(key string, value any) {
		if value != nil {
			switch v := value.(type) {
			case string:
				if v != "" {
					attrs = append(attrs, slog.String(key, v))
				}
			case int:
				if v != 0 {
					attrs = append(attrs, slog.Int(key, v))
				}
			case int64:
				if v != 0 {
					attrs = append(attrs, slog.Int64(key, v))
				}
			case bool:
				attrs = append(attrs, slog.Bool(key, v))
			default:
				attrs = append(attrs, slog.Any(key, v))
			}
		}
	}

	add("timestamp", e.Timestamp.Format(time.RFC3339))
	add("event_type", e.EventType)
	add("status", e.Status)
	add("duration_ms", e.DurationMs)

	add("service_name", e.ServiceName)
	add("service_version", e.ServiceVersion)
	add("region", e.Region)

	add("workflow_id", e.WorkflowID)
	add("run_id", e.RunID)
	add("trace_id", e.TraceID)
	add("request_id", e.RequestID)

	add("job_id", e.JobID)
	add("row_key", e.RowKey)
	add("user_id", e.UserID)
	add("stage", e.Stage)

	add("total_rows", e.TotalRows)
	add("completed", e.Completed)
	add("failed", e.Failed)

	if e.Error != nil {
		add("error_type", e.Error.Type)
		add("error_message", e.Error.Message)
		add("error_code", e.Error.Code)
		add("error_retriable", e.Error.Retriable)
	}

	if len(e.Metadata) > 0 {
		for k, v := range e.Metadata {
			add("meta_"+k, v)
		}
	}

	if len(e.Stages) > 0 {
		add("stages", e.Stages)
	}

	return attrs
}

// Context key for wide events
type contextKey string

const eventContextKey contextKey = "wide_event_context"

// WithContext stores a wide event in the context
func WithContext(ctx context.Context, event *WideEvent) context.Context {
	return context.WithValue(ctx, eventContextKey, event)
}

// FromContext retrieves a wide event from the context
func FromContext(ctx context.Context) *WideEvent {
	if event, ok := ctx.Value(eventContextKey).(*WideEvent); ok {
		return event
	}
	return nil
}
