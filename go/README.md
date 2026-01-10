# AmpleData Go Enrichment Pipeline

A fast data enrichment tool that takes a list of items and fills in missing information using web search and AI.

## How It Works

The pipeline processes items one at a time through 4 stages:

1. **Search** - Finds web results for each item
2. **Decide** - Picks which websites to visit
3. **Crawl** - Gets content from those websites
4. **Extract** - Pulls out the information you need

If no data is found, the pipeline will try again with a different website.

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

Check `./test_enrichment.sh`
