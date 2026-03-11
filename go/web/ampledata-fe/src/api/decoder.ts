// src/api/decoder.ts

import type {
  SourceListResponse,
  SourceDetail,
  JobProgressResponse,
  RowsProgressResponse,
} from "./types";

export function decodeSourceList(data: unknown): SourceListResponse {
  return data as SourceListResponse;
}

export function decodeSourceDetail(data: unknown): SourceDetail {
  return data as SourceDetail;
}

export function decodeJobProgress(data: unknown): JobProgressResponse {
  return data as JobProgressResponse;
}

export function decodeRowsProgress(data: unknown): RowsProgressResponse {
  return data as RowsProgressResponse;
}
