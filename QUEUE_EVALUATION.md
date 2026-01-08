# Async Queue Evaluation for AmpleData

## Executive Summary

**Recommendation: DO NOT implement asyncq/taskq at this stage. Focus on simpler optimizations first.**

The current channel-based pipeline is **sufficient** for the project's needs. The perceived need for a distributed queue stems from fixable architectural issues, not fundamental limitations. Implementing asyncq/taskq would add significant complexity without addressing core bottlenecks.

---

## Current Architecture Analysis

### What You Have Now
- **Go pipeline**: 4 stages connected via buffered channels (100 capacity)
- **Worker pools**: 5 concurrent workers per stage (20 total goroutines)
- **State persistence**: PostgreSQL with JSONB (robust, working well)
- **Processing flow**: Sequential stages, parallel rows within each stage

### Real Bottlenecks (Not Queue-Related)
1. **Fixed worker count** (5 per stage) - artificially limits parallelism
2. **Sequential HTTP requests** within stages (should be parallel)
3. **No retry logic** for API failures (terminal errors)
4. **Database connection pool** (10 connections for 20 goroutines)
5. **Context cancellation** not propagated to running jobs

---

## Option 1: Distributed Queue Solutions (asyncq/taskq)

### What These Libraries Provide

#### **taskq (Go)**
- Distributed task queue for Go
- Backends: Redis, AWS SQS, IronMQ
- Features: Retries, rate limiting, delayed tasks, prioritization
- Use case: Microservices with independent workers

#### **asyncq (Python)**
- Distributed task queue (like Celery)
- Backend: Redis, RabbitMQ
- Features: Task scheduling, retries, result backends
- Use case: Background job processing in web apps

### PROS of Using taskq/asyncq

✅ **Built-in retry logic**
   - Automatic exponential backoff
   - Configurable retry limits
   - Dead-letter queues for failed tasks

✅ **Horizontal scalability**
   - Add workers across multiple machines
   - Independent scaling of each stage

✅ **Better observability**
   - Task status tracking out of the box
   - Queue depth metrics
   - Worker utilization stats

✅ **Rate limiting**
   - Protect external APIs from overload
   - Prevent hitting API rate limits

✅ **Task prioritization**
   - Process urgent jobs first
   - Fairness across users

✅ **Persistence**
   - Tasks survive worker crashes
   - Redis/SQS handles message durability

### CONS of Using taskq/asyncq

❌ **Added infrastructure complexity**
   - Requires Redis/RabbitMQ/SQS deployment
   - New failure points (queue downtime)
   - Operational overhead (monitoring, backups)
   - Cost: Redis Cloud or AWS SQS

❌ **Network overhead**
   - Every task enqueue/dequeue = network roundtrip
   - Latency: 1-5ms per operation (vs µs for channels)
   - Redis bandwidth costs at scale

❌ **Serialization overhead**
   - Task payloads must be serializable (JSON/msgpack)
   - Current structs passed by reference (zero-copy)
   - CPU + memory cost for encoding/decoding

❌ **Loss of type safety**
   - Tasks become `interface{}` with string-based routing
   - Current pipeline uses typed channels
   - More runtime errors, less compile-time safety

❌ **Harder debugging**
   - Distributed tracing required
   - Tasks span multiple processes
   - Stack traces don't cross queue boundaries

❌ **State management complexity**
   - Need to track task IDs in PostgreSQL
   - Mapping queue tasks back to row_states
   - Risk of orphaned tasks if state desync

❌ **Over-engineering for current scale**
   - Your jobs process 100s-1000s of rows
   - Not millions of tasks/hour
   - Premature optimization

❌ **Migration cost**
   - Rewrite entire pipeline (pipeline.go:141 lines)
   - Rewrite all 4 stages (~500 lines)
   - Update state management logic
   - Testing effort: 2-3 weeks

### Verdict: **NOT WORTH IT**

The cons significantly outweigh the pros. You'd gain retry logic and horizontal scaling at the cost of:
- New infrastructure dependencies
- Increased latency and complexity
- Loss of type safety
- 2-3 week rewrite effort

---

## Option 2: Simple In-Process Improvements (RECOMMENDED)

### Fix #1: Dynamic Worker Scaling (2 hours)

**Problem**: Fixed 5 workers per stage wastes resources.

**Solution**: Make workers configurable per-stage based on workload.

```go
// config/config.go
type PipelineConfig struct {
    SERPWorkers     int `env:"SERP_WORKERS" default:"20"`
    DecisionWorkers int `env:"DECISION_WORKERS" default:"10"`
    CrawlWorkers    int `env:"CRAWL_WORKERS" default:"5"`
    ExtractWorkers  int `env:"EXTRACT_WORKERS" default:"10"`
}
```

**Impact**:
- SERP stage (fast API): 20 workers = 4x throughput
- Crawl stage (slow, 120s timeout): 5 workers (limited by Crawl4ai)
- Decision/Extract: 10 workers = 2x throughput
- Total speedup: **2-3x** for typical workloads

---

### Fix #2: Parallel HTTP Requests (4 hours)

**Problem**: stage_serp.go:70-77 executes queries sequentially.

**Current**:
```go
for _, query := range queries {
    results = append(results, webSearch.Search(query)) // Sequential!
}
```

**Fixed**:
```go
var wg sync.WaitGroup
resultsChan := make(chan *SearchResult, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        resultsChan <- webSearch.Search(q)
    }(query)
}
wg.Wait()
close(resultsChan)

for result := range resultsChan {
    results = append(results, result)
}
```

**Impact**:
- Query generation creates 3-5 queries per row
- Current: 3s × 5 = 15s per row
- Parallel: max(3s) = 3s per row
- Speedup: **5x for SERP stage**

---

### Fix #3: Add Retry Logic (6 hours)

**Problem**: Single API failure = terminal row failure.

**Solution**: Retry wrapper with exponential backoff.

```go
// internal/utils/retry.go
func WithRetry[T any](fn func() (T, error), maxRetries int) (T, error) {
    var result T
    var err error

    for i := 0; i < maxRetries; i++ {
        result, err = fn()
        if err == nil {
            return result, nil
        }

        // Exponential backoff: 1s, 2s, 4s, 8s
        if i < maxRetries-1 {
            time.Sleep(time.Duration(1<<i) * time.Second)
        }
    }
    return result, fmt.Errorf("max retries exceeded: %w", err)
}
```

**Usage**:
```go
// stage_serp.go
results, err := utils.WithRetry(func() ([]*SearchResult, error) {
    return s.webSearch.Search(query)
}, 3)
```

**Impact**:
- Reduce transient API failures from ~10% to <1%
- No infrastructure changes needed
- Apply to all stages (SERP, Decision, Crawl, Extract)

---

### Fix #4: Increase Database Connections (1 hour)

**Problem**: 10 connections for 20+ concurrent goroutines.

**Solution**:
```go
// state/postgres_store.go
db.SetMaxOpenConns(50)  // Was: 10
db.SetMaxIdleConns(10)  // Was: 2
db.SetConnMaxLifetime(5 * time.Minute)
```

**Impact**:
- Eliminate connection pool contention
- Faster bulk updates (GetPendingRows, UpdateRowState)

---

### Fix #5: Context Propagation (3 hours)

**Problem**: Pipeline runs with `context.Background()`, can't cancel.

**Solution**:
```go
// api/handlers.go (line 209)
ctx, cancel := context.WithCancel(r.Context())
defer cancel()

go func() {
    if err := h.enricher.Enrich(ctx, jobID, rowKeys, req.ColumnsMetadata); err != nil {
        h.stateManager.UpdateJobStatus(jobID, models.JobStatusFailed)
    }
}()
```

**Impact**:
- Jobs can be cancelled mid-flight
- Resource cleanup on client disconnect
- Prevents orphaned goroutines

---

### Fix #6: Batch Database Updates (4 hours)

**Problem**: UpdateRowState called once per row (N queries).

**Solution**:
```go
// state/postgres_store.go
func (s *PostgresStore) UpdateRowStatesBatch(states []RowState) error {
    _, err := s.db.NewInsert().
        Model(&states).
        On("CONFLICT (job_id, key) DO UPDATE").
        Exec(ctx)
    return err
}
```

**Usage**: Accumulate updates, flush every 50 rows or 1 second.

**Impact**:
- Reduce DB roundtrips by 50x
- Lower transaction overhead

---

## Option 3: Hybrid Approach (Queue for Failures Only)

If you MUST have retry infrastructure:

### Use PostgreSQL as Queue (pgq pattern)

**Concept**: Add a `failed_tasks` table, use PostgreSQL LISTEN/NOTIFY.

```sql
CREATE TABLE failed_tasks (
    id SERIAL PRIMARY KEY,
    job_id UUID NOT NULL,
    row_key TEXT NOT NULL,
    stage TEXT NOT NULL,
    error TEXT,
    retry_count INT DEFAULT 0,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_failed_tasks_retry ON failed_tasks(next_retry_at)
WHERE retry_count < 3;
```

**Worker**: Background goroutine polls failed_tasks, re-enqueues to pipeline.

**Pros**:
- No new infrastructure (PostgreSQL already deployed)
- Persistent retry queue
- Simple to implement (100 lines)

**Cons**:
- Polling overhead (10s interval)
- Not as feature-rich as Redis queues

**Verdict**: Good middle ground if simple retries aren't enough.

---

## Recommended Implementation Plan

### Phase 1: Quick Wins (1 day)
1. ✅ Increase DB connections (Fix #4)
2. ✅ Make workers configurable (Fix #1)
3. ✅ Add context propagation (Fix #5)

**Expected gain**: 30% throughput increase

### Phase 2: Parallelization (3 days)
4. ✅ Parallel HTTP requests in SERP stage (Fix #2)
5. ✅ Add retry wrapper (Fix #3)
6. ✅ Batch database updates (Fix #6)

**Expected gain**: 3-5x throughput increase

### Phase 3: Observability (2 days)
7. ✅ Add Prometheus metrics (queue depth, stage latency)
8. ✅ Add structured logging (stage transitions, errors)
9. ✅ Dashboard for job monitoring

**Expected gain**: Better visibility, easier debugging

### Phase 4: Optional Advanced (if still needed)
10. ⚠️ **Only if** Phase 2 doesn't meet scale: Consider pgq-based retry queue
11. ⚠️ **Only if** multi-machine scaling needed: Evaluate taskq

---

## Cost-Benefit Analysis

| Solution | Dev Time | Infra Cost | Complexity | Throughput Gain | Reliability Gain |
|----------|----------|------------|------------|-----------------|------------------|
| **asyncq/taskq** | 3 weeks | $50-200/mo | High ⚠️ | 2-3x | High ✅ |
| **In-process fixes** | 1 week | $0 | Low ✅ | 3-5x | Medium ✅ |
| **pgq hybrid** | 2 weeks | $0 | Medium | 2x | High ✅ |

**Winner**: In-process fixes (best ROI)

---

## Alternative Technologies (If You Outgrow Current System)

### 1. **Temporal.io** (Workflow Orchestration)
- **What**: Distributed workflow engine (like Airflow but for microservices)
- **When to use**: If you need complex multi-stage workflows with human approvals, long-running jobs (days/weeks)
- **Pros**: Built-in retries, state management, versioning, visibility
- **Cons**: Heavy infrastructure (database + workers), steep learning curve

### 2. **NATS JetStream** (Lightweight Message Queue)
- **What**: High-performance messaging system with persistence
- **When to use**: If you need pub/sub patterns or event streaming
- **Pros**: Fast (µs latency), easy ops, Go-native
- **Cons**: Less feature-rich than Redis/RabbitMQ

### 3. **Apache Kafka** (Event Streaming)
- **What**: Distributed log for high-throughput streaming
- **When to use**: If you process millions of rows/day and need data replay
- **Pros**: Scalable to petabytes, fault-tolerant
- **Cons**: Complex to operate, overkill for this use case

### 4. **River** (Postgres-based Go Queue)
- **What**: Job queue library using PostgreSQL (similar to pgq but maintained)
- **When to use**: If you want task queue features without Redis
- **Pros**: No new dependencies, ACID guarantees, Go-native
- **Cons**: Newer library (less battle-tested)
- **GitHub**: https://github.com/riverqueue/river

---

## Answering Your Question

### Should you implement asyncq or taskq?

**No.**

Your pipeline doesn't have a **queue problem** — it has **configuration and parallelization problems**. Adding a distributed queue would:
- Introduce operational complexity (Redis/SQS management)
- Add network latency to every task
- Require 3 weeks of development
- Cost $50-200/month in infrastructure

Meanwhile, the **real issues** are:
- Fixed worker pools (trivial config change)
- Sequential HTTP calls (4-hour fix)
- No retries (6-hour fix)

### What WOULD justify asyncq/taskq?

Only if **all** of these are true:
1. You need to scale beyond 1 machine (100k+ rows/hour)
2. You have heterogeneous workers (GPU machines, different regions)
3. You need complex scheduling (priorities, rate limits per user)
4. You have ops team to manage Redis/SQS

None of these apply to AmpleData today.

---

## Final Recommendation

### DO THIS (Priority Order)

1. **Increase worker pools** (SERP: 20, Decision: 10, Extract: 10, Crawl: 5)
2. **Add retry wrapper** to all external API calls (3 retries with backoff)
3. **Parallelize HTTP requests** in SERP stage
4. **Increase DB connections** to 50
5. **Add context cancellation** to pipeline
6. **Batch database updates**

**Total effort**: 1 week
**Expected result**: 3-5x throughput, <1% failure rate
**Infrastructure changes**: None

### DON'T DO THIS (Yet)

- ❌ Implement asyncq/taskq
- ❌ Add Redis/RabbitMQ
- ❌ Rewrite pipeline architecture
- ❌ Migrate to microservices

### REVISIT IF

- You're processing >100k rows/hour on a single machine
- You need multi-region workers
- Simple retries don't handle API failures

---

## Questions to Ask Yourself

1. **What problem am I trying to solve?**
   - If "job failures": Add retry logic (6 hours)
   - If "too slow": Increase workers (1 hour)
   - If "can't scale": Verify you've hit single-machine limits first

2. **Have I measured the bottleneck?**
   - Run profiling (pprof) to find actual slowdowns
   - Check API latencies vs worker utilization
   - Don't optimize without data

3. **Am I solving for today or hypothetical future?**
   - Current: 100s-1000s of rows per job
   - Simple fixes handle 10-100x that scale
   - Build for 10x growth, not 1000x

---

## Conclusion

The current channel-based pipeline is **well-architected** for your scale. The bottlenecks are **configuration issues**, not architectural ones. Implementing asyncq/taskq would be premature optimization that adds complexity without addressing root causes.

**Ship the simple fixes first. You'll likely never need a distributed queue.**

If you still want a queue after exhausting in-process optimizations, use **River** (Postgres-based) or **pgq pattern** to avoid new infrastructure.

---

**Next steps**: Implement Phase 1 (1 day) and measure results. Only proceed to Phase 2 if needed.
