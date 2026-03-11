import type { ConfidenceConfig, ConfidenceEntry } from "../../types";
import { Badge } from "@/components/ui/badge";

export function ConfidenceHeader({
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
