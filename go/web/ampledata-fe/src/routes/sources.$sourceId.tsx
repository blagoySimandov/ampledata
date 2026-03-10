import { createFileRoute, Link } from "@tanstack/react-router";
import {
  useApi,
  useSource,
  useEnrich,
  useAllJobsRows,
  useJobProgress,
  useSourceData,
} from "../hooks";
import { useMemo, useState } from "react";
import { AgGridReact } from "ag-grid-react";
import { type ColDef, type ColGroupDef, themeQuartz } from "ag-grid-community";
import {
  ArrowLeft,
  RefreshCw,
  AlertCircle,
  Info,
  PanelRightOpen,
  PanelRightClose,
  Link2,
  ExternalLink,
  Plus,
  Trash2,
  Settings2,
  Loader2,
  PlayCircle,
  StopCircle,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { ColumnMetadata, SourceJobSummary } from "../api";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface ConfidenceEntry {
  score: number;
  reason: string;
}

interface ConfidenceConfig {
  label: string;
  color: string;
  bg: string;
  borderColor: string;
}

interface RowData {
  __index: string;
  __confidence?: Record<string, ConfidenceEntry>;
  __stages?: Record<string, string>;
  __sources?: Record<string, string[]>;
  [field: string]: unknown;
}

interface ConfidenceDataRendererParams {
  colDef: { field: string };
  data: RowData;
  value: unknown;
}

interface MergedDataResult {
  rows: RowData[];
  sourceColumns: string[];
  enrichedColumns: string[];
}

interface EnrichPayload {
  columns_metadata: ColumnMetadata[];
  key_columns?: string[];
  key_column_description?: string;
}

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

const gridTheme = themeQuartz.withParams({
  accentColor: "#0f172a",
  backgroundColor: "#ffffff",
  borderColor: "#e2e8f0",
  borderRadius: "8px",
  browserColorScheme: "light",
  chromeBackgroundColor: "#f8fafc",
  fontFamily: "inherit",
  fontSize: "14px",
  foregroundColor: "#0f172a",
  headerBackgroundColor: "#f8fafc",
  headerFontSize: "14px",
  headerFontWeight: "600",
  headerTextColor: "#64748b",
  rowBorder: { color: "#f1f5f9" },
  wrapperBorder: true,
  wrapperBorderRadius: "12px",
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const UNKNOWN_CONFIDENCE: ConfidenceConfig = {
  label: "Unknown",
  color: "text-slate-400",
  bg: "bg-slate-200",
  borderColor: "border-slate-200",
};

function getConfidenceConfig(score: number): ConfidenceConfig {
  if (score >= 0.8)
    return {
      label: "Very High",
      color: "text-emerald-500",
      bg: "bg-emerald-500",
      borderColor: "border-emerald-200",
    };
  if (score >= 0.7)
    return {
      label: "High",
      color: "text-green-500",
      bg: "bg-green-500",
      borderColor: "border-green-200",
    };
  if (score >= 0.5)
    return {
      label: "Medium",
      color: "text-amber-500",
      bg: "bg-amber-500",
      borderColor: "border-amber-200",
    };
  return {
    label: "Low",
    color: "text-red-500",
    bg: "bg-red-500",
    borderColor: "border-red-200",
  };
}

function extractDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
}

const IN_PROGRESS_STAGES = new Set(["COMPLETED", "FAILED", "CANCELLED"]);

function isTerminalStage(stage: string): boolean {
  return IN_PROGRESS_STAGES.has(stage);
}

// ---------------------------------------------------------------------------
// ConfidenceDataRenderer
// ---------------------------------------------------------------------------

function ConfidenceDataRenderer(params: ConfidenceDataRendererParams) {
  const field = params.colDef.field;
  const confidence = params.data.__confidence?.[field];
  const stage = params.data.__stages?.[field];
  const sources = params.data.__sources?.[field];
  const hasValue =
    params.value !== undefined && params.value !== null && params.value !== "";

  const confConfig = confidence
    ? getConfidenceConfig(confidence.score)
    : UNKNOWN_CONFIDENCE;

  const content = renderCellContent(hasValue, params.value, stage);

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button className="flex items-center justify-between w-full group cursor-pointer text-left focus:outline-none h-full min-h-[32px]">
          <div className="flex-1 truncate mr-2">{content}</div>
          {(confidence || (sources && sources.length > 0)) && (
            <div
              className={`flex items-center gap-1 opacity-40 group-hover:opacity-100 transition-opacity ${confConfig.color}`}
            >
              {sources && sources.length > 0 && (
                <Link2 className="w-3.5 h-3.5 text-slate-400" />
              )}
              {confidence && <Info className="w-3.5 h-3.5" />}
            </div>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-80 shadow-2xl border-slate-200 p-0 overflow-hidden z-50">
        <div className={`h-1.5 w-full ${confConfig.bg}`} />
        <div className="p-4 space-y-4">
          <ConfidenceHeader config={confConfig} confidence={confidence} />
          {sources && sources.length > 0 && <SourcesList sources={sources} />}
          {stage && <StageFooter stage={stage} />}
        </div>
      </PopoverContent>
    </Popover>
  );
}

function renderCellContent(
  hasValue: boolean,
  value: unknown,
  stage: string | undefined,
) {
  if (hasValue) {
    return <span className="font-medium text-slate-900">{String(value)}</span>;
  }

  if (stage && !isTerminalStage(stage)) {
    return (
      <CellBadge variant="blue" icon={RefreshCw} spin>
        {stage.replace("_", " ")}
      </CellBadge>
    );
  }

  if (stage === "FAILED" || stage === "CANCELLED") {
    return (
      <CellBadge variant="red" icon={AlertCircle}>
        FAILED
      </CellBadge>
    );
  }

  if (stage === "COMPLETED") {
    return (
      <CellBadge variant="amber" icon={AlertCircle}>
        Missing
      </CellBadge>
    );
  }

  return (
    <div className="h-2.5 bg-slate-200/80 rounded w-full max-w-[80%] animate-pulse" />
  );
}

// ---------------------------------------------------------------------------
// Small presentational sub-components
// ---------------------------------------------------------------------------

const BADGE_VARIANTS = {
  blue: "bg-blue-50 text-blue-700 border-blue-200",
  red: "bg-red-50 text-red-700 border-red-200",
  amber: "bg-amber-50 text-amber-700 border-amber-200",
} as const;

type BadgeVariant = keyof typeof BADGE_VARIANTS;

function CellBadge({
  variant,
  icon: Icon,
  spin = false,
  children,
}: {
  variant: BadgeVariant;
  icon: LucideIcon;
  spin?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div
      className={`flex items-center gap-1.5 px-2 py-0.5 rounded border text-xs font-bold uppercase tracking-tight w-fit ${BADGE_VARIANTS[variant]}`}
    >
      <Icon className={`w-3 h-3 ${spin ? "animate-spin" : ""}`} /> {children}
    </div>
  );
}

function ConfidenceHeader({
  config,
  confidence,
}: {
  config: ConfidenceConfig;
  confidence: ConfidenceEntry | undefined;
}) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="font-black text-xs uppercase tracking-widest text-slate-400">
          Field Intelligence
        </h4>
        <Badge
          variant="outline"
          className={`text-xs font-black ${config.color} ${config.borderColor} bg-white`}
        >
          {config.label} CONFIDENCE
        </Badge>
      </div>
      <p className="text-xs text-slate-500 leading-relaxed italic">
        {confidence ? `"${confidence.reason}"` : "No rationale captured."}
      </p>
    </div>
  );
}

function SourcesList({ sources }: { sources: string[] }) {
  return (
    <div className="pt-3 border-t border-slate-100 space-y-2">
      <div className="flex items-center justify-between">
        <h4 className="font-black text-xs uppercase tracking-widest text-slate-400 flex items-center gap-1.5">
          <Link2 className="w-3 h-3" /> Sources
        </h4>
        <Badge
          variant="secondary"
          className="text-[10px] font-bold px-1.5 py-0 h-4 min-h-0"
        >
          {sources.length} LINKS
        </Badge>
      </div>
      <div className="max-h-[150px] overflow-y-auto space-y-1 pr-1">
        {sources.map((src, i) => {
          const domain = extractDomain(src);
          return (
            <a
              key={i}
              href={src}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center justify-between p-2 rounded-lg hover:bg-slate-50 border border-transparent hover:border-slate-100 transition-colors group"
            >
              <div className="flex items-center gap-2 truncate min-w-0">
                <div className="w-6 h-6 rounded bg-slate-100 flex items-center justify-center shrink-0 border border-slate-200">
                  <img
                    src={`https://icon.horse/icon/${domain}`}
                    alt=""
                    className="w-3 h-3 opacity-70"
                    onError={(e) => {
                      (e.currentTarget as HTMLImageElement).style.display =
                        "none";
                    }}
                  />
                </div>
                <span className="text-xs font-medium text-slate-700 truncate group-hover:text-primary transition-colors">
                  {domain}
                </span>
              </div>
              <ExternalLink className="w-3 h-3 text-slate-300 group-hover:text-primary shrink-0 ml-2 transition-colors" />
            </a>
          );
        })}
      </div>
    </div>
  );
}

function StageFooter({ stage }: { stage: string }) {
  return (
    <div className="pt-2 border-t border-slate-100 flex items-center justify-between">
      <span className="text-xs font-bold text-slate-400 uppercase tracking-tight">
        Stage
      </span>
      <span className="text-xs font-black text-slate-900 uppercase tracking-widest">
        {stage}
      </span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// AddColumnsDialog
// ---------------------------------------------------------------------------

interface AddColumnsDialogProps {
  sourceId: string;
  mostRecentJob?: SourceJobSummary;
}

function AddColumnsDialog({ sourceId, mostRecentJob }: AddColumnsDialogProps) {
  const api = useApi();
  const enrich = useEnrich(api, sourceId);
  const { data: sourceData } = useSourceData(api, sourceId);
  const sourceColumns = sourceData?.headers ?? [];

  const [open, setOpen] = useState(false);
  const [columnsMetadata, setColumnsMetadata] = useState<ColumnMetadata[]>([]);
  const [selectedKeyColumns, setSelectedKeyColumns] = useState<string[]>([]);
  const [keyColumnDescription, setKeyColumnDescription] = useState("");

  const hasExistingJob = mostRecentJob !== undefined;
  const canStart =
    columnsMetadata.length > 0 &&
    columnsMetadata.every((c) => c.name) &&
    hasExistingJob;

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
        <Button className="font-bold gap-2" disabled={!hasExistingJob}>
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

// ---------------------------------------------------------------------------
// JobStats
// ---------------------------------------------------------------------------

const STATUS_ICON_MAP: Record<string, { icon: LucideIcon; color: string }> = {
  COMPLETED: { icon: CheckCircle2, color: "text-emerald-500" },
  CANCELLED: { icon: XCircle, color: "text-red-500" },
  PAUSED: { icon: StopCircle, color: "text-amber-500" },
};

const DEFAULT_STATUS_ICON = { icon: PlayCircle, color: "text-blue-500" };

function JobStats({ jobId }: { jobId: string }) {
  const api = useApi();
  const { data: progress } = useJobProgress(api, jobId);

  if (!progress) {
    return (
      <div className="p-4 border-b border-slate-100 bg-slate-50/50 shrink-0 flex items-center justify-center h-24">
        <Loader2 className="w-5 h-5 text-slate-400 animate-spin" />
      </div>
    );
  }

  const stages = progress.rows_by_stage ?? {};
  const completed = stages["COMPLETED"] ?? 0;
  const failed = stages["FAILED"] ?? 0;
  const pending = stages["PENDING"] ?? 0;
  const inProgress = progress.total_rows - completed - failed - pending;
  const pctCompleted =
    progress.total_rows > 0 ? (completed / progress.total_rows) * 100 : 0;

  const { icon: StatusIcon, color: statusColor } =
    STATUS_ICON_MAP[progress.status] ?? DEFAULT_STATUS_ICON;

  return (
    <div className="p-6 border-b border-slate-100 bg-white shrink-0 space-y-4 shadow-sm relative z-10">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <StatusIcon className={`w-5 h-5 ${statusColor}`} />
            <h3 className="text-sm font-black text-slate-900 tracking-tight">
              {progress.status}
            </h3>
          </div>
          <p className="text-xs text-slate-500 font-medium ml-7">
            Started {new Date(progress.started_at).toLocaleString()}
          </p>
        </div>
        <Badge
          variant="secondary"
          className="font-bold bg-slate-100 text-slate-700"
        >
          {progress.total_rows} ROWS
        </Badge>
      </div>

      <div className="space-y-2 pt-2">
        <div className="flex justify-between text-xs font-bold text-slate-500 mb-1">
          <span>Overall Progress</span>
          <span className="text-slate-900">{pctCompleted.toFixed(1)}%</span>
        </div>
        <Progress value={pctCompleted} className="h-2.5 bg-slate-100" />
      </div>

      <div className="grid grid-cols-3 gap-2 pt-2">
        <StatCard value={completed} label="Done" variant="emerald" />
        <StatCard value={inProgress} label="Active" variant="blue" />
        <StatCard value={failed} label="Failed" variant="red" />
      </div>
    </div>
  );
}

type StatVariant = "emerald" | "blue" | "red";

const STAT_STYLES: Record<StatVariant, string> = {
  emerald:
    "bg-emerald-50 border-emerald-100 text-emerald-600 [&_.label]:text-emerald-700/70",
  blue: "bg-blue-50 border-blue-100 text-blue-600 [&_.label]:text-blue-700/70",
  red: "bg-red-50 border-red-100 text-red-600 [&_.label]:text-red-700/70",
};

function StatCard({
  value,
  label,
  variant,
}: {
  value: number;
  label: string;
  variant: StatVariant;
}) {
  return (
    <div
      className={`border rounded-lg p-2.5 flex flex-col items-center justify-center ${STAT_STYLES[variant]}`}
    >
      <span className="text-lg font-black">{value}</span>
      <span className="label text-[10px] font-bold uppercase tracking-widest mt-0.5">
        {label}
      </span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// JobRunsSidebar
// ---------------------------------------------------------------------------

interface JobRunsSidebarProps {
  jobs: SourceJobSummary[];
  selectedJobId: string | null;
  onSelect: (jobId: string) => void;
  onClose: () => void;
}

const JOB_STATUS_STYLES: Record<string, string> = {
  COMPLETED: "bg-emerald-50 text-emerald-700 border-emerald-200",
  RUNNING: "bg-blue-50 text-blue-700 border-blue-200",
  CANCELLED: "bg-red-50 text-red-700 border-red-200",
};

const DEFAULT_JOB_STATUS_STYLE = "bg-slate-100 text-slate-600 border-slate-200";

function JobRunsSidebar({
  jobs,
  selectedJobId,
  onSelect,
  onClose,
}: JobRunsSidebarProps) {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-slate-50/50">
      <div className="p-4 border-b border-slate-100 flex items-center justify-between bg-white/50 backdrop-blur-sm sticky top-0 z-10">
        <h2 className="text-xs font-black uppercase tracking-widest text-slate-500 flex items-center gap-2">
          <Settings2 className="w-3.5 h-3.5" /> All Runs
        </h2>
        <button
          onClick={onClose}
          className="bg-white border border-slate-200 p-1.5 rounded-md shadow-sm hover:bg-slate-50 transition-all text-slate-400 hover:text-slate-700 active:scale-95"
        >
          <PanelRightClose className="w-3.5 h-3.5" />
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2.5">
        {jobs.map((job) => (
          <JobRunCard
            key={job.job_id}
            job={job}
            isSelected={job.job_id === selectedJobId}
            onSelect={() => onSelect(job.job_id)}
          />
        ))}
      </div>
    </div>
  );
}

interface JobRunCardProps {
  job: SourceJobSummary;
  isSelected: boolean;
  onSelect: () => void;
}

function JobRunCard({ job, isSelected, onSelect }: JobRunCardProps) {
  const columnsLabel =
    job.columns_metadata?.map((c) => c.name).join(", ") || "—";
  const statusStyle = JOB_STATUS_STYLES[job.status] ?? DEFAULT_JOB_STATUS_STYLE;

  return (
    <button
      onClick={onSelect}
      className={`w-full text-left p-3.5 rounded-xl border transition-all ${
        isSelected
          ? "border-primary bg-white shadow-md ring-1 ring-primary/10"
          : "border-slate-200 bg-white hover:border-slate-300 hover:shadow-sm"
      }`}
    >
      <div className="flex items-center justify-between mb-2">
        <span
          className={`text-[10px] font-bold px-2 py-0.5 rounded border uppercase tracking-wider ${statusStyle}`}
        >
          {job.status}
        </span>
        <span className="text-[11px] font-medium text-slate-400">
          {new Date(job.created_at).toLocaleDateString()}
        </span>
      </div>
      <div className="space-y-1.5">
        <p className="text-xs text-slate-600 leading-relaxed line-clamp-2">
          <span className="font-bold text-slate-900">Fields: </span>
          {columnsLabel}
        </p>
        <div className="flex items-center gap-2 text-[11px] font-medium text-slate-500">
          <Badge
            variant="outline"
            className="text-[10px] px-1.5 py-0 h-4 min-h-0 bg-slate-50 text-slate-500"
          >
            {job.total_rows} rows
          </Badge>
          {job.key_columns && job.key_columns.length > 0 && (
            <span className="truncate">Key: {job.key_columns.join(", ")}</span>
          )}
        </div>
      </div>
    </button>
  );
}

// ---------------------------------------------------------------------------
// DataTable
// ---------------------------------------------------------------------------

interface DataTableProps {
  sourceId: string;
  jobs: SourceJobSummary[];
}

function useMergedData(
  sourceId: string,
  jobs: SourceJobSummary[],
): { data: MergedDataResult; isFetching: boolean } {
  const api = useApi();
  const { data: sourceData, isFetching: sourceFetching } = useSourceData(
    api,
    sourceId,
  );
  const jobQueries = useAllJobsRows(api, jobs);

  const isFetching = sourceFetching || jobQueries.some((q) => q.isFetching);

  const mergedData = useMemo<MergedDataResult>(() => {
    if (!sourceData) {
      return { rows: [], sourceColumns: [], enrichedColumns: [] };
    }

    const rowMap = new Map<string, RowData>();
    const enrichedCols = new Set<string>();

    // Build base rows from source CSV
    sourceData.rows.forEach((csvRow, index) => {
      const rowObj: RowData = { __index: index.toString() };
      sourceData.headers.forEach((header, i) => {
        rowObj[header] = csvRow[i];
      });
      rowMap.set(index.toString(), rowObj);
    });

    // Process from oldest to newest so newer runs overwrite
    const sortedQueries = [...jobQueries].reverse();
    const sortedJobs = [...jobs].reverse();

    sortedQueries.forEach((query, qIndex) => {
      if (!query.data) return;

      const job = sortedJobs[qIndex];
      const keyCols = job.key_columns ?? [];
      if (keyCols.length === 0) return;

      const keyIndices = keyCols.map((kc) => sourceData.headers.indexOf(kc));

      const jobRowsByKey = new Map<string, (typeof query.data.rows)[number]>();
      for (const row of query.data.rows) {
        jobRowsByKey.set(row.key, row);
      }

      sourceData.rows.forEach((csvRow, rowIndex) => {
        const jobKey = keyIndices
          .map((idx) => (idx !== -1 && idx < csvRow.length ? csvRow[idx] : ""))
          .join("||");

        const jobRow = jobRowsByKey.get(jobKey);
        if (!jobRow) return;

        const existing = rowMap.get(rowIndex.toString());
        if (!existing) return;

        if (jobRow.extracted_data) {
          Object.assign(existing, jobRow.extracted_data);
          for (const key of Object.keys(jobRow.extracted_data)) {
            enrichedCols.add(key);
          }
        }

        if (jobRow.confidence) {
          existing.__confidence ??= {};
          Object.assign(existing.__confidence, jobRow.confidence);
          for (const key of Object.keys(jobRow.confidence)) {
            enrichedCols.add(key);
          }
        }

        if (jobRow.sources) {
          existing.__sources ??= {};
          if (job.columns_metadata) {
            for (const col of job.columns_metadata) {
              existing.__sources[col.name] = jobRow.sources;
            }
          } else if (jobRow.extracted_data) {
            for (const key of Object.keys(jobRow.extracted_data)) {
              existing.__sources[key] = jobRow.sources;
            }
          }
        }

        if (job.columns_metadata) {
          existing.__stages ??= {};
          for (const col of job.columns_metadata) {
            existing.__stages[col.name] = jobRow.stage;
          }
        }
      });
    });

    return {
      rows: Array.from(rowMap.values()),
      sourceColumns: sourceData.headers,
      enrichedColumns: Array.from(enrichedCols).sort(),
    };
  }, [sourceData, jobQueries, jobs]);

  return { data: mergedData, isFetching };
}

function DataTable({ sourceId, jobs }: DataTableProps) {
  const { data: mergedData, isFetching } = useMergedData(sourceId, jobs);

  const columnDefs = useMemo<(ColDef | ColGroupDef)[]>(() => {
    if (mergedData.sourceColumns.length === 0) return [];

    return [
      {
        headerName: "Original Data",
        children: mergedData.sourceColumns.map(
          (col, idx): ColDef => ({
            field: col,
            headerName: col,
            pinned: idx === 0 ? "left" : undefined,
            minWidth: 150,
            cellClass: "bg-slate-50 font-medium",
          }),
        ),
      },
      {
        headerName: "Enriched Data",
        children: mergedData.enrichedColumns.map(
          (col): ColDef => ({
            field: col,
            headerName: col,
            flex: 1,
            minWidth: 150,
            cellRenderer: ConfidenceDataRenderer,
          }),
        ),
      },
    ];
  }, [mergedData.sourceColumns, mergedData.enrichedColumns]);

  return (
    <div className="bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col flex-1 min-h-0">
      <div className="flex items-center gap-2 px-4 py-2 border-b border-slate-100 bg-slate-50 shrink-0">
        <span className="text-xs font-black uppercase tracking-widest text-slate-400">
          Merged Data View
        </span>
        {isFetching && (
          <RefreshCw className="w-3 h-3 text-blue-600 animate-spin" />
        )}
      </div>
      <div className="w-full flex-1 min-h-0">
        <AgGridReact
          rowData={mergedData.rows}
          columnDefs={columnDefs}
          theme={gridTheme}
          pagination
          paginationPageSize={100}
          paginationPageSizeSelector={[50, 100, 200, 500]}
          defaultColDef={{ resizable: true, sortable: true, filter: true }}
        />
      </div>
      <div className="p-4 border-t border-slate-100 bg-slate-50 flex justify-between items-center shrink-0">
        <div className="text-xs font-bold text-slate-400 uppercase tracking-widest">
          {mergedData.rows.length} total records
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

function SourceDetailPage() {
  const { sourceId } = Route.useParams();
  const api = useApi();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const { data: source, isLoading, isError } = useSource(api, sourceId);

  const mostRecentJob = source?.jobs[0];
  const activeJobId = selectedJobId ?? mostRecentJob?.job_id ?? null;
  const activeJob = source?.jobs.find((j) => j.job_id === activeJobId);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 text-gray-500 gap-3 font-medium">
        <RefreshCw className="w-5 h-5 animate-spin" /> Loading source...
      </div>
    );
  }

  if (isError || !source) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardContent className="pt-6 text-red-700 font-medium">
          Failed to load source.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="fixed left-0 right-0 top-16 bottom-0 flex bg-slate-50/50 animate-in fade-in duration-500 z-0 overflow-hidden">
      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0 transition-all duration-300 h-full p-4 sm:p-6 lg:p-8">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 shrink-0">
          <div className="flex items-center gap-4">
            <Link
              to="/"
              className="inline-flex items-center text-xs font-bold text-slate-500 hover:text-primary transition-colors gap-1 uppercase tracking-widest bg-white px-3 py-2 rounded-lg border border-slate-200 shadow-sm active:scale-95"
            >
              <ArrowLeft className="w-3.5 h-3.5" /> BACK
            </Link>
            <h2 className="text-2xl font-black text-slate-900 tracking-tight">
              Dataset Explorer
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <AddColumnsDialog
              sourceId={sourceId}
              mostRecentJob={mostRecentJob}
            />
            {!sidebarOpen && (
              <button
                onClick={() => setSidebarOpen(true)}
                className="bg-white border border-slate-200 text-slate-600 px-4 py-2 rounded-lg shadow-sm hover:bg-slate-50 transition-all active:scale-95 flex items-center gap-2 font-bold text-xs uppercase tracking-widest"
              >
                <PanelRightOpen className="w-4 h-4" /> RUNS
              </button>
            )}
          </div>
        </div>

        {source.jobs.length > 0 ? (
          <DataTable sourceId={sourceId} jobs={source.jobs} />
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center text-slate-400">
              <Settings2 className="w-12 h-12 mx-auto mb-3 opacity-40" />
              <p className="font-bold">No enrichment runs yet</p>
              <p className="text-sm mt-1">
                Click "ADD COLUMNS" to start your first enrichment.
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Sidebar */}
      <div
        className={`shrink-0 bg-white border-l border-slate-200 shadow-2xl transition-all duration-300 ease-in-out h-full flex flex-col z-10 relative overflow-hidden ${
          sidebarOpen
            ? "w-[400px] opacity-100"
            : "w-0 opacity-0 border-transparent pointer-events-none"
        }`}
      >
        {sidebarOpen && (
          <>
            {activeJob && <JobStats jobId={activeJob.job_id} />}
            <JobRunsSidebar
              jobs={source.jobs}
              selectedJobId={activeJobId}
              onSelect={setSelectedJobId}
              onClose={() => setSidebarOpen(false)}
            />
          </>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute("/sources/$sourceId")({
  component: SourceDetailPage,
});
