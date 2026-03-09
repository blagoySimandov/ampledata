// src/api/types.ts

export type JobStatus = 'PENDING' | 'RUNNING' | 'PAUSED' | 'CANCELLED' | 'COMPLETED';

export type JobType = 'enrichment' | 'imputation';

export type ColumnType = 'string' | 'number' | 'boolean' | 'date';

export type RowStage = 'PENDING' | 'SERP_FETCHED' | 'DECISION_MADE' | 'CRAWLED' | 'ENRICHED' | 'COMPLETED' | 'FAILED' | 'CANCELLED';

export interface ColumnMetadata {
  name: string;
  type: ColumnType;
  job_type: JobType;
  description?: string | null;
}

export interface JobSummary {
  job_id: string;
  status: JobStatus;
  total_rows: number;
  file_path: string;
  created_at: string;
  started_at?: string | null;
}

export interface JobListResponse {
  jobs: JobSummary[];
  total_count: number;
}

export interface JobProgressResponse {
  job_id: string;
  total_rows: number;
  rows_by_stage: Record<string, number>;
  started_at: string;
  status: JobStatus;
}

export interface FieldConfidenceInfo {
  score: number;
  reason: string;
}

export interface RowProgressItem {
  key: string;
  stage: RowStage;
  updated_at: string;
  confidence?: Record<string, FieldConfidenceInfo> | null;
  error?: string | null;
  extracted_data?: Record<string, unknown> | null;
  sources?: string[] | null;
}

export interface PaginationInfo {
  total: number;
  offset: number;
  limit: number;
  has_more: boolean;
}

export interface RowsProgressResponse {
  rows: RowProgressItem[];
  pagination: PaginationInfo;
}

export interface EnrichmentResult {
  key: string;
  extracted_data: Record<string, unknown>;
  sources: string[];
  confidence?: Record<string, FieldConfidenceInfo> | null;
  error?: string | null;
}

export interface SignedURLRequest {
  contentType: 'text/csv' | 'application/json';
  length: number;
}

export interface SignedURLResponse {
  url: string;
  jobId: string;
}

export interface SelectKeyRequest {
  job_id: string;
  columns_metadata?: ColumnMetadata[] | null;
}

export interface SelectKeyResponse {
  selected_key: string;
  all_keys: string[];
  reasoning: string;
}

export interface StartJobRequest {
  key_columns: string[];
  columns_metadata: ColumnMetadata[];
  key_column_description?: string | null;
}

export interface StartJobResponse {
  job_id: string;
  message: string;
}
