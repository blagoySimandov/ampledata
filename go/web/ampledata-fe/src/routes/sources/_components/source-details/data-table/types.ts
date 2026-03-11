export interface ConfidenceEntry {
  score: number;
  reason: string;
}

export interface ConfidenceConfig {
  label: string;
  color: string;
  bg: string;
  borderColor: string;
}

export interface RowData {
  __index: string;
  __confidence?: Record<string, ConfidenceEntry>;
  __stages?: Record<string, string>;
  __sources?: Record<string, string[]>;
  [field: string]: unknown;
}

export interface MergedDataResult {
  rows: RowData[];
  sourceColumns: string[];
  enrichedColumns: string[];
}
