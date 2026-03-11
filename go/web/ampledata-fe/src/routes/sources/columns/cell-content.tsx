import { RefreshCw, AlertCircle, type LucideIcon } from "lucide-react";

const IN_PROGRESS_STAGES = new Set(["COMPLETED", "FAILED", "CANCELLED"]);

function isTerminalStage(stage: string): boolean {
  return IN_PROGRESS_STAGES.has(stage);
}

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
