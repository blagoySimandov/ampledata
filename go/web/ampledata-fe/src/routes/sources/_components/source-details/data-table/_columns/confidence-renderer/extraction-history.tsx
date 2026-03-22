import { History, ChevronDown, ChevronUp } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import type { HistoryEntryForField } from "../../types";

function confidenceColor(score: number): string {
  if (score >= 0.8) return "text-emerald-500";
  if (score >= 0.7) return "text-green-500";
  if (score >= 0.5) return "text-amber-500";
  return "text-red-500";
}

function confidenceBg(score: number): string {
  if (score >= 0.8) return "bg-emerald-500";
  if (score >= 0.7) return "bg-green-500";
  if (score >= 0.5) return "bg-amber-500";
  return "bg-red-500";
}

export function ExtractionHistory({
  history,
}: {
  history: HistoryEntryForField[];
}) {
  const [expanded, setExpanded] = useState(false);

  if (history.length < 2) return null;

  return (
    <div className="pt-3 border-t border-slate-100 space-y-2">
      <button
        onClick={() => setExpanded((v) => !v)}
        className="flex items-center justify-between w-full group"
      >
        <h4 className="font-black text-xs uppercase tracking-widest text-slate-400 flex items-center gap-1.5">
          <History className="w-3 h-3" /> Attempt History
        </h4>
        <div className="flex items-center gap-1.5">
          <Badge
            variant="secondary"
            className="text-[10px] font-bold px-1.5 py-0 h-4 min-h-0"
          >
            {history.length} ATTEMPTS
          </Badge>
          {expanded ? (
            <ChevronUp className="w-3 h-3 text-slate-400" />
          ) : (
            <ChevronDown className="w-3 h-3 text-slate-400" />
          )}
        </div>
      </button>

      {expanded && (
        <div className="space-y-2">
          {history.map((entry) => (
            <div
              key={entry.attempt_number}
              className="rounded-lg border border-slate-100 bg-slate-50 p-2.5 space-y-1.5"
            >
              <div className="flex items-center justify-between">
                <span className="text-[10px] font-black uppercase tracking-widest text-slate-400">
                  Attempt {entry.attempt_number}
                </span>
                {entry.confidence && (
                  <div className="flex items-center gap-1.5">
                    <div
                      className={`w-1.5 h-1.5 rounded-full ${confidenceBg(entry.confidence.score)}`}
                    />
                    <span
                      className={`text-[10px] font-bold ${confidenceColor(entry.confidence.score)}`}
                    >
                      {Math.round(entry.confidence.score * 100)}%
                    </span>
                  </div>
                )}
              </div>

              <p className="text-xs font-medium text-slate-700">
                {entry.value !== undefined && entry.value !== null
                  ? String(entry.value)
                  : <span className="italic text-slate-400">no value</span>}
              </p>

              {entry.confidence?.reason && (
                <p className="text-[11px] text-slate-400 italic leading-snug">
                  "{entry.confidence.reason}"
                </p>
              )}

              {entry.reasoning && (
                <p className="text-[11px] text-slate-500 leading-snug border-t border-slate-200 pt-1.5 mt-1.5">
                  {entry.reasoning}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
