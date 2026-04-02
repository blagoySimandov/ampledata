import type {
  RowProgressItem,
  RowsProgressResponse,
  SourceDataResponse,
  SourceJobSummary,
} from "@/api";
import { useApi, useSourceData, useAllJobsRows } from "@/hooks";
import { useMemo } from "react";
import type { RowData, MergedDataResult, HistoryEntryForField } from "../types";
import type { UseQueryResult } from "@tanstack/react-query";

// ─── Row building ────────────────────────────────────────────────────────────

function buildRowFromCsv(
  csvRow: unknown[],
  headers: string[],
  index: number,
): RowData {
  const row: RowData = { __index: index.toString() };
  headers.forEach((header, i) => {
    row[header] = csvRow[i];
  });
  return row;
}

function buildBaseRowMap(sourceData: SourceDataResponse): Map<string, RowData> {
  const rowMap = new Map<string, RowData>();
  sourceData.rows.forEach((csvRow, index) => {
    rowMap.set(
      index.toString(),
      buildRowFromCsv(csvRow, sourceData.headers, index),
    );
  });
  return rowMap;
}

// ─── Key resolution ──────────────────────────────────────────────────────────

function resolveKeyIndices(keyCols: string[], headers: string[]): number[] {
  return keyCols.map((kc) => headers.indexOf(kc));
}

function buildCompositeKey(csvRow: unknown[], keyIndices: number[]): string {
  return keyIndices
    .map((idx) =>
      idx !== -1 && idx < (csvRow as unknown[]).length ? csvRow[idx] : "",
    )
    .join("||");
}

function indexJobRowsByKey(
  jobRows: RowProgressItem[],
): Map<string, RowProgressItem> {
  const map = new Map<string, RowProgressItem>();
  for (const row of jobRows) map.set(row.key, row);
  return map;
}

// ─── Pending state ───────────────────────────────────────────────────────────

function markJobColumnsAsPending(row: RowData, job: SourceJobSummary) {
  if (!job.columns_metadata) return;
  row.__stages ??= {};
  for (const col of job.columns_metadata) {
    row.__stages[col.name] ??= "PENDING";
  }
}

function markAllRowsAsPendingForJob(
  job: SourceJobSummary,
  rowMap: Map<string, RowData>,
  totalRows: number,
) {
  for (let i = 0; i < totalRows; i++) {
    const row = rowMap.get(i.toString());
    if (row) markJobColumnsAsPending(row, job);
  }
}

// ─── Enriching rows ──────────────────────────────────────────────────────────

function applyExtractedData(
  row: RowData,
  jobRow: RowProgressItem,
  enrichedCols: Set<string>,
) {
  if (!jobRow.extracted_data) return;
  Object.assign(row, jobRow.extracted_data);
  for (const key of Object.keys(jobRow.extracted_data)) enrichedCols.add(key);
}

function applyConfidence(
  row: RowData,
  jobRow: RowProgressItem,
  enrichedCols: Set<string>,
) {
  if (!jobRow.confidence) return;
  row.__confidence ??= {};
  Object.assign(row.__confidence, jobRow.confidence);
  for (const key of Object.keys(jobRow.confidence)) enrichedCols.add(key);
}

function applySourceLinks(
  row: RowData,
  jobRow: RowProgressItem,
  job: SourceJobSummary,
) {
  if (!jobRow.sources) return;
  row.__sources ??= {};

  const targetKeys = job.columns_metadata
    ? job.columns_metadata.map((col) => col.name)
    : Object.keys(jobRow.extracted_data ?? {});

  for (const key of targetKeys) row.__sources[key] = jobRow.sources;
}

function applyStages(
  row: RowData,
  jobRow: RowProgressItem,
  job: SourceJobSummary,
) {
  if (!job.columns_metadata) return;
  row.__stages ??= {};
  for (const col of job.columns_metadata) row.__stages[col.name] = jobRow.stage;
}

function applyExtractionHistory(row: RowData, jobRow: RowProgressItem) {
  if (!jobRow.extraction_history || jobRow.extraction_history.length === 0)
    return;
  row.__extractionHistory ??= {};
  for (const entry of jobRow.extraction_history) {
    const fields = new Set([
      ...Object.keys(entry.extracted_data ?? {}),
      ...Object.keys(entry.confidence ?? {}),
    ]);
    for (const field of fields) {
      row.__extractionHistory[field] ??= [];
      row.__extractionHistory[field].push({
        attempt_number: entry.attempt_number,
        value: entry.extracted_data?.[field],
        confidence: entry.confidence?.[field],
        sources: entry.sources ?? undefined,
        reasoning: entry.reasoning ?? undefined,
      } satisfies HistoryEntryForField);
    }
  }
}

function enrichRowWithJobResult(
  row: RowData,
  jobRow: RowProgressItem,
  job: SourceJobSummary,
  enrichedCols: Set<string>,
) {
  applyExtractedData(row, jobRow, enrichedCols);
  applyConfidence(row, jobRow, enrichedCols);
  applySourceLinks(row, jobRow, job);
  applyStages(row, jobRow, job);
  applyExtractionHistory(row, jobRow);
}

// ─── Job application ─────────────────────────────────────────────────────────

function collectEnrichedColumnsFromJobMetadata(
  job: SourceJobSummary,
  enrichedCols: Set<string>,
) {
  for (const col of job.columns_metadata ?? []) enrichedCols.add(col.name);
}

function applyJobResultsToRows(
  job: SourceJobSummary,
  jobRows: RowProgressItem[],
  sourceData: SourceDataResponse,
  rowMap: Map<string, RowData>,
  enrichedCols: Set<string>,
) {
  const keyCols = job.key_columns ?? [];
  if (keyCols.length === 0) return;

  collectEnrichedColumnsFromJobMetadata(job, enrichedCols);

  const keyIndices = resolveKeyIndices(keyCols, sourceData.headers);
  const jobRowsByKey = indexJobRowsByKey(jobRows);

  sourceData.rows.forEach((csvRow, rowIndex) => {
    const row = rowMap.get(rowIndex.toString());
    if (!row) return;

    const key = buildCompositeKey(csvRow, keyIndices);
    const jobRow = jobRowsByKey.get(key);

    if (!jobRow) {
      // Row was not included in this job (e.g. excluded by max_rows) — leave it untouched.
      return;
    }

    enrichRowWithJobResult(row, jobRow, job, enrichedCols);
  });
}

function applyJobToRows(
  job: SourceJobSummary,
  query: UseQueryResult<RowsProgressResponse>,
  sourceData: SourceDataResponse,
  rowMap: Map<string, RowData>,
  enrichedCols: Set<string>,
) {
  const keyCols = job.key_columns ?? [];
  if (keyCols.length === 0) return;

  collectEnrichedColumnsFromJobMetadata(job, enrichedCols);

  if (!query.data) {
    // While loading, mark only the rows that will actually be processed.
    // job.total_rows reflects the capped count (respecting max_rows).
    markAllRowsAsPendingForJob(job, rowMap, job.total_rows);
    return;
  }

  applyJobResultsToRows(job, query.data.rows, sourceData, rowMap, enrichedCols);
}

/**
 * Merges CSV source data with enrichment job results into a unified row set.
 *
 * Fetches the raw source rows for `sourceId` and overlays enriched columns
 * from each job in `jobs`. Newer jobs take precedence over older ones for
 * overlapping columns. Rows with no matching job result are marked as PENDING.
 *
 * @param sourceId - The ID of the source CSV to fetch
 * @param jobs     - Ordered list of enrichment jobs (newest first)
 * @returns        - Merged rows, original source columns, enriched column names,
 *                   and a combined fetching flag
 *
 * @example
 * const { data, isFetching } = useMergedData("src_123", jobs);
 * // data.rows        → all rows with enriched fields merged in
 * // data.enrichedColumns → ["email", "company", ...]
 */
export function useMergedData(
  sourceId: string,
  jobs: SourceJobSummary[],
): { data: MergedDataResult; isFetching: boolean } {
  const api = useApi();
  const { data: sourceData, isFetching: sourceFetching } = useSourceData(
    api,
    sourceId,
  );
  const jobQueries = useAllJobsRows(api, jobs);

  const isFetching = sourceFetching || jobQueries.some((q) => q.isFetching);

  const mergedData = useMemo<MergedDataResult>(() => {
    if (!sourceData)
      return { rows: [], sourceColumns: [], enrichedColumns: [] };

    const rowMap = buildBaseRowMap(sourceData);
    const enrichedCols = new Set<string>();

    // Reverse so newest jobs overwrite oldest for the same columns
    const jobsOldestFirst = [...jobs].reverse();
    const queriesOldestFirst = [...jobQueries].reverse();

    queriesOldestFirst.forEach((query, i) => {
      applyJobToRows(
        jobsOldestFirst[i],
        query,
        sourceData,
        rowMap,
        enrichedCols,
      );
    });

    return {
      rows: Array.from(rowMap.values()),
      sourceColumns: sourceData.headers,
      enrichedColumns: Array.from(enrichedCols).sort(),
    };
  }, [sourceData, jobQueries, jobs]);

  return { data: mergedData, isFetching };
}
