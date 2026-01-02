# AmpleData Go Enrichment Pipeline

A fast data enrichment tool that takes a list of items and fills in missing information using web search and AI.

## How It Works

The pipeline processes items one at a time through 4 stages:

1. **Search** - Finds web results for each item
2. **Decide** - Picks which websites to visit
3. **Crawl** - Gets content from those websites
4. **Extract** - Pulls out the information you need

Each item moves to the next stage as soon as it's done, so no waiting around.

## Setup

### 1. Install PostgreSQL

PostgreSQL database is required. Tables are created automatically when you start the server.

### 2. Start the Python Service

```bash
cd ../python/crawl4ai_service
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8000
```

### 3. Set Your API Keys

```bash
export DATABASE_URL="postgres://user:password@localhost/ampledata"
export SERVER_ADDR=":8080"
export SERPER_API_KEY="your_serper_key"
export GROQ_API_KEY="your_groq_key"
export CRAWL4AI_URL="http://localhost:8000"
export WORKERS_PER_STAGE=5
export CHANNEL_BUFFER_SIZE=100
```

### 4. Start the Server

```bash
go run cmd/server/main.go
```

Server runs on port 8080.

## Using the API

### Start a Job

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
    }
  ],
  "entity_type": "company"
}
```

Returns:
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Enrichment started"
}
```

### Check Progress

```bash
GET /api/v1/jobs/{jobID}/progress
```

Shows how many items are at each stage.

### Get Results

```bash
GET /api/v1/jobs/{jobID}/results
```

Returns:
```json
[
  {
    "key": "Apple Inc",
    "extracted_data": {
      "founded_year": 1976,
      "headquarters": "Cupertino, California"
    },
    "sources": ["https://en.wikipedia.org/wiki/Apple_Inc."]
  }
]
```

### Cancel a Job

```bash
POST /api/v1/jobs/{jobID}/cancel
```

## Settings

Change these with environment variables:

| Name | Default | What It Does |
|------|---------|-------------|
| DATABASE_URL | postgres://localhost/ampledata | Where to store data |
| SERVER_ADDR | :8080 | Port to run on |
| SERPER_API_KEY | - | Search API key |
| GROQ_API_KEY | - | AI API key |
| CRAWL4AI_URL | http://localhost:8000 | Web scraper location |
| WORKERS_PER_STAGE | 5 | How many things to process at once |
| CHANNEL_BUFFER_SIZE | 100 | How many items to hold in memory |

## What You Need

- Go 1.21 or newer
- PostgreSQL database
- Serper API key (for web search)
- Groq API key (for AI)
- Python Crawl4ai service running
