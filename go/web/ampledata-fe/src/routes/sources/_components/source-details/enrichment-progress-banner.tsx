import { useApi, useJobProgress } from "@/hooks";
import type { SourceJobSummary } from "@/api";
import { Loader2, CheckCircle2, XCircle } from "lucide-react";

const STAGE_LABEL: Record<string, string> = {
  PENDING: "Queued",
  SERP_FETCHED: "Searching",
  DECISION_MADE: "Analysing",
  CRAWLED: "Reading sources",
  ENRICHED: "Extracting",
  COMPLETED: "Done",
  FAILED: "Failed",
  CANCELLED: "Cancelled",
};

function dominantActiveStage(rowsByStage: Record<string, number>): string {
  const order = ["ENRICHED", "CRAWLED", "DECISION_MADE", "SERP_FETCHED", "PENDING"];
  for (const stage of order) {
    if ((rowsByStage[stage] ?? 0) > 0) return stage;
  }
  return "PENDING";
}

interface EnrichmentProgressBannerProps {
  activeJob: SourceJobSummary;
}

export function EnrichmentProgressBanner({ activeJob }: EnrichmentProgressBannerProps) {
  const api = useApi();
  const { data: progress } = useJobProgress(api, activeJob.job_id, 3000);

  const isActive = activeJob.status === "RUNNING" || activeJob.status === "PENDING";
  if (!isActive) return null;

  if (!progress) {
    return (
      <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 border-b border-blue-100 text-xs text-blue-600 font-medium">
        <Loader2 className="w-3.5 h-3.5 animate-spin shrink-0" />
        <span>Starting enrichment job…</span>
      </div>
    );
  }

  const stages = progress.rows_by_stage ?? {};
  const completed = stages["COMPLETED"] ?? 0;
  const failed = stages["FAILED"] ?? 0;
  const total = progress.total_rows;
  const done = completed + failed;
  const pct = total > 0 ? (completed / total) * 100 : 0;
  const activeStage = dominantActiveStage(stages);
  const stageLabel = STAGE_LABEL[activeStage] ?? activeStage;

  const isCompleted = progress.status === "COMPLETED";
  const isFailed = progress.status === "CANCELLED" || (done === total && failed > 0 && completed === 0);

  if (isCompleted) {
    return (
      <div className="flex items-center gap-2 px-4 py-2 bg-emerald-50 border-b border-emerald-100 text-xs text-emerald-700 font-medium">
        <CheckCircle2 className="w-3.5 h-3.5 shrink-0" />
        <span>Enrichment complete — {completed} rows enriched</span>
      </div>
    );
  }

  if (isFailed) {
    return (
      <div className="flex items-center gap-2 px-4 py-2 bg-red-50 border-b border-red-100 text-xs text-red-700 font-medium">
        <XCircle className="w-3.5 h-3.5 shrink-0" />
        <span>Job stopped — {failed} rows failed</span>
      </div>
    );
  }

  return (
    <div className="px-4 py-2.5 bg-blue-50 border-b border-blue-100 space-y-1.5">
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-1.5 text-blue-700 font-semibold">
          <Loader2 className="w-3.5 h-3.5 animate-spin shrink-0" />
          <span>{stageLabel}…</span>
        </div>
        <span className="text-blue-600 font-bold tabular-nums">
          {completed} / {total} rows · {pct.toFixed(0)}%
        </span>
      </div>
      <div className="w-full bg-blue-100 rounded-full h-1.5 overflow-hidden">
        <div
          className="bg-blue-500 h-full rounded-full transition-all duration-700 ease-out"
          style={{ width: `${Math.max(pct, 2)}%` }}
        />
      </div>
    </div>
  );
}
