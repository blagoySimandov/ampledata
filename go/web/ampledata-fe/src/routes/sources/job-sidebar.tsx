import type { SourceJobSummary } from "@/api";
import { Badge } from "@/components/ui/badge";
import { Settings2 } from "lucide-react";

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

export function JobRunsSidebar({
  jobs,
  selectedJobId,
  onSelect,
}: JobRunsSidebarProps) {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-slate-50/50">
      <div className="p-4 border-b border-slate-100 flex items-center justify-between bg-white/50 backdrop-blur-sm sticky top-0 z-10">
        <h2 className="text-xs font-black uppercase tracking-widest text-slate-500 flex items-center gap-2">
          <Settings2 className="w-3.5 h-3.5" /> All Runs
        </h2>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2.5">
        {jobs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center text-slate-400 py-12">
            <Settings2 className="w-8 h-8 mb-3 opacity-40" />
            <p className="font-bold text-sm">No enrichment runs yet</p>
            <p className="text-xs mt-1">
              Click "ADD COLUMNS" to start your first enrichment.
            </p>
          </div>
        ) : (
          jobs.map((job) => (
            <JobRunCard
              key={job.job_id}
              job={job}
              isSelected={job.job_id === selectedJobId}
              onSelect={() => onSelect(job.job_id)}
            />
          ))
        )}
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
