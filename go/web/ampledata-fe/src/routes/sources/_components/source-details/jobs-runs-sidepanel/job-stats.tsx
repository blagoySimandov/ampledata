import { useApi, useJobProgress, useCancelJob } from "@/hooks";
import {
  type LucideIcon,
  CheckCircle2,
  XCircle,
  StopCircle,
  PlayCircle,
  Loader2,
  Ban,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";

const STATUS_ICON_MAP: Record<string, { icon: LucideIcon; color: string }> = {
  COMPLETED: { icon: CheckCircle2, color: "text-emerald-500" },
  CANCELLED: { icon: XCircle, color: "text-red-500" },
  PAUSED: { icon: StopCircle, color: "text-amber-500" },
};

const DEFAULT_STATUS_ICON = { icon: PlayCircle, color: "text-blue-500" };

const CANCELLABLE_STATUSES = new Set(["RUNNING", "PENDING"]);

export function JobStats({ jobId }: { jobId: string }) {
  const api = useApi();
  const { data: progress } = useJobProgress(api, jobId);
  const { mutate: cancelJob, isPending: isCancelling } = useCancelJob(api);

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

  const canCancel = CANCELLABLE_STATUSES.has(progress.status);

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
        <div className="flex items-center gap-2">
          {canCancel && (
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2.5 text-xs font-bold text-red-600 border-red-200 hover:bg-red-50 hover:border-red-300 hover:text-red-700"
              onClick={() => cancelJob(jobId)}
              disabled={isCancelling}
            >
              {isCancelling ? (
                <Loader2 className="w-3 h-3 animate-spin" />
              ) : (
                <Ban className="w-3 h-3" />
              )}
              {isCancelling ? "Cancelling…" : "Cancel"}
            </Button>
          )}
          <Badge
            variant="secondary"
            className="font-bold bg-slate-100 text-slate-700"
          >
            {progress.total_rows} ROWS
          </Badge>
        </div>
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
