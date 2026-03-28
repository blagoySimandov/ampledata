import type { SourceJobSummary, ColumnMetadata } from "@/api";

export interface AddColumnsDialogProps {
  sourceId: string;
  mostRecentJob?: SourceJobSummary;
}

export interface ColumnEditorProps {
  column: ColumnMetadata;
  sourceColumns: string[];
  onUpdate: (updates: Partial<ColumnMetadata>) => void;
  onRemove: () => void;
}
