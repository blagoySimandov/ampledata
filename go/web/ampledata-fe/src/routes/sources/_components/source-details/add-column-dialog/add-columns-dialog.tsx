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
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Plus, Settings2, Loader2, ChevronDown } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import type { AddColumnsDialogProps } from "./types";
import { ColumnEditor } from "./column-editor";
import { EmptyFieldsPlaceholder } from "./empty-fields-placeholder";

const COMMON_DOMAINS = [
  "linkedin.com",
  "crunchbase.com",
  "bloomberg.com",
  "reuters.com",
  "wikipedia.org",
];

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
  const [allowedDomains, setAllowedDomains] = useState<string[]>([]);
  const [customDomain, setCustomDomain] = useState("");

  const canStart =
    columnsMetadata.length > 0 && columnsMetadata.every((c) => c.name);

  const hasImputation = columnsMetadata.some((c) => c.job_type === "imputation");
  const hasEnrichment = columnsMetadata.some((c) => c.job_type === "enrichment");
  const startLabel = hasImputation && !hasEnrichment ? "START IMPUTATION" : hasImputation ? "START RUN" : "START ENRICHMENT";

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

  const toggleDomain = (domain: string) =>
    setAllowedDomains((prev) =>
      prev.includes(domain) ? prev.filter((d) => d !== domain) : [...prev, domain],
    );

  const addCustomDomain = () => {
    const domain = customDomain.trim().replace(/^https?:\/\//, "").replace(/\/.*$/, "");
    if (domain && !allowedDomains.includes(domain)) {
      setAllowedDomains((prev) => [...prev, domain]);
    }
    setCustomDomain("");
  };

  const removeDomain = (domain: string) =>
    setAllowedDomains((prev) => prev.filter((d) => d !== domain));

  const resetForm = () => {
    setColumnsMetadata([]);
    setSelectedKeyColumns([]);
    setKeyColumnDescription("");
    setAllowedDomains([]);
    setCustomDomain("");
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
      if (allowedDomains.length > 0)
        payload.allowed_domains = allowedDomains;

      await enrich.mutateAsync(payload);
      setOpen(false);
      resetForm();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to start job");
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
            Add Columns
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
                <Collapsible>
                  <CollapsibleTrigger className="flex items-center gap-1 text-xs text-slate-400 font-bold hover:text-slate-600 transition-colors pt-2 group">
                    <ChevronDown className="w-3 h-3 transition-transform group-data-[state=open]:rotate-180" />
                    Advanced
                  </CollapsibleTrigger>
                  <CollapsibleContent className="space-y-1 pt-2">
                    <Label className="text-xs text-slate-500 font-bold">
                      Key Column Definition (AI Context)
                    </Label>
                    <Input
                      placeholder="Optional rules to locate the key columns in raw text..."
                      value={keyColumnDescription}
                      onChange={(e) => setKeyColumnDescription(e.target.value)}
                      className="text-sm h-9"
                    />
                  </CollapsibleContent>
                </Collapsible>
              </div>
            </div>

            <div className="space-y-4">
              <Label className="text-xs font-black uppercase tracking-widest text-slate-400 flex items-center gap-2">
                <Settings2 className="w-3 h-3" /> Source Domains
              </Label>
              <p className="text-[10px] text-slate-400 leading-tight">
                Restrict searches to specific websites. Leave empty to search everywhere.
              </p>
              <div className="flex flex-wrap gap-1.5">
                {COMMON_DOMAINS.map((domain) => {
                  const isSelected = allowedDomains.includes(domain);
                  return (
                    <Badge
                      key={domain}
                      variant={isSelected ? "default" : "outline"}
                      className={`cursor-pointer transition-colors ${
                        isSelected
                          ? ""
                          : "text-slate-500 hover:bg-slate-100 hover:text-slate-900 border-slate-200"
                      }`}
                      onClick={() => toggleDomain(domain)}
                    >
                      {domain}
                    </Badge>
                  );
                })}
              </div>
              {allowedDomains.filter((d) => !COMMON_DOMAINS.includes(d)).length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {allowedDomains
                    .filter((d) => !COMMON_DOMAINS.includes(d))
                    .map((domain) => (
                      <Badge
                        key={domain}
                        variant="default"
                        className="cursor-pointer gap-1"
                        onClick={() => removeDomain(domain)}
                      >
                        {domain} ×
                      </Badge>
                    ))}
                </div>
              )}
              <div className="flex gap-2">
                <Input
                  placeholder="Add custom domain (e.g. example.com)"
                  value={customDomain}
                  onChange={(e) => setCustomDomain(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && addCustomDomain()}
                  className="text-sm h-8 flex-1"
                />
                <Button
                  variant="outline"
                  size="sm"
                  onClick={addCustomDomain}
                  disabled={!customDomain.trim()}
                  className="h-8 text-xs font-black px-3"
                >
                  ADD
                </Button>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <Label className="text-xs font-black uppercase tracking-widest text-slate-400 flex items-center gap-2">
                  <Settings2 className="w-3 h-3" /> Columns
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
                      sourceColumns={sourceColumns}
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
                startLabel
              )}
            </Button>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
