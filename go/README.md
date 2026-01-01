# AmpleData Go Enrichment Pipeline

Go implementation of the AmpleData enrichment pipeline with a streaming consumer-producer architecture using goroutines and channels.

## Architecture

This Go implementation eliminates the stage-gate bottleneck from the Python version by using a **streaming pipeline architecture**:

- Each row flows immediately through stages via buffered channels
- Worker pools at each stage process rows concurrently
- No waiting for all rows to complete a stage before moving to the next
- PostgreSQL persistence for fault tolerance

## Project Structure

```
go/
├── cmd/
│   └── server/
│       └── main.go                 # HTTP server entry point
├── internal/
│   ├── api/                        # HTTP API layer
│   ├── config/                     # Configuration
│   ├── models/                     # Data structures
│   ├── pipeline/                   # Streaming pipeline with stages
│   ├── state/                      # State management & PostgreSQL
│   ├── services/                   # External API clients
│   └── enricher/                   # Main orchestrator
├── go.mod
└── go.sum
```

## Prerequisites

1. **PostgreSQL** database
2. **Go 1.21+**
3. **Python Crawl4ai service** running on localhost:8000
4. **API Keys**:
   - Serper API key (for web search)
   - Groq API key (for LLM)

## Setup

### 1. Set up PostgreSQL

The application will automatically create the required tables on startup.

### 2. Start the Python Crawl4ai Service

```bash
cd ../python/crawl4ai_service

pip install -r requirements.txt

uvicorn main:app --host 0.0.0.0 --port 8000
```

### 3. Set Environment Variables

```bash
export DATABASE_URL="postgres://user:password@localhost/ampledata"
export SERVER_ADDR=":8080"
export SERPER_API_KEY="your_serper_key"
export GROQ_API_KEY="your_groq_key"
export CRAWL4AI_URL="http://localhost:8000"
export WORKERS_PER_STAGE=5
export CHANNEL_BUFFER_SIZE=100
```

### 4. Run the Go Server

```bash
go run cmd/server/main.go
```

The server will start on port 8080 (or the port specified in SERVER_ADDR).

## API Endpoints

### Start Enrichment Job

```bash
POST /api/v1/enrich

{
  "row_keys": ["Apple Inc", "Microsoft", "Google"],
  "columns_metadata": [
    {
      "name": "founded_year",
      "type": "number",
      "description": "Year the company was founded"
    },
    {
      "name": "headquarters",
      "type": "string",
      "description": "Location of company headquarters"
    },
    {
      "name": "ceo",
      "type": "string",
      "description": "Current CEO name"
    }
  ],
  "entity_type": "company"
}
```

Response:
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Enrichment started"
}
```

### Get Job Progress

```bash
GET /api/v1/jobs/{jobID}/progress
```

Response:
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_rows": 3,
  "rows_by_stage": {
    "PENDING": 0,
    "SERP_FETCHED": 0,
    "DECISION_MADE": 0,
    "CRAWLED": 0,
    "ENRICHED": 1,
    "COMPLETED": 2
  },
  "started_at": "2025-12-29T02:00:00Z",
  "status": "RUNNING"
}
```

### Cancel Job

```bash
POST /api/v1/jobs/{jobID}/cancel
```

### Get Job Results

```bash
GET /api/v1/jobs/{jobID}/results
```

Response:
```json
[
  {
    "key": "Apple Inc",
    "extracted_data": {
      "founded_year": 1976,
      "headquarters": "Cupertino, California",
      "ceo": "Tim Cook"
    },
    "sources": [
      "https://en.wikipedia.org/wiki/Apple_Inc."
    ]
  }
]
```

## Pipeline Architecture

The streaming pipeline consists of 4 stages connected by buffered channels:

```
Input → [SERP] → [Decision] → [Crawl] → [Extract] → Output
          ↓         ↓            ↓          ↓
       Channel    Channel     Channel    Channel
       (buf:100)  (buf:100)   (buf:100)  (buf:100)

Each stage has 5 worker goroutines processing messages concurrently
```

### Stages

1. **SERP Stage**: Builds search query, fetches Google search results via Serper API
2. **Decision Stage**: Uses Groq LLM to decide which URLs to crawl (or extract from snippets)
3. **Crawl Stage**: Calls Python Crawl4ai service to fetch and parse web content
4. **Extract Stage**: Uses Groq LLM to extract structured data from content

### Key Features

- **Immediate Flow**: Rows move to next stage as soon as processing completes
- **Concurrent Processing**: 5 workers per stage (configurable)
- **State Persistence**: Each transition saved to PostgreSQL
- **Fault Tolerance**: Can resume interrupted jobs
- **Cancellation**: Context-based cancellation across all goroutines

## Configuration

All configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_URL | postgres://localhost/ampledata | PostgreSQL connection string |
| SERVER_ADDR | :8080 | HTTP server address |
| SERPER_API_KEY | - | Serper API key (required) |
| GROQ_API_KEY | - | Groq API key (required) |
| CRAWL4AI_URL | http://localhost:8000 | Crawl4ai service URL |
| WORKERS_PER_STAGE | 5 | Number of workers per pipeline stage |
| CHANNEL_BUFFER_SIZE | 100 | Buffer size for inter-stage channels |

## Development

### Build

```bash
go build -o bin/server cmd/server/main.go
```

### Run

```bash
./bin/server
```

## Comparison with Python Version

| Aspect | Python (AsyncEnricher) | Go (Pipeline) |
|--------|------------------------|---------------|
| Concurrency | asyncio.gather per stage | Goroutines with channels |
| Stage Flow | Stage-gate (wait for all) | Streaming (immediate flow) |
| Latency | High (batch processing) | Low (row-level pipelining) |
| Resource Usage | Limited by asyncio | Native OS threads |
| Scalability | Moderate | High |

## Notes

- The Go server expects column metadata in each request (unlike Python which uses globals)
- Column metadata and entity type can be provided per-request for flexibility
- The pipeline will automatically persist state after each stage transition
- Failed rows are marked as FAILED and don't block other rows
