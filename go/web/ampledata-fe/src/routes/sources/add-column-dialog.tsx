// ---------------------------------------------------------------------------
// AddColumnsDialog
// ---------------------------------------------------------------------------

import type { SourceJobSummary, ColumnMetadata } from "@/api";
import { Button } from "@/components/ui/button";
import {
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  Dialog,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
  Select,
} from "@/components/ui/select";
import { useApi, useEnrich, useSourceData } from "@/hooks";
import { Plus, Settings2, Loader2, Trash2 } from "lucide-react";
import { useState } from "react";
interface EnrichPayload {
  columns_metadata: ColumnMetadata[];
  key_columns?: string[];
  key_column_description?: string;
}

interface AddColumnsDialogProps {
  sourceId: string;
  mostRecentJob?: SourceJobSummary;
}

export function AddColumnsDialog({
  sourceId,
  mostRecentJob,
}: AddColumnsDialogProps) {
  const api = useApi();
  const enrich = useEnrich(api, sourceId);
  const { data: sourceData } = useSourceData(api, sourceId);
  const sourceColumns = sourceData?.headers ?? [];

  const [open, setOpen] = useState(false);
  const [columnsMetadata, setColumnsMetadata] = useState<ColumnMetadata[]>([]);
  const [selectedKeyColumns, setSelectedKeyColumns] = useState<string[]>([]);
  const [keyColumnDescription, setKeyColumnDescription] = useState("");

  const canStart =
    columnsMetadata.length > 0 && columnsMetadata.every((c) => c.name);

  const addColumn = () =>
    setColumnsMetadata((prev) => [
      ...prev,
      { name: "", type: "string", job_type: "enrichment" },
    ]);

  const removeColumn = (index: number) =>
    setColumnsMetadata((prev) => prev.filter((_, i) => i !== index));

  const updateColumn = (index: number, updates: Partial<ColumnMetadata>) =>
    setColumnsMetadata((prev) =>
      prev.map((col, i) => (i === index ? { ...col, ...updates } : col)),
    );

  const toggleKeyColumn = (col: string) =>
    setSelectedKeyColumns((prev) =>
      prev.includes(col) ? prev.filter((c) => c !== col) : [...prev, col],
    );

  const resetForm = () => {
    setColumnsMetadata([]);
    setSelectedKeyColumns([]);
    setKeyColumnDescription("");
  };

  const handleOpenChange = (val: boolean) => {
    setOpen(val);
    if (!val) {
      resetForm();
    } else if (mostRecentJob?.key_columns) {
      setSelectedKeyColumns(mostRecentJob.key_columns);
    }
  };

  const handleEnrich = async () => {
    try {
      const payload: EnrichPayload = { columns_metadata: columnsMetadata };

      if (selectedKeyColumns.length > 0) {
        payload.key_columns = selectedKeyColumns;
      }

      const trimmedDescription = keyColumnDescription.trim();
      if (trimmedDescription) {
        payload.key_column_description = trimmedDescription;
      }

      await enrich.mutateAsync(payload);
      setOpen(false);
      resetForm();
    } catch (e) {
      console.error("Enrich failed", e);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button className="font-bold gap-2">
          <Plus className="w-4 h-4" /> ADD COLUMNS
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[560px] max-h-[90vh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="p-6 pb-0">
          <DialogTitle className="text-2xl font-black">
            Add Enrichment Columns
          </DialogTitle>
        </DialogHeader>
        <ScrollArea className="flex-1 px-6">
          <div className="py-6 space-y-6">
            {/* Key Column Settings */}
            <div className="space-y-4">
              <Label className="text-xs font-black uppercase tracking-widest text-slate-400 flex items-center gap-2">
                <Settings2 className="w-3 h-3" /> Key Column Settings
              </Label>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label className="text-xs text-slate-500 font-bold">
                    Select Key Columns
                  </Label>
                  <p className="text-[10px] text-slate-400 leading-tight">
                    Select columns that uniquely identify the entity (e.g.
                    Company Name, Website) for the AI to search with.
                  </p>
                  <div className="flex flex-wrap gap-1.5 pt-1">
                    {sourceColumns.length > 0 ? (
                      sourceColumns.map((col) => {
                        const isSelected = selectedKeyColumns.includes(col);
                        return (
                          <Badge
                            key={col}
                            variant={isSelected ? "default" : "outline"}
                            className={`cursor-pointer transition-colors ${
                              isSelected
                                ? ""
                                : "text-slate-500 hover:bg-slate-100 hover:text-slate-900 border-slate-200"
                            }`}
                            onClick={() => toggleKeyColumn(col)}
                          >
                            {col}
                          </Badge>
                        );
                      })
                    ) : (
                      <div className="text-xs text-slate-400 italic">
                        Loading columns...
                      </div>
                    )}
                  </div>
                </div>
                <div className="space-y-1 pt-2">
                  <Label className="text-xs text-slate-500 font-bold">
                    Key Column Definition (AI Context)
                  </Label>
                  <Input
                    placeholder="Optional rules to locate the key columns in raw text..."
                    value={keyColumnDescription}
                    onChange={(e) => setKeyColumnDescription(e.target.value)}
                    className="text-sm h-9"
                  />
                </div>
              </div>
            </div>

            {/* New Fields */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <Label className="text-xs font-black uppercase tracking-widest text-slate-400 flex items-center gap-2">
                  <Settings2 className="w-3 h-3" /> New Fields
                </Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={addColumn}
                  className="h-7 text-xs font-black px-2 hover:bg-slate-100"
                >
                  <Plus className="w-3 h-3 mr-1" /> ADD FIELD
                </Button>
              </div>

              {columnsMetadata.length === 0 ? (
                <EmptyFieldsPlaceholder onAdd={addColumn} />
              ) : (
                <div className="space-y-3">
                  {columnsMetadata.map((col, index) => (
                    <ColumnEditor
                      key={index}
                      column={col}
                      onUpdate={(updates) => updateColumn(index, updates)}
                      onRemove={() => removeColumn(index)}
                    />
                  ))}
                </div>
              )}
            </div>

            <Button
              className="w-full font-black h-12"
              onClick={handleEnrich}
              disabled={!canStart || enrich.isPending}
            >
              {enrich.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                  STARTING...
                </>
              ) : (
                "START ENRICHMENT"
              )}
            </Button>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

function EmptyFieldsPlaceholder({ onAdd }: { onAdd: () => void }) {
  return (
    <div className="bg-slate-50 border border-slate-100 rounded-2xl p-6 flex flex-col items-center text-center gap-3">
      <div className="w-10 h-10 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100">
        <Settings2 className="w-5 h-5 text-slate-400" />
      </div>
      <div className="space-y-1">
        <p className="text-sm font-bold text-slate-900">Define new fields</p>
        <p className="text-xs text-slate-500 max-w-[280px] leading-relaxed">
          Add columns to enrich. Previous runs are not re-processed.
        </p>
      </div>
      <Button
        variant="outline"
        size="sm"
        onClick={onAdd}
        className="mt-1 font-bold h-8 px-4 border-slate-200"
      >
        <Plus className="w-3 h-3 mr-2" /> ADD YOUR FIRST FIELD
      </Button>
    </div>
  );
}

interface ColumnEditorProps {
  column: ColumnMetadata;
  onUpdate: (updates: Partial<ColumnMetadata>) => void;
  onRemove: () => void;
}

function ColumnEditor({ column, onUpdate, onRemove }: ColumnEditorProps) {
  return (
    <div className="flex flex-col gap-2 p-3 bg-white border border-slate-100 rounded-xl shadow-sm">
      <div className="flex items-center gap-2">
        <Input
          placeholder="Field name"
          value={column.name}
          onChange={(e) => onUpdate({ name: e.target.value })}
          className="h-9 text-xs font-medium"
        />
        <Select
          value={column.job_type}
          onValueChange={(v: "enrichment" | "imputation") =>
            onUpdate({ job_type: v })
          }
        >
          <SelectTrigger className="h-9 w-[120px] text-xs font-medium bg-slate-50 border-transparent">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="enrichment">Enrich</SelectItem>
            <SelectItem value="imputation">Impute</SelectItem>
          </SelectContent>
        </Select>
        <Select
          value={column.type}
          onValueChange={(v: "string" | "number" | "boolean" | "date") =>
            onUpdate({ type: v })
          }
        >
          <SelectTrigger className="h-9 w-[100px] text-xs font-medium bg-slate-50 border-transparent">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="string">String</SelectItem>
            <SelectItem value="number">Number</SelectItem>
            <SelectItem value="boolean">Bool</SelectItem>
            <SelectItem value="date">Date</SelectItem>
          </SelectContent>
        </Select>
        <Button
          variant="ghost"
          size="icon"
          onClick={onRemove}
          className="h-9 w-9 text-slate-400 hover:text-red-500 hover:bg-red-50 shrink-0"
        >
          <Trash2 className="w-4 h-4" />
        </Button>
      </div>
      <Input
        placeholder="Optional AI instructions"
        value={column.description ?? ""}
        onChange={(e) => onUpdate({ description: e.target.value })}
        className="h-8 text-xs bg-slate-50/50 border-slate-100 placeholder:text-slate-400"
      />
    </div>
  );
}
