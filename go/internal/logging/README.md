# Structured Logging System

This package implements a clean, context-based structured logging system for the enrichment pipeline.

## Overview

The logging system uses **WideEvent** - a structured log entry that captures the complete lifecycle of a request. Events are incrementally populated as requests flow through different system components, without requiring god objects to be passed around.

## Key Features

- **Context-Based**: Log data is attached to `context.Context`, so no god objects need to be passed
- **Incremental**: Each layer enriches the event with its own data
- **Structured**: Uses Go's `log/slog` for JSON-formatted logs
- **Dual-Purpose**: Acts as both logging and distributed tracing
- **Clean**: One log per request at the HTTP level, plus per-row events at the pipeline level

## Architecture

### Event Types

1. **`http_request`** - One event per HTTP request
   - Emitted at the end of each HTTP request
   - Contains: HTTP metadata, user info, job info, errors

2. **`job_processing`** - One event per enrichment job
   - Emitted at the end of pipeline execution
   - Contains: job stats, row counts, pipeline stages

3. **`row_serp_completed/failed`** - Per-row SERP stage events
4. **`row_decision_completed/failed`** - Per-row decision stage events
5. **`row_crawl_completed/failed/skipped`** - Per-row crawl stage events
6. **`row_extract_completed/failed`** - Per-row extract stage events

### Data Flow

```
HTTP Request → Middleware creates WideEvent
    ↓
Auth Middleware → Enriches with user data
    ↓
Handler → Enriches with job data
    ↓
Middleware → Emits HTTP request log
```

```
Pipeline Start → Creates job-level WideEvent
    ↓
Each Stage → Emits per-row events
    ↓
Pipeline End → Emits job completion log
```

## Usage

### In Middleware

```go
// Create event
event := logging.NewWideEvent("http_request")
ctx := logging.WithContext(r.Context(), event)

// Enrich throughout the request
logging.EnrichHTTP(ctx, r.Method, r.RequestURI)
logging.EnrichUser(ctx, userID, email, name)
logging.EnrichJob(ctx, jobID, status)

// Emit at the end
logging.Emit(ctx)
```

### In Pipeline Stages

```go
start := time.Now()

// ... do work ...

duration := time.Since(start)
logging.EmitRowEvent(ctx, "row_serp_completed", jobID, rowKey, stage, duration, nil)
```

### Enrichment Functions

- `EnrichHTTP(ctx, method, path)` - HTTP request metadata
- `EnrichHTTPStatus(ctx, statusCode)` - Response status
- `EnrichHTTPDuration(ctx, duration)` - Request duration
- `EnrichUser(ctx, userID, email, name)` - User information
- `EnrichJob(ctx, jobID, status)` - Job information
- `EnrichJobRows(ctx, total, success, failed)` - Row statistics
- `EnrichRow(ctx, rowKey, stage)` - Row-level data
- `EnrichPipeline(ctx, stage, patterns)` - Pipeline stage info
- `EnrichError(ctx, err, stage)` - Error information
- `EnrichMetadata(ctx, key, value)` - Custom metadata

## Log Output

All logs are emitted as JSON to stdout using `slog.JSONHandler`.

### Example HTTP Request Log

```json
{
  "level": "INFO",
  "msg": "wide_event",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "http_request",
  "timestamp": "2025-01-06T12:34:56Z",
  "http_method": "POST",
  "http_path": "/api/v1/jobs/abc123/start",
  "http_status_code": 200,
  "http_duration_ms": 1250,
  "user_id": "user_01",
  "user_email": "user@example.com",
  "user_name": "John Doe",
  "job_id": "abc123",
  "job_status": "running",
  "total_rows": 100
}
```

### Example Row Event Log

```json
{
  "level": "INFO",
  "msg": "row_event",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "row_serp_completed",
  "timestamp": "2025-01-06T12:34:57Z",
  "job_id": "abc123",
  "row_key": "ACME Corp",
  "row_stage": "serp_fetched",
  "pipeline_stage": "serp_fetched",
  "stage_duration_ms": 342,
  "user_email": "user@example.com"
}
```

## Design Principles

1. **No God Objects**: Data is enriched via context, not passed as parameters
2. **One Log Per Request**: HTTP requests emit exactly one structured log at the end
3. **Row-Level Tracing**: Pipeline emits one event per row per stage for detailed tracing
4. **Fail-Safe**: If context is missing a WideEvent, enrichment functions are no-ops
5. **Clean**: Old `log.Printf` statements have been completely removed

## Implementation Files

- `wide_event.go` - Core logging types and functions
- See usage in:
  - `internal/api/middleware.go` - HTTP logging
  - `internal/auth/middleware.go` - User enrichment
  - `internal/api/handlers.go` - Job enrichment
  - `internal/pipeline/pipeline.go` - Job-level events
  - `internal/pipeline/stage_*.go` - Row-level events
