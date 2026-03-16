export type JobStatus =
  | "PENDING"
  | "RUNNING"
  | "PAUSED"
  | "CANCELLED"
  | "COMPLETED";

export type JobType = "enrichment" | "imputation";

export type ColumnType = "string" | "number" | "boolean" | "date";

export type RowStage =
  | "PENDING"
  | "SERP_FETCHED"
  | "DECISION_MADE"
  | "CRAWLED"
  | "ENRICHED"
  | "COMPLETED"
  | "FAILED"
  | "CANCELLED";

export interface ColumnMetadata {
  name: string;
  type: ColumnType;
  job_type: JobType;
  description?: string | null;
}

export interface SourceJobSummary {
  job_id: string;
  status: JobStatus;
  total_rows: number;
  key_columns?: string[] | null;
  key_column_description?: string | null;
  columns_metadata?: ColumnMetadata[] | null;
  created_at: string;
  started_at?: string | null;
}

export interface SourceSummary {
  source_id: string;
  type: string;
  created_at: string;
  job_count: number;
  latest_job_status?: JobStatus | null;
}

export interface SourceDetail {
  source_id: string;
  type: string;
  created_at: string;
  jobs: SourceJobSummary[];
}

export interface SourceListResponse {
  sources: SourceSummary[];
  total_count: number;
}

export interface SourceDataResponse {
  headers: string[];
  rows: string[][];
}

export interface EnrichRequest {
  columns_metadata: ColumnMetadata[];
  key_columns?: string[] | null;
  key_column_description?: string | null;
}

export interface EnrichResponse {
  job_id: string;
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

export interface SignedURLRequest {
  contentType: "text/csv" | "application/json";
  length: number;
}

export interface SignedURLResponse {
  url: string;
  sourceId: string;
}

export interface SelectKeyRequest {
  source_id: string;
  columns_metadata?: ColumnMetadata[] | null;
}

export interface SelectKeyResponse {
  selected_key: string;
  all_keys: string[];
  reasoning: string;
}

export interface TierResponse {
  id: string;
  display_name: string;
  monthly_price_cents: number;
  included_tokens: number;
  overage_price_cents_decimal: string;
}

export interface SubscriptionStatusResponse {
  tier: string | null;
  tokens_included: number;
  tokens_used: number;
  cancel_at_period_end: boolean;
  current_period_start: string | null;
  current_period_end: string | null;
}

export interface PortalSessionResponse {
  url: string;
}

export interface CreateSubscriptionRequest {
  success_url: string;
  cancel_url: string;
  tier_id: string;
}

export interface CreateCheckoutResponse {
  checkout_url: string;
  session_id: string;
}

export interface UserResponse {
  first_name: string;
  last_name: string;
  email: string;
}
