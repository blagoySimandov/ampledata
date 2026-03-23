import { useEffect, useState } from "react";
import { HERO_ENRICHED_HEADERS, HERO_SOURCE_HEADERS, MOCK_ROWS } from "./constants";

function SkeletonCell({ active }: { active?: boolean }) {
  return (
    <div
      className={`h-4 rounded w-20 ${active ? "bg-primary/20 animate-pulse" : "bg-slate-200 animate-pulse"}`}
    />
  );
}

export function EnrichmentMockHero() {
  const [enrichedCount, setEnrichedCount] = useState(0);
  const total = MOCK_ROWS.length;
  const progress = Math.round((enrichedCount / total) * 100);

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout>;
    if (enrichedCount < total) {
      timeoutId = setTimeout(() => setEnrichedCount((count) => count + 1), 1400);
    } else {
      timeoutId = setTimeout(() => setEnrichedCount(0), 2800);
    }
    return () => clearTimeout(timeoutId);
  }, [enrichedCount, total]);

  return (
    <div className="bg-white rounded-2xl shadow-2xl border border-slate-200 overflow-hidden text-left w-full">
      <div className="flex items-center gap-2 px-4 py-2.5 bg-slate-50 border-b border-slate-200">
        <div className="flex gap-1.5 shrink-0">
          <div className="w-3 h-3 rounded-full bg-red-400" />
          <div className="w-3 h-3 rounded-full bg-yellow-400" />
          <div className="w-3 h-3 rounded-full bg-green-400" />
        </div>
        <div className="flex-1 h-6 rounded bg-slate-200/70 flex items-center px-3 text-[11px] text-slate-400 font-medium select-none">
          Dataset Explorer — companies.csv
        </div>
        {enrichedCount < total ? (
          <span className="text-[10px] font-black text-blue-600 bg-blue-50 border border-blue-100 px-2 py-0.5 rounded-full animate-pulse shrink-0">
            Enriching…
          </span>
        ) : (
          <span className="text-[10px] font-black text-emerald-600 bg-emerald-50 border border-emerald-100 px-2 py-0.5 rounded-full shrink-0">
            ✓ Complete
          </span>
        )}
      </div>

      <div className="grid grid-cols-5 bg-slate-50/60 border-b border-slate-100">
        {HERO_SOURCE_HEADERS.map((header) => (
          <div
            key={header}
            className="px-3 py-2 text-[10px] font-black uppercase tracking-widest text-slate-400"
          >
            {header}
          </div>
        ))}
        {HERO_ENRICHED_HEADERS.map((header) => (
          <div
            key={header}
            className="px-3 py-2 text-[10px] font-black uppercase tracking-widest text-primary/60"
          >
            {header}
          </div>
        ))}
      </div>

      {MOCK_ROWS.map((row, i) => {
        const done = i < enrichedCount;
        const active = i === enrichedCount && enrichedCount < total;

        return (
          <div
            key={row.company}
            className={`grid grid-cols-5 border-b border-slate-100 last:border-0 transition-colors duration-500 ${active ? "bg-primary/5" : "bg-white"}`}
          >
            <div className="px-3 py-2.5 text-xs font-semibold text-slate-900 truncate">
              {row.company}
            </div>
            <div className="px-3 py-2.5 text-xs text-slate-500 truncate">
              {row.website}
            </div>

            {done ? (
              <>
                <div className="px-3 py-2.5 flex items-center gap-1.5">
                  <span className="text-xs text-slate-900 truncate">{row.ceo}</span>
                  <span className="text-[9px] font-black bg-emerald-100 text-emerald-700 px-1 rounded shrink-0">
                    97%
                  </span>
                </div>
                <div className="px-3 py-2.5 flex items-center gap-1.5">
                  <span className="text-xs text-slate-900">{row.founded}</span>
                  <span className="text-[9px] font-black bg-emerald-100 text-emerald-700 px-1 rounded shrink-0">
                    99%
                  </span>
                </div>
                <div className="px-3 py-2.5 flex items-center gap-1.5">
                  <span className="text-xs text-slate-900">{row.revenue}</span>
                  <span className="text-[9px] font-black bg-yellow-100 text-yellow-700 px-1 rounded shrink-0">
                    84%
                  </span>
                </div>
              </>
            ) : (
              <>
                <div className="px-3 py-2.5">
                  <SkeletonCell active={active} />
                </div>
                <div className="px-3 py-2.5">
                  <SkeletonCell active={active} />
                </div>
                <div className="px-3 py-2.5">
                  <SkeletonCell active={active} />
                </div>
              </>
            )}
          </div>
        );
      })}

      <div className="px-4 py-3 bg-slate-50 border-t border-slate-100">
        <div className="flex items-center justify-between text-[10px] font-bold text-slate-500 mb-1.5">
          <span>Enrichment progress</span>
          <span>
            {enrichedCount}/{total} rows
          </span>
        </div>
        <div className="h-1.5 bg-slate-200 rounded-full overflow-hidden">
          <div
            className="h-full bg-primary rounded-full transition-all duration-700 ease-out"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>
    </div>
  );
}
