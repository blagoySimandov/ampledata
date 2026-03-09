// src/api/decoder.ts

import type {
  JobListResponse,
  JobProgressResponse,
  RowsProgressResponse,
  EnrichmentResult,
} from './types';

// Simple decoders for now that just cast the data.
// This can be expanded with Zod or a custom validation logic later.

export function decodeJobList(data: unknown): JobListResponse {
  return data as JobListResponse;
}

export function decodeJobProgress(data: unknown): JobProgressResponse {
  return data as JobProgressResponse;
}

export function decodeRowsProgress(data: unknown): RowsProgressResponse {
  return data as RowsProgressResponse;
}

export function decodeEnrichmentResults(data: unknown): EnrichmentResult[] {
  return data as EnrichmentResult[];
}
