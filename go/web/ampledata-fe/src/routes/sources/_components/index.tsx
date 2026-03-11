import { Badge } from "@/components/ui/badge";
import { Link2, ExternalLink } from "lucide-react";
import type { ConfidenceConfig, ConfidenceEntry } from "../types";

function extractDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
}
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

export function SourcesList({ sources }: { sources: string[] }) {
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

export function StageFooter({ stage }: { stage: string }) {
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
