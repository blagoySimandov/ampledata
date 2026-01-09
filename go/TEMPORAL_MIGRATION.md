# Temporal.io Migration Guide

This document describes the migration of the AmpleData enrichment pipeline from a channel-based Go pipeline to Temporal.io workflows.

## Architecture Overview

### Previous Architecture (Channel-Based)
```
CSV Upload → Row Keys → Channel Pipeline
                       ↓
                    [SERP Stage]
                       ↓
                   [Decision Stage]
                       ↓
                    [Crawl Stage]
                       ↓
                   [Extract Stage]
                       ↓
                    PostgreSQL
```

### New Architecture (Temporal-Based)
```
CSV Upload → Row Keys → Job Workflow
                       ↓
            Spawns Child Workflows (one per row)
                       ↓
            [Enrichment Workflow per Row]
              ├─ SERP Fetch Activity
              ├─ Decision Activity
              ├─ Crawl Activity
              ├─ Extract Activity
              └─ Feedback Analysis Activity
                       ↓
              State Update Activities
                       ↓
                    PostgreSQL
```

## Key Components

### 1. Workflows

#### **JobWorkflow** (`internal/temporal/workflows/job.go`)
- **Purpose**: Orchestrates enrichment for all rows in a job
- **Responsibilities**:
  - Initializes job in database
  - Generates query patterns
  - Spawns child workflows for each row
  - Tracks overall progress
  - Completes the job

#### **EnrichmentWorkflow** (`internal/temporal/workflows/enrichment.go`)
- **Purpose**: Processes a single row through all enrichment stages
- **Responsibilities**:
  - SERP fetch
  - Decision making
  - Web crawling
  - Data extraction
  - Feedback analysis (for future feedback loops)
  - State updates at each stage

### 2. Activities

All activities are implemented in `internal/temporal/activities/activities.go`:

- **GeneratePatterns**: Creates search query patterns using Gemini AI
- **SerpFetch**: Performs web search using Serper API
- **MakeDecision**: Analyzes SERP results and decides what to crawl
- **Crawl**: Fetches content from selected URLs
- **Extract**: Extracts structured data using Gemini AI
- **UpdateState**: Persists row state to PostgreSQL
- **AnalyzeFeedback**: Analyzes results for feedback needs (future use)
- **InitializeJob**: Creates initial row states
- **CompleteJob**: Marks job as completed

### 3. Worker

The Temporal worker (`internal/temporal/worker/worker.go`) registers all workflows and activities, and processes tasks from the task queue.

### 4. Client

The Temporal client (`internal/temporal/client/client.go`) connects to the Temporal server and is used to start workflows.

### 5. Enricher

The new `TemporalEnricher` (`internal/enricher/temporal_enricher.go`) replaces the old channel-based enricher and provides the same interface:
- `Enrich()`: Starts a job workflow
- `GetProgress()`: Returns job progress
- `Cancel()`: Cancels a running job
- `GetResults()`: Returns enrichment results

## Configuration

New environment variables for Temporal:

```bash
# Temporal server configuration
TEMPORAL_HOST_PORT=localhost:7233    # Default: localhost:7233
TEMPORAL_NAMESPACE=default           # Default: default
TEMPORAL_TASK_QUEUE=ampledata-enrichment  # Default: ampledata-enrichment
```

All other existing environment variables remain the same.

## Starting Temporal Server

### Development (Local)

Use the Temporal CLI to start a local dev server:

```bash
temporal server start-dev
```

This starts:
- Temporal server on `localhost:7233`
- Web UI on `http://localhost:8233`
- Metrics on `http://localhost:52781/metrics`

### Production

For production, you should:
1. Deploy Temporal Server (self-hosted or use Temporal Cloud)
2. Configure connection settings via environment variables
3. Set up proper authentication and authorization

## Running the Application

```bash
# Set environment variables
export DATABASE_URL_ENRICH="postgres://enrichment:enrichment@localhost:5432/enrichment?sslmode=disable"
export SERPER_API_KEY="your-key"
export GROQ_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
export CRAWL4AI_URL="http://localhost:8000"
export TEMPORAL_HOST_PORT="localhost:7233"

# Run the server
cd go
go run cmd/server/main.go
```

The server will:
1. Connect to PostgreSQL
2. Connect to Temporal at `localhost:7233`
3. Start a Temporal worker
4. Start the HTTP API server on `:8080`

## Benefits of Temporal Migration

### 1. **Durability**
- Workflows are durable and survive process crashes
- Automatic state persistence
- No need to manually manage state transitions

### 2. **Observability**
- Built-in UI for monitoring workflows
- Complete execution history
- Easy debugging with workflow replay

### 3. **Scalability**
- Workers can be scaled independently
- Automatic load balancing across workers
- Support for distributed execution

### 4. **Reliability**
- Automatic retries with configurable policies
- Guaranteed execution semantics
- Graceful handling of transient failures

### 5. **Flexibility**
- Easy to add new stages
- Support for complex workflow patterns (loops, conditionals, parallel execution)
- Built-in support for long-running operations

### 6. **Feedback Loop Support**
- Workflows can be continued with new inputs
- Support for cyclic execution patterns
- Easy to implement retry logic based on confidence scores

## Future: Feedback Loop Implementation

The architecture is designed to support feedback loops for rows with:
- Missing columns
- Low confidence scores
- Failed extraction

### Planned Implementation

In `EnrichmentWorkflow`, after feedback analysis:

```go
if feedbackOutput.NeedsFeedback && input.RetryCount < maxRetries {
    // Generate new query patterns targeting missing columns
    newPatterns := generateFeedbackPatterns(
        feedbackOutput.MissingColumns,
        feedbackOutput.LowConfidenceColumns,
    )

    // Continue as new with updated input
    return workflow.NewContinueAsNewError(ctx, EnrichmentWorkflow, EnrichmentWorkflowInput{
        JobID:           input.JobID,
        RowKey:          input.RowKey,
        ColumnsMetadata: input.ColumnsMetadata,
        QueryPatterns:   newPatterns,
        RetryCount:      input.RetryCount + 1,
    })
}
```

This creates a cyclic loop where:
1. Initial enrichment runs with default patterns
2. If results are unsatisfactory, feedback is analyzed
3. New patterns are generated targeting missing/low-confidence data
4. Workflow continues with new patterns
5. Process repeats until confidence threshold is met or max retries reached

## Monitoring

### Temporal Web UI

Access the Temporal UI at `http://localhost:8233` to:
- View running workflows
- See workflow execution history
- Debug failed workflows
- Monitor task queue metrics

### Workflow ID Format

- Job workflows: `job-{jobID}`
- Row workflows: Auto-generated by Temporal

### Querying Workflows

```bash
# List all workflows
temporal workflow list

# Describe a workflow
temporal workflow describe --workflow-id job-{jobID}

# Show workflow history
temporal workflow show --workflow-id job-{jobID}
```

## Troubleshooting

### Connection Errors

If you see "failed to create Temporal client":
1. Ensure Temporal server is running: `temporal server start-dev`
2. Check the `TEMPORAL_HOST_PORT` environment variable
3. Verify network connectivity to Temporal server

### Worker Not Processing Tasks

If workflows are stuck:
1. Check worker logs for errors
2. Verify worker is registered with correct task queue
3. Ensure all required services (Serper, Groq, Gemini, Crawl4ai) are accessible

### State Not Updating

If PostgreSQL state is not updating:
1. Check database connectivity
2. Verify state update activities are executing
3. Review activity logs for errors

## API Compatibility

The HTTP API remains unchanged. All existing endpoints work the same way:
- `POST /api/v1/enrichment-signed-url`
- `POST /api/v1/jobs/{jobID}/start`
- `GET /api/v1/jobs`
- `GET /api/v1/jobs/{jobID}/progress`
- `GET /api/v1/jobs/{jobID}/results`
- `POST /api/v1/jobs/{jobID}/cancel`

## Migration Checklist

- [x] Add Temporal SDK dependency
- [x] Create Temporal client setup
- [x] Implement activities for each pipeline stage
- [x] Create EnrichmentWorkflow for row processing
- [x] Create JobWorkflow for job orchestration
- [x] Implement worker registration
- [x] Update StateManager to support workflow IDs
- [x] Create TemporalEnricher
- [x] Update main server to use Temporal
- [x] Add configuration for Temporal settings
- [x] Design feedback loop architecture
- [x] Document migration

## Next Steps

1. **Test the Migration**: Run end-to-end tests with the new Temporal-based system
2. **Performance Tuning**: Adjust worker counts and activity timeouts
3. **Implement Feedback Loop**: Add the cyclic feedback mechanism
4. **Add Monitoring**: Set up metrics and alerting
5. **Production Deployment**: Deploy Temporal server and migrate traffic

## Code Structure

```
internal/
├── temporal/
│   ├── activities/
│   │   └── activities.go        # All activity implementations
│   ├── workflows/
│   │   ├── enrichment.go        # Row enrichment workflow
│   │   └── job.go               # Job orchestration workflow
│   ├── worker/
│   │   └── worker.go            # Worker registration
│   └── client/
│       └── client.go            # Temporal client setup
├── enricher/
│   ├── enricher.go              # Old channel-based enricher (kept for reference)
│   └── temporal_enricher.go    # New Temporal-based enricher
├── state/
│   └── manager.go               # Updated with workflow ID tracking
├── config/
│   └── config.go                # Updated with Temporal config
└── ...
```

## Additional Resources

- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Temporal Patterns](https://docs.temporal.io/dev-guide/go/features)
- [Temporal Best Practices](https://docs.temporal.io/kb)
