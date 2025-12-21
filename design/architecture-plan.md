# Data Enrichment Framework - Design Plan

## Overview

Building a Python-based data enrichment framework where users provide keys and columns to enrich, and the system orchestrates parallel web search, crawling (using crawl4ai), and LLM-based enrichment.

## Core Design Decisions

- **Orchestrator Name**: `EnrichmentPipeline` (pure pipeline executor, not intelligent agent)
- **Execution Model**: Asynchronous with immediate return and progress tracking
- **Parallelization**: Configurable parallel batches processing
- **Error Handling**: Retry N times (configurable), then skip failed rows with state preservation
- **State Management**: Persist state at job/batch/row levels to avoid losing work

## Architecture Components

### 1. Data Models (ampledata/models.py)

#### Input Model

```python
class EnrichmentRequest:
    keys: List[str]           # Columns that uniquely identify rows (what to search on)
    columns_to_enrich: List[str]      # Target columns for enrichment
    search_context: Dict[str, Any]    # Optional metadata for queries
```

#### Configuration Model

```python
class PipelineConfig:
    batch_size: int                   # K rows per batch
    max_parallel_batches: int         # Concurrent batch limit
    max_retries: int                  # Retry attempts before skip
    retry_delay_seconds: float        # Wait between retries
    enable_state_persistence: bool
    state_persistence_path: str
    crawl_config: CrawlConfig

class CrawlConfig:
    max_depth: int
    max_results_per_query: int
    timeout_seconds: int
    strategy: CrawlStrategy           # depth_first | breadth_first | selective
    javascript_enabled: bool
```

#### State Model (Three-Level Hierarchy)

```python
class JobState:
    job_id: str
    status: JobStatus                 # pending | running | completed | failed
    request: EnrichmentRequest
    config: PipelineConfig
    batches: Dict[str, BatchState]
    rows: Dict[str, RowState]
    total_rows: int
    completed_rows: int
    failed_rows: int
    created_at, started_at, completed_at: datetime

class BatchState:
    batch_id: str
    status: BatchStatus               # pending | processing | completed | failed
    row_ids: List[str]
    started_at, completed_at: datetime

class RowState:
    row_id: str
    status: RowStatus                 # pending | searching | crawling | enriching | completed | failed
    retry_count: int
    last_error: Optional[str]
    search_results: Optional[List[Dict]]
    crawl_results: Optional[List[Dict]]
    enriched_data: Optional[Dict]
    started_at, completed_at: datetime
```

#### Result Model

```python
class EnrichmentResult:
    job_id: str
    status: JobStatus
    enriched_rows: List[EnrichedRow]
    total_rows, successful_rows, failed_rows, skipped_rows: int
    started_at, completed_at: datetime
    total_duration_seconds: float
    config: PipelineConfig

class EnrichedRow:
    row_id: str
    original_data: Dict[str, Any]
    enriched_columns: Dict[str, Any]
    metadata: EnrichmentMetadata      # sources, confidence, processing_time, etc.
    status: RowStatus
    error: Optional[str]
```

### 2. Component Interfaces (ampledata/interfaces.py)

All components use async abstract interfaces:

```python
class QueryStringBuilder(ABC):
    async def build_queries(
        rows: List[Dict],
        columns_to_enrich: List[str],
        primary_keys: List[str]
    ) -> List[str]

class SearchService(ABC):
    async def search(queries: List[str]) -> List[Dict[str, Any]]

class CrawlDecisionMaker(ABC):
    async def select_urls_to_crawl(
        search_results: List[Dict],
        max_urls: int
    ) -> List[str]

class WebCrawler(ABC):
    async def crawl(
        urls: List[str],
        config: CrawlConfig
    ) -> List[Dict[str, Any]]

    # Implementation will use crawl4ai library

class LLMEnricher(ABC):
    async def enrich_batch(
        crawl_data: List[Dict],
        batch_rows: List[Dict],
        columns_to_enrich: List[str]
    ) -> List[Dict[str, Any]]

class StateStore(ABC):
    async def save_job_state(job_state: JobState) -> None
    async def load_job_state(job_id: str) -> Optional[JobState]
    async def list_job_ids(status: Optional[JobStatus]) -> List[str]
    async def delete_job_state(job_id: str) -> None
```

### 3. Pipeline Flow

Based on diagram analysis, the flow per batch of K rows:

```
1. BATCH PREPARATION
   ├─> Create batch of K rows from M total rows
   ├─> Generate row_ids from primary keys
   └─> Initialize RowState for each row

2. QUERY BUILDING STAGE
   ├─> QueryStringBuilder.build_queries(K rows, columns, keys)
   ├─> Update RowState.status = SEARCHING for all K rows
   └─> Persist state

3. SEARCH STAGE
   ├─> SearchService.search(queries)
   ├─> Store RowState.search_results for each row
   └─> Persist state

4. CRAWL DECISION STAGE
   ├─> CrawlDecisionMaker.select_urls_to_crawl(search_results, max_urls)
   ├─> Update RowState.status = CRAWLING
   └─> Persist state

5. CRAWLING STAGE
   ├─> WebCrawler.crawl(selected_urls, crawl_config)
   ├─> Store RowState.crawl_results for each row
   └─> Persist state

6. BATCH-LEVEL LLM ENRICHMENT STAGE
   ├─> Collect all crawl results for K rows in batch
   ├─> LLMEnricher.enrich_batch(all_crawl_data, K_rows, columns)
   ├─> Store RowState.enriched_data for each row
   ├─> Update RowState.status = COMPLETED
   └─> Persist state

ERROR HANDLING (any stage):
   ├─> Increment RowState.retry_count
   ├─> Store RowState.last_error
   ├─> If retry_count < max_retries:
   │   ├─> Wait retry_delay_seconds
   │   └─> Retry from failed stage
   └─> Else:
       ├─> RowState.status = FAILED
       └─> Continue processing other rows
```

**Key Pipeline Characteristics:**

- LLM Enricher operates at **batch-level** (all K rows together), not per-row
- Each stage persists state before proceeding
- Stages are sequential within a batch but batches run in parallel
- Future enhancement: Iterative refinement loop (LLM → decide if more crawling needed → repeat)

### 4. Orchestrator Responsibilities (ampledata/pipeline.py)

#### EnrichmentPipeline Class Owns

1. **Job Lifecycle**: Create, track, cancel, resume jobs
2. **Batch Coordination**: Divide rows into batches, schedule based on parallelism
3. **State Persistence**: Save after each stage, enable recovery
4. **Progress Tracking**: Aggregate metrics, emit events
5. **Error Recovery**: Retry logic, failure handling

#### EnrichmentPipeline Delegates

1. **Query Building** → QueryStringBuilder
2. **Search** → SearchService (wraps SERPER)
3. **Crawl Decisions** → CrawlDecisionMaker
4. **Crawling** → WebCrawler (wraps crawl4ai)
5. **Enrichment** → LLMEnricher (wraps litellm)

#### Parallelization Strategy

- **Batch-Level**: Multiple batches run concurrently (controlled by `max_parallel_batches`)
- **Within-Batch**: Rows in same batch processed in parallel where possible
- **Stage-Level**: Some operations can be batched (e.g., SERPER bulk API, parallel crawls)

### 5. State Persistence (ampledata/state_store.py)

#### FileSystemStateStore Implementation

- Stores each job as JSON file: `{state_path}/{job_id}.json`
- Atomic writes to prevent corruption
- Async I/O for performance

#### What Gets Persisted

- Complete JobState after each stage
- All intermediate results (search_results, crawl_results)
- Allows granular recovery from any failure point

#### Recovery Mechanism

```python
async def resume_job(job_id: str):
    1. Load JobState from storage
    2. Identify incomplete rows (status != COMPLETED)
    3. Reconstruct batches from incomplete rows
    4. Resume processing from last checkpoint
    5. Skip stages with existing results
```

### 6. API Design (ampledata/api.py)

#### REST API Endpoints

```
POST   /enrichment/jobs              # Submit job, returns job_id
GET    /enrichment/jobs/{id}/status  # Get progress (percent, ETA, counts)
GET    /enrichment/jobs/{id}/result  # Get enriched data (when completed)
POST   /enrichment/jobs/{id}/cancel  # Cancel running job
POST   /enrichment/jobs/{id}/resume  # Resume failed/paused job
GET    /enrichment/jobs              # List all jobs
```

#### WebSocket (optional)

```
WS     /enrichment/jobs/{id}/stream  # Real-time progress events
```

#### Response Format

```json
{
  "job_id": "uuid",
  "status": "running",
  "progress": {
    "total": 1000,
    "completed": 450,
    "failed": 10,
    "skipped": 5,
    "percent": 45.0
  },
  "started_at": "2025-12-20T10:00:00Z",
  "estimated_completion": "2025-12-20T10:15:00Z"
}
```

## Project Structure

```
ampledata/
├── ampledata/
│   ├── __init__.py
│   ├── models.py              # Pydantic models (Request, Config, State, Result)
│   ├── interfaces.py          # Abstract component interfaces
│   ├── pipeline.py            # EnrichmentPipeline orchestrator
│   ├── state_store.py         # StateStore implementations
│   ├── api.py                 # FastAPI REST endpoints
│   ├── components/
│   │   ├── __init__.py
│   │   ├── query_builder.py  # QueryStringBuilder implementation
│   │   ├── search.py          # SearchService (SERPER wrapper)
│   │   ├── crawler.py         # WebCrawler (crawl4ai wrapper)
│   │   ├── decision.py        # CrawlDecisionMaker
│   │   └── enricher.py        # LLMEnricher (litellm wrapper)
│   └── utils/
│       ├── __init__.py
│       ├── logging.py         # Structured logging setup
│       └── metrics.py         # Observability helpers
├── tests/
│   ├── unit/
│   ├── integration/
│   └── conftest.py
├── pyproject.toml
└── README.md
```

## Critical Implementation Files

1. **ampledata/models.py**: All Pydantic models for type safety
2. **ampledata/interfaces.py**: Abstract base classes defining component contracts
3. **ampledata/pipeline.py**: EnrichmentPipeline orchestrator with core execution logic
4. **ampledata/state_store.py**: StateStore with FileSystemStateStore implementation
5. **ampledata/api.py**: FastAPI application for job management

## Future Enhancements

1. **Iterative Refinement**: LLM analyzes crawl results, determines if more crawling needed, triggers additional searches
2. **Alternative State Stores**: Redis (distributed), PostgreSQL (ACID), S3 (cloud-native)
3. **Smart Query Building**: LLM-assisted query generation
4. **Result Quality Scoring**: LLM evaluates enrichment confidence
5. **Cost Optimization**: Cache search/crawl results, deduplicate queries

## Design Principles

1. **Async-First**: All I/O operations use async/await
2. **Fail-Safe**: State persistence prevents work loss
3. **Observable**: Structured logging, progress events, metrics
4. **Testable**: Abstract interfaces enable mocking
5. **Extensible**: Swap implementations (different search providers, LLMs, etc.)
6. **Simple**: Pure orchestrator, no complex decision-making (for now)
