package logging

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const (
	contextKeyWideEvent contextKey = "wide_event"
	contextKeyTraceID   contextKey = "trace_id"
)

// WideEvent represents a single structured log entry that captures the full lifecycle of a request.
// It is incrementally populated as the request flows through different system components.
type WideEvent struct {
	// Core identifiers
	TraceID   string    `json:"trace_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`

	// Request metadata
	HTTPMethod     string            `json:"http_method,omitempty"`
	HTTPPath       string            `json:"http_path,omitempty"`
	HTTPStatusCode int               `json:"http_status_code,omitempty"`
	HTTPDurationMs int64             `json:"http_duration_ms,omitempty"`
	HTTPHeaders    map[string]string `json:"http_headers,omitempty"`

	// User context
	UserID    string `json:"user_id,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
	UserName  string `json:"user_name,omitempty"`

	// Job context
	JobID       string `json:"job_id,omitempty"`
	JobStatus   string `json:"job_status,omitempty"`
	TotalRows   int    `json:"total_rows,omitempty"`
	RowsSuccess int    `json:"rows_success,omitempty"`
	RowsFailed  int    `json:"rows_failed,omitempty"`

	// Row-level context
	RowKey        string `json:"row_key,omitempty"`
	RowStage      string `json:"row_stage,omitempty"`
	StageDuration int64  `json:"stage_duration_ms,omitempty"`

	// Pipeline context
	PipelineStage   string   `json:"pipeline_stage,omitempty"`
	QueryPatterns   []string `json:"query_patterns,omitempty"`
	URLsCrawled     []string `json:"urls_crawled,omitempty"`
	ExtractedFields []string `json:"extracted_fields,omitempty"`

	// Error tracking
	Error          string `json:"error,omitempty"`
	ErrorStage     string `json:"error_stage,omitempty"`
	ErrorRowKey    string `json:"error_row_key,omitempty"`
	PanicRecovered bool   `json:"panic_recovered,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewWideEvent creates a new WideEvent with a trace ID and timestamp
func NewWideEvent(eventType string) *WideEvent {
	return &WideEvent{
		TraceID:     uuid.New().String(),
		EventType:   eventType,
		Timestamp:   time.Now(),
		HTTPHeaders: make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
}

// WithContext attaches a WideEvent to a context
func WithContext(ctx context.Context, event *WideEvent) context.Context {
	ctx = context.WithValue(ctx, contextKeyWideEvent, event)
	ctx = context.WithValue(ctx, contextKeyTraceID, event.TraceID)
	return ctx
}

// FromContext retrieves the WideEvent from a context
func FromContext(ctx context.Context) *WideEvent {
	if event, ok := ctx.Value(contextKeyWideEvent).(*WideEvent); ok {
		return event
	}
	return nil
}

// GetTraceID retrieves just the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(contextKeyTraceID).(string); ok {
		return traceID
	}
	return ""
}

// Enrich helpers - these allow different parts of the code to enrich the event

func EnrichHTTP(ctx context.Context, method, path string) {
	if event := FromContext(ctx); event != nil {
		event.HTTPMethod = method
		event.HTTPPath = path
	}
}

func EnrichHTTPStatus(ctx context.Context, statusCode int) {
	if event := FromContext(ctx); event != nil {
		event.HTTPStatusCode = statusCode
	}
}

func EnrichHTTPDuration(ctx context.Context, duration time.Duration) {
	if event := FromContext(ctx); event != nil {
		event.HTTPDurationMs = duration.Milliseconds()
	}
}

func EnrichHTTPHeader(ctx context.Context, key, value string) {
	if event := FromContext(ctx); event != nil {
		event.HTTPHeaders[key] = value
	}
}

func EnrichUser(ctx context.Context, userID, email, name string) {
	if event := FromContext(ctx); event != nil {
		event.UserID = userID
		event.UserEmail = email
		event.UserName = name
	}
}

func EnrichJob(ctx context.Context, jobID, status string) {
	if event := FromContext(ctx); event != nil {
		event.JobID = jobID
		event.JobStatus = status
	}
}

func EnrichJobRows(ctx context.Context, total, success, failed int) {
	if event := FromContext(ctx); event != nil {
		event.TotalRows = total
		event.RowsSuccess = success
		event.RowsFailed = failed
	}
}

func EnrichRow(ctx context.Context, rowKey, stage string) {
	if event := FromContext(ctx); event != nil {
		event.RowKey = rowKey
		event.RowStage = stage
	}
}

func EnrichStageDuration(ctx context.Context, duration time.Duration) {
	if event := FromContext(ctx); event != nil {
		event.StageDuration = duration.Milliseconds()
	}
}

func EnrichPipeline(ctx context.Context, stage string, patterns []string) {
	if event := FromContext(ctx); event != nil {
		event.PipelineStage = stage
		event.QueryPatterns = patterns
	}
}

func EnrichURLs(ctx context.Context, urls []string) {
	if event := FromContext(ctx); event != nil {
		event.URLsCrawled = urls
	}
}

func EnrichExtractedFields(ctx context.Context, fields []string) {
	if event := FromContext(ctx); event != nil {
		event.ExtractedFields = fields
	}
}

func EnrichError(ctx context.Context, err error, stage string) {
	if event := FromContext(ctx); event != nil {
		if err != nil {
			event.Error = err.Error()
			event.ErrorStage = stage
		}
	}
}

func EnrichErrorRow(ctx context.Context, rowKey string) {
	if event := FromContext(ctx); event != nil {
		event.ErrorRowKey = rowKey
	}
}

func EnrichPanic(ctx context.Context) {
	if event := FromContext(ctx); event != nil {
		event.PanicRecovered = true
	}
}

func EnrichMetadata(ctx context.Context, key string, value interface{}) {
	if event := FromContext(ctx); event != nil {
		event.Metadata[key] = value
	}
}

// Emit outputs the WideEvent as a structured log
func Emit(ctx context.Context) {
	event := FromContext(ctx)
	if event == nil {
		return
	}

	// Build slog attributes from the event
	attrs := []slog.Attr{
		slog.String("trace_id", event.TraceID),
		slog.String("event_type", event.EventType),
		slog.Time("timestamp", event.Timestamp),
	}

	// HTTP metadata
	if event.HTTPMethod != "" {
		attrs = append(attrs, slog.String("http_method", event.HTTPMethod))
	}
	if event.HTTPPath != "" {
		attrs = append(attrs, slog.String("http_path", event.HTTPPath))
	}
	if event.HTTPStatusCode != 0 {
		attrs = append(attrs, slog.Int("http_status_code", event.HTTPStatusCode))
	}
	if event.HTTPDurationMs != 0 {
		attrs = append(attrs, slog.Int64("http_duration_ms", event.HTTPDurationMs))
	}
	if len(event.HTTPHeaders) > 0 {
		attrs = append(attrs, slog.Any("http_headers", event.HTTPHeaders))
	}

	// User context
	if event.UserID != "" {
		attrs = append(attrs, slog.String("user_id", event.UserID))
	}
	if event.UserEmail != "" {
		attrs = append(attrs, slog.String("user_email", event.UserEmail))
	}
	if event.UserName != "" {
		attrs = append(attrs, slog.String("user_name", event.UserName))
	}

	// Job context
	if event.JobID != "" {
		attrs = append(attrs, slog.String("job_id", event.JobID))
	}
	if event.JobStatus != "" {
		attrs = append(attrs, slog.String("job_status", event.JobStatus))
	}
	if event.TotalRows != 0 {
		attrs = append(attrs, slog.Int("total_rows", event.TotalRows))
	}
	if event.RowsSuccess != 0 {
		attrs = append(attrs, slog.Int("rows_success", event.RowsSuccess))
	}
	if event.RowsFailed != 0 {
		attrs = append(attrs, slog.Int("rows_failed", event.RowsFailed))
	}

	// Row-level context
	if event.RowKey != "" {
		attrs = append(attrs, slog.String("row_key", event.RowKey))
	}
	if event.RowStage != "" {
		attrs = append(attrs, slog.String("row_stage", event.RowStage))
	}
	if event.StageDuration != 0 {
		attrs = append(attrs, slog.Int64("stage_duration_ms", event.StageDuration))
	}

	// Pipeline context
	if event.PipelineStage != "" {
		attrs = append(attrs, slog.String("pipeline_stage", event.PipelineStage))
	}
	if len(event.QueryPatterns) > 0 {
		attrs = append(attrs, slog.Any("query_patterns", event.QueryPatterns))
	}
	if len(event.URLsCrawled) > 0 {
		attrs = append(attrs, slog.Any("urls_crawled", event.URLsCrawled))
	}
	if len(event.ExtractedFields) > 0 {
		attrs = append(attrs, slog.Any("extracted_fields", event.ExtractedFields))
	}

	// Error tracking
	if event.Error != "" {
		attrs = append(attrs, slog.String("error", event.Error))
	}
	if event.ErrorStage != "" {
		attrs = append(attrs, slog.String("error_stage", event.ErrorStage))
	}
	if event.ErrorRowKey != "" {
		attrs = append(attrs, slog.String("error_row_key", event.ErrorRowKey))
	}
	if event.PanicRecovered {
		attrs = append(attrs, slog.Bool("panic_recovered", event.PanicRecovered))
	}

	// Metadata
	if len(event.Metadata) > 0 {
		attrs = append(attrs, slog.Any("metadata", event.Metadata))
	}

	// Emit the log with appropriate level
	level := slog.LevelInfo
	if event.Error != "" || event.PanicRecovered {
		level = slog.LevelError
	}

	slog.LogAttrs(ctx, level, "wide_event", attrs...)
}

// EmitRowEvent creates and emits a row-level event (for pipeline stages)
func EmitRowEvent(ctx context.Context, eventType, jobID, rowKey, stage string, duration time.Duration, err error) {
	// Create a new event for this row processing
	event := &WideEvent{
		TraceID:       GetTraceID(ctx),
		EventType:     eventType,
		Timestamp:     time.Now(),
		JobID:         jobID,
		RowKey:        rowKey,
		RowStage:      stage,
		PipelineStage: stage,
		StageDuration: duration.Milliseconds(),
		Metadata:      make(map[string]interface{}),
	}

	// Inherit user context if available
	if parentEvent := FromContext(ctx); parentEvent != nil {
		event.UserID = parentEvent.UserID
		event.UserEmail = parentEvent.UserEmail
		event.UserName = parentEvent.UserName
	}

	if err != nil {
		event.Error = err.Error()
		event.ErrorStage = stage
		event.ErrorRowKey = rowKey
	}

	// Build attrs
	attrs := []slog.Attr{
		slog.String("trace_id", event.TraceID),
		slog.String("event_type", event.EventType),
		slog.Time("timestamp", event.Timestamp),
		slog.String("job_id", event.JobID),
		slog.String("row_key", event.RowKey),
		slog.String("row_stage", event.RowStage),
		slog.String("pipeline_stage", event.PipelineStage),
		slog.Int64("stage_duration_ms", event.StageDuration),
	}

	if event.UserEmail != "" {
		attrs = append(attrs, slog.String("user_email", event.UserEmail))
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", event.Error))
		attrs = append(attrs, slog.String("error_stage", event.ErrorStage))
		slog.LogAttrs(ctx, slog.LevelError, "row_event", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "row_event", attrs...)
	}
}
