import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import type { RowData, ConfidenceConfig } from "../../types";
import { Link2, Info } from "lucide-react";
import { ConfidenceHeader } from "./header";
import { SourcesList } from "./source-list";
import { CellContent } from "./cell-content";
import { StageFooter } from "./stage-footer";
import { ExtractionHistory } from "./extraction-history";

interface ConfidenceDataRendererParams {
  colDef: { field: string };
  data: RowData;
  value: unknown;
}

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

export function ConfidenceDataRenderer(params: ConfidenceDataRendererParams) {
  const field = params.colDef.field;
  const confidence = params.data.__confidence?.[field];
  const stage = params.data.__stages?.[field];
  const sources = params.data.__sources?.[field];
  const extractionHistory = params.data.__extractionHistory?.[field] ?? [];
  const hasValue =
    params.value !== undefined && params.value !== null && params.value !== "";

  const confConfig = confidence
    ? getConfidenceConfig(confidence.score)
    : UNKNOWN_CONFIDENCE;

  const content = (
    <CellContent hasValue={hasValue} value={params.value} stage={stage} />
  );

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
          {extractionHistory.length > 0 && (
            <ExtractionHistory history={extractionHistory} />
          )}
          {stage && <StageFooter stage={stage} />}
        </div>
      </PopoverContent>
    </Popover>
  );
}
