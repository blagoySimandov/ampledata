import type { SourceJobSummary } from "@/api";
import { useApi, useSourceData, useAllJobsRows } from "@/hooks";
import { useMemo } from "react";
import type { RowData, MergedDataResult } from "../types";

function seedEnrichedColsFromMetadata(
  job: SourceJobSummary,
  enrichedCols: Set<string>,
) {
  for (const col of job.columns_metadata ?? []) {
    enrichedCols.add(col.name);
  }
}

function markRowPending(job: SourceJobSummary, row: RowData) {
  if (!job.columns_metadata) return;
  row.__stages ??= {};
  for (const col of job.columns_metadata) {
    row.__stages[col.name] ??= "PENDING";
  }
}

function markAllRowsPending(
  job: SourceJobSummary,
  rowMap: Map<string, RowData>,
  csvRows: unknown[],
) {
  csvRows.forEach((_, rowIndex) => {
    const existing = rowMap.get(rowIndex.toString());
    if (existing) markRowPending(job, existing);
  });
}

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
    if (!sourceData) {
      return { rows: [], sourceColumns: [], enrichedColumns: [] };
    }

    const rowMap = new Map<string, RowData>();
    const enrichedCols = new Set<string>();

    // Build base rows from source CSV
    sourceData.rows.forEach((csvRow, index) => {
      const rowObj: RowData = { __index: index.toString() };
      sourceData.headers.forEach((header, i) => {
        rowObj[header] = csvRow[i];
      });
      rowMap.set(index.toString(), rowObj);
    });

    // Process from oldest to newest so newer runs overwrite
    const sortedQueries = [...jobQueries].reverse();
    const sortedJobs = [...jobs].reverse();

    sortedQueries.forEach((query, qIndex) => {
      const job = sortedJobs[qIndex];
      const keyCols = job.key_columns ?? [];
      if (keyCols.length === 0) return;

      seedEnrichedColsFromMetadata(job, enrichedCols);

      if (!query.data) {
        markAllRowsPending(job, rowMap, sourceData.rows);
        return;
      }

      const keyIndices = keyCols.map((kc) => sourceData.headers.indexOf(kc));

      const jobRowsByKey = new Map<string, (typeof query.data.rows)[number]>();
      for (const row of query.data.rows) {
        jobRowsByKey.set(row.key, row);
      }

      sourceData.rows.forEach((csvRow, rowIndex) => {
        const jobKey = keyIndices
          .map((idx) => (idx !== -1 && idx < csvRow.length ? csvRow[idx] : ""))
          .join("||");

        const existing = rowMap.get(rowIndex.toString());
        if (!existing) return;

        const jobRow = jobRowsByKey.get(jobKey);
        if (!jobRow) {
          markRowPending(job, existing);
          return;
        }

        if (jobRow.extracted_data) {
          Object.assign(existing, jobRow.extracted_data);
          for (const key of Object.keys(jobRow.extracted_data)) {
            enrichedCols.add(key);
          }
        }

        if (jobRow.confidence) {
          existing.__confidence ??= {};
          Object.assign(existing.__confidence, jobRow.confidence);
          for (const key of Object.keys(jobRow.confidence)) {
            enrichedCols.add(key);
          }
        }

        if (jobRow.sources) {
          existing.__sources ??= {};
          if (job.columns_metadata) {
            for (const col of job.columns_metadata) {
              existing.__sources[col.name] = jobRow.sources;
            }
          } else if (jobRow.extracted_data) {
            for (const key of Object.keys(jobRow.extracted_data)) {
              existing.__sources[key] = jobRow.sources;
            }
          }
        }

        if (job.columns_metadata) {
          existing.__stages ??= {};
          for (const col of job.columns_metadata) {
            existing.__stages[col.name] = jobRow.stage;
          }
        }
      });
    });

    return {
      rows: Array.from(rowMap.values()),
      sourceColumns: sourceData.headers,
      enrichedColumns: Array.from(enrichedCols).sort(),
    };
  }, [sourceData, jobQueries, jobs]);

  return { data: mergedData, isFetching };
}
