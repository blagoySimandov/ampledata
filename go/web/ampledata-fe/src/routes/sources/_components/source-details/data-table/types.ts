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

export interface HistoryEntryForField {
  attempt_number: number;
  value: unknown;
  confidence: ConfidenceEntry | undefined;
  sources: string[] | undefined;
  reasoning: string | undefined;
}

export interface RowData {
  __index: string;
  __confidence?: Record<string, ConfidenceEntry>;
  __stages?: Record<string, string>;
  __sources?: Record<string, string[]>;
  __extractionHistory?: Record<string, HistoryEntryForField[]>;
  [field: string]: unknown;
}

export interface MergedDataResult {
  rows: RowData[];
  sourceColumns: string[];
  enrichedColumns: string[];
}
