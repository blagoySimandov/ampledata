import type { ColumnMetadata, EnrichRequest } from "@/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useApi, useEnrich, useSourceData } from "@/hooks";
import { Plus, Settings2, Loader2 } from "lucide-react";
import { useState } from "react";
import type { AddColumnsDialogProps } from "./types";
import { ColumnEditor } from "./column-editor";
import { EmptyFieldsPlaceholder } from "./empty-fields-placeholder";

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
      const payload: EnrichRequest = { columns_metadata: columnsMetadata };
      if (selectedKeyColumns.length > 0)
        payload.key_columns = selectedKeyColumns;
      const trimmedDescription = keyColumnDescription.trim();
      if (trimmedDescription)
        payload.key_column_description = trimmedDescription;

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
      <DialogContent
        className="sm:max-w-[560px] max-h-[90vh] flex flex-col p-0 overflow-hidden"
        onOpenAutoFocus={(e) => e.preventDefault()}
      >
        <DialogHeader className="p-6 pb-0">
          <DialogTitle className="text-2xl font-black">
            Add Enrichment Columns
          </DialogTitle>
        </DialogHeader>
        <ScrollArea className="flex-1 px-6">
          <div className="py-6 space-y-6">
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
