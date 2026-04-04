import type { SourceJobSummary } from "@/api";
import { Badge } from "@/components/ui/badge";

interface JobRunCardProps {
  job: SourceJobSummary;
  isSelected: boolean;
  onSelect: () => void;
}

const JOB_STATUS_STYLES: Record<string, string> = {
  COMPLETED: "bg-emerald-50 text-emerald-700 border-emerald-200",
  RUNNING: "bg-blue-50 text-blue-700 border-blue-200",
  CANCELLED: "bg-red-50 text-red-700 border-red-200",
};

const DEFAULT_JOB_STATUS_STYLE = "bg-slate-100 text-slate-600 border-slate-200";

export function JobRunCard({ job, isSelected, onSelect }: JobRunCardProps) {
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
          <Badge
            variant="outline"
            className="text-[10px] px-1.5 py-0 h-4 min-h-0 bg-slate-50 text-slate-500"
          >
            {job.cost_credits} credits
          </Badge>
          {job.key_columns && job.key_columns.length > 0 && (
            <span className="truncate">Key: {job.key_columns.join(", ")}</span>
          )}
        </div>
      </div>
    </button>
  );
}
