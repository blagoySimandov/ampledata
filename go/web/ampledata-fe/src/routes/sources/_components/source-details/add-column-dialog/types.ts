import type { SourceJobSummary, ColumnMetadata, Template } from "@/api";

export interface AddColumnsDialogProps {
  sourceId: string;
  mostRecentJob?: SourceJobSummary;
  initialTemplate?: Template;
}

export interface ColumnEditorProps {
  column: ColumnMetadata;
  sourceColumns: string[];
  onUpdate: (updates: Partial<ColumnMetadata>) => void;
  onRemove: () => void;
}
