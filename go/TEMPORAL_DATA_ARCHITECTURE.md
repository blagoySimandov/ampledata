# Temporal Data Architecture

## TL;DR

**✅ YES - This is the CORRECT way to use Temporal!**

- **PostgreSQL**: Stores only final business results (`extracted_data`, `confidence`)
- **Temporal**: Automatically stores ALL intermediate data in its execution history
- **Debugging**: Use Temporal UI (localhost:8233) to inspect intermediate data

---

## The Problem You Noticed

You found that these columns in PostgreSQL were empty:
- `serp_data` ❌
- `decision` ❌
- `crawl_results` ❌

Only these were filled:
- `extracted_data` ✅
- `confidence` ❌ (was a bug, now fixed)

## Why This is Actually Correct

### Temporal is NOT a Database

Temporal is a **workflow orchestration system** that happens to keep a complete execution history. It's designed for:
- ✅ Workflow execution and retry logic
- ✅ Auditing and debugging (via UI)
- ✅ Failure recovery and replay
- ❌ NOT for querying business data
- ❌ NOT for your application's data layer

### PostgreSQL is Your Business Database

PostgreSQL should only store data that your **application needs to query**:
- ✅ Final enrichment results
- ✅ Confidence scores for quality assessment
- ✅ Sources (URLs where data was found) for transparency
- ✅ Job status and progress tracking
- ❌ NOT intermediate pipeline artifacts (raw HTML, full search results)

---

## Where is My Data?

### Data in PostgreSQL

```sql
SELECT key, stage, extracted_data, confidence, sources, error
FROM row_states
WHERE job_id = 'your-job-id';
```

**Contains:**
- `extracted_data`: Final enriched fields (e.g., `{"email": "john@example.com", "phone": "+1-555-0123"}`)
- `confidence`: Confidence scores per field (e.g., `{"email": {"score": 0.95, "reason": "Found in contact section"}}`)
- `sources`: URLs where data was found (e.g., `["https://linkedin.com/in/john", "https://example.com/about"]`)
- `stage`: Current enrichment stage
- `error`: Error message if failed

### Data in Temporal

**How to access:**
1. Open Temporal UI: http://localhost:8233
2. Click on a workflow execution
3. View the complete history including:
   - `serp_data`: All search results from Google/Bing
   - `decision`: Which URLs the decision maker chose to crawl
   - `crawl_results`: Raw HTML/text content from websites
   - All activity inputs/outputs
   - Exact timestamps and retry attempts

**Example workflow inspection:**
```
Workflow: EnrichmentWorkflow
├─ Activity: SerpFetch ✅
│  ├─ Input: {"rowKey": "john-doe", "columns": [...]}
│  └─ Output: {"serpData": {"queries": ["john doe email"], "results": [...]}}
├─ Activity: MakeDecision ✅
│  ├─ Input: {"serpData": {...}, "columns": [...]}
│  └─ Output: {"decision": {"urlsToCrawl": ["linkedin.com/john"], "reasoning": "..."}}
├─ Activity: Crawl ✅
│  ├─ Input: {"urls": ["linkedin.com/john"]}
│  └─ Output: {"crawlResults": {"content": "<html>...</html>", "sources": [...]}}
└─ Activity: Extract ✅
   ├─ Input: {"crawlResults": {...}, "columns": [...]}
   └─ Output: {"extractedData": {"email": "..."}, "confidence": {...}}
```

---

## Benefits of This Architecture

### 1. **Performance**
- PostgreSQL queries are fast (no huge JSONB fields to deserialize)
- Only stores data you actually need to retrieve

### 2. **Cost**
- Smaller database size
- Less storage costs
- Faster backups

### 3. **Clarity**
- Clear separation: Temporal = orchestration, PostgreSQL = business data
- API responses only return relevant data

### 4. **Debugging**
- Temporal UI provides better visualization than database queries
- Can see exact execution flow, timings, and retry attempts
- Built-in filtering and search

### 5. **Maintainability**
- Simpler database schema
- Less code to maintain (36 lines removed!)
- Fewer bugs (no sync issues between Temporal and PostgreSQL)

---

## Common Questions

### "Can I query Temporal for this data?"

**Technically yes, but you shouldn't:**
```go
// DON'T DO THIS - Temporal is not a query database
rows := temporalClient.DescribeWorkflowExecution(...)
// Extract data from workflow history...
```

**Temporal's query APIs are designed for:**
- Workflow status checks
- Debugging specific executions
- Admin/monitoring tools

**They are NOT designed for:**
- Bulk data retrieval
- Application business logic
- Serving API responses

### "What if I need intermediate data later?"

**For debugging:** Use Temporal UI - it's designed for this!

**For application logic:** If you find yourself needing intermediate data in your API responses, you should reconsider whether it's truly "intermediate" or actually part of your business data model.

**For feedback loops:** Pass the data as workflow parameters, not by querying storage. Example:
```go
// Bad: Query for old data
serpData := getFromDatabase(jobID)
newWorkflow.Execute(..., serpData)

// Good: Temporal already has it
workflow.ContinueAsNew(input) // Carries data forward
```

### "What about data retention?"

**Temporal:**
- Default retention: 30 days (configurable)
- After retention period, workflow history is deleted
- Fine for debugging - you don't need 2-year-old search results

**PostgreSQL:**
- Keeps final results forever (or per your retention policy)
- This is your source of truth for business data

### "How do I inspect data for failed workflows?"

1. Go to Temporal UI: http://localhost:8233
2. Filter by status: "Failed"
3. Click on the failed workflow
4. View the complete execution history
5. See exactly which activity failed and why
6. Inspect the inputs/outputs of each step

This is MUCH better than digging through database dumps!

---

## Migration Notes

### Existing Databases

If you have existing data with populated `serp_data`/`decision`/`crawl_results` columns:

**Option 1: Leave them (safest)**
- Old rows keep their data
- New rows will have these columns empty
- No migration needed

**Option 2: Drop columns (cleaner)**
```sql
-- After verifying everything works
ALTER TABLE row_states DROP COLUMN serp_data;
ALTER TABLE row_states DROP COLUMN decision;
ALTER TABLE row_states DROP COLUMN crawl_results;
```

### Testing Confidence Fix

The bug where `confidence` wasn't saved has been fixed. Test it:

```go
// Run an enrichment job
jobID := startEnrichmentJob(...)

// Wait for completion
waitForCompletion(jobID)

// Check confidence is now saved
rows, _ := db.Query("SELECT confidence FROM row_states WHERE job_id = $1", jobID)
// Should see: {"email": {"score": 0.95, "reason": "..."}, ...}
```

---

## Summary

| Data Type | Storage Location | Access Method | Use Case |
|-----------|-----------------|---------------|----------|
| Final results | PostgreSQL | SQL queries | API responses, reporting |
| Confidence scores | PostgreSQL | SQL queries | Quality assessment, filtering |
| Sources (URLs) | PostgreSQL | SQL queries | Transparency, verification |
| Search results | Temporal history | Temporal UI | Debugging, auditing |
| Crawl decisions | Temporal history | Temporal UI | Debugging, auditing |
| Raw HTML content | Temporal history | Temporal UI | Debugging, auditing |
| Workflow status | Both | SQL + Temporal API | Progress tracking |

**Key Principle:** PostgreSQL = what you need to QUERY, Temporal = everything else

This architecture gives you the best of both worlds:
- Fast, efficient business data queries
- Complete execution history for debugging
- Clear separation of concerns
- Industry-standard Temporal pattern

You're doing it right! 🎉
