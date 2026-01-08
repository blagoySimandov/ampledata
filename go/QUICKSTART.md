# Temporal Migration - Quick Start Guide

## Prerequisites

1. **Temporal Server** running locally
2. **PostgreSQL** database
3. **Required API Keys**:
   - SERPER_API_KEY (for web search)
   - GROQ_API_KEY (for decision making)
   - GEMINI_API_KEY (for pattern generation and extraction)
4. **Crawl4ai service** running on port 8000

## Starting Temporal Server (Development)

```bash
# Install Temporal CLI if not already installed
# See: https://docs.temporal.io/cli

# Start local dev server
temporal server start-dev
```

This starts:
- Server at `localhost:7233`
- Web UI at `http://localhost:8233`
- Metrics at `http://localhost:52781/metrics`

## Running the Application

```bash
# Navigate to the Go directory
cd go

# Set required environment variables
export DATABASE_URL_ENRICH="postgres://enrichment:enrichment@localhost:5432/enrichment?sslmode=disable"
export SERPER_API_KEY="your-serper-key"
export GROQ_API_KEY="your-groq-key"
export GEMINI_API_KEY="your-gemini-key"
export CRAWL4AI_URL="http://localhost:8000"
export TEMPORAL_HOST_PORT="localhost:7233"
export TEMPORAL_NAMESPACE="default"
export TEMPORAL_TASK_QUEUE="ampledata-enrichment"

# Run the server
go run cmd/server/main.go
```

The server will:
1. Connect to PostgreSQL
2. Connect to Temporal server
3. Start a Temporal worker
4. Start HTTP API server on `:8080`

## Using the System

### 1. Upload CSV File
```bash
curl -X POST http://localhost:8080/api/v1/enrichment-signed-url \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "companies.csv"
  }'
```

### 2. Start Enrichment Job
```bash
curl -X POST http://localhost:8080/api/v1/jobs/{jobID}/start \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "key_column": "company_name",
    "columns_metadata": [
      {
        "name": "industry",
        "type": "string",
        "description": "The industry the company operates in"
      },
      {
        "name": "employee_count",
        "type": "number",
        "description": "Number of employees"
      }
    ],
    "entity_type": "company"
  }'
```

### 3. Check Progress
```bash
curl http://localhost:8080/api/v1/jobs/{jobID}/progress \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. Get Results
```bash
curl http://localhost:8080/api/v1/jobs/{jobID}/results \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Monitoring with Temporal UI

Open `http://localhost:8233` in your browser to:

- **View Workflows**: See all running and completed workflows
- **Track Progress**: Monitor enrichment progress for each row
- **Debug Failures**: View detailed error traces and retry attempts
- **Inspect State**: See workflow execution history and activity results

## Architecture Overview

```
HTTP API Request
      ↓
JobWorkflow (Orchestrator)
      ↓
EnrichmentWorkflow (per row) ← → Activities
      ↓                           - GeneratePatterns
PostgreSQL State                  - SerpFetch
      ↓                           - MakeDecision
Results API                       - Crawl
                                  - Extract
                                  - AnalyzeFeedback
```

## Key Features

### Durable Execution
- Workflows survive process crashes
- Automatic state persistence
- Resume from any point

### Observability
- Complete execution history in Temporal UI
- Activity-level logging
- Performance metrics

### Scalability
- Horizontal scaling of workers
- Parallel row processing
- Independent service scaling

### Feedback Loop Ready
The architecture supports cyclic enrichment:
1. Initial enrichment identifies low-confidence data
2. Feedback analysis generates targeted queries
3. Workflow continues with adjusted parameters
4. Process repeats until confidence threshold met

## Troubleshooting

### "Failed to create Temporal client"
- Ensure Temporal server is running: `temporal server start-dev`
- Check `TEMPORAL_HOST_PORT` is set correctly

### "Worker not processing tasks"
- Verify worker logs for errors
- Check task queue name matches: `TEMPORAL_TASK_QUEUE=ampledata-enrichment`
- Ensure all API keys are valid

### "Database connection error"
- Verify PostgreSQL is running
- Check `DATABASE_URL_ENRICH` is correct
- Run migrations if needed: `go run cmd/migrate/main.go`

### Workflows stuck in queue
- Check worker is running and connected
- View worker status in Temporal UI
- Verify task queue configuration

## Development Tips

### Viewing Workflow History
```bash
temporal workflow show --workflow-id job-{jobID}
```

### Canceling a Workflow
```bash
temporal workflow cancel --workflow-id job-{jobID}
```

### Listing All Workflows
```bash
temporal workflow list
```

### Tailing Worker Logs
The server outputs detailed logs including:
- Activity execution
- State transitions
- Error details
- Performance metrics

## Next Steps

1. **Test the system** with sample data
2. **Monitor in Temporal UI** to understand workflow execution
3. **Implement feedback loops** for iterative enrichment
4. **Scale workers** based on load
5. **Deploy to production** with proper Temporal server setup

For more details, see `TEMPORAL_MIGRATION.md`.
