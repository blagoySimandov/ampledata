import { AlertCircle, PackageOpen, RefreshCw } from "lucide-react";
import { CellBadge } from "./cell-badge";

const TERMINAL_STAGES = new Set(["COMPLETED", "FAILED", "CANCELLED"]);

export function CellContent({
  hasValue,
  value,
  stage,
}: {
  hasValue: boolean;
  value: unknown;
  stage: string | undefined;
}) {
  if (hasValue) {
    return <span className="font-medium text-slate-900">{String(value)}</span>;
  }

  if (!stage) {
    return (
      <div className="h-2.5 bg-slate-200/80 rounded w-full max-w-[80%] animate-pulse" />
    );
  }

  if (!TERMINAL_STAGES.has(stage)) {
    return (
      <CellBadge variant="blue" icon={RefreshCw} spin>
        {stage.replace(/_/g, " ")}
      </CellBadge>
    );
  }

  if (stage === "FAILED" || stage === "CANCELLED") {
    return (
      <CellBadge variant="red" icon={AlertCircle}>
        {stage}
      </CellBadge>
    );
  }

  return (
    <CellBadge variant="amber" icon={PackageOpen}>
      Missing
    </CellBadge>
  );
}
