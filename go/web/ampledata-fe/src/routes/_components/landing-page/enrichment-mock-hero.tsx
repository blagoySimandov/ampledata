import { useState, useEffect } from "react";
import { HERO_ROWS, HERO_COLS, HERO_CONFIDENCES } from "./constants";

type HeroRow = (typeof HERO_ROWS)[number];
const GRID_STYLE = { gridTemplateColumns: "110px 90px 1fr 90px 90px" };
const TOTAL = HERO_ROWS.length;

function MacDots() {
  return (
    <div className="flex gap-1.5">
      {(["bg-red-400", "bg-yellow-400", "bg-green-400"] as const).map((c) => (
        <div key={c} className={`w-2.5 h-2.5 rounded-full ${c}`} />
      ))}
    </div>
  );
}

function TitleBar({ step }: { step: number }) {
  const done = step >= TOTAL;
  return (
    <div className="flex items-center gap-2 px-3 py-2 bg-secondary border-b border-border">
      <MacDots />
      <div className="flex-1 h-5 bg-card border border-border rounded text-[10px] text-muted-foreground font-semibold flex items-center px-2.5 min-w-0">
        AmpleData — yc-w26-companies.csv
      </div>
      {done ? (
        <span className="text-[9px] font-black text-emerald-700 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-full shrink-0">
          Complete
        </span>
      ) : (
        <span className="text-[9px] font-black text-blue-700 bg-blue-50 border border-blue-200 px-2 py-0.5 rounded-full animate-pulse shrink-0">
          Enriching...
        </span>
      )}
    </div>
  );
}

function ColumnHeaders() {
  return (
    <div className="grid bg-secondary border-b border-border" style={GRID_STYLE}>
      {HERO_COLS.map((col, i) => (
        <div
          key={col}
          className={`px-2.5 py-1.5 text-[8.5px] font-black uppercase tracking-wider truncate ${i >= 2 ? "text-primary/70" : "text-muted-foreground"}`}
        >
          {i >= 2 ? `${col} ✶` : col}
        </div>
      ))}
    </div>
  );
}

function ConfidenceBadge({ value }: { value: number }) {
  return (
    <span className="shrink-0 text-[8px] font-black px-1 py-0.5 rounded bg-emerald-100 text-emerald-700 mt-0.5">
      {value}%
    </span>
  );
}

function LoadingCells({ active }: { active: boolean }) {
  return (
    <>
      {[80, 50, 55].map((w, j) => (
        <div key={j} className="px-2.5">
          <div
            className={`h-2.5 rounded-sm ${active ? "bg-primary/15 animate-pulse" : "bg-secondary"}`}
            style={{ width: w }}
          />
        </div>
      ))}
    </>
  );
}

function EnrichedCells({ row, i }: { row: HeroRow; i: number }) {
  return (
    <>
      <div className="px-2.5 flex items-center gap-1.5 animate-cell-pop min-w-0">
        <span className="text-[11px] text-foreground truncate min-w-0 flex-1">{row.founderStmt}</span>
        <ConfidenceBadge value={HERO_CONFIDENCES[i]} />
      </div>
      <div className="px-2.5 text-[11px] text-muted-foreground">{row.retracted}</div>
      <div className="px-2.5 flex items-center gap-1">
        <span className="text-[11px] text-foreground">{row.remote}</span>
        <ConfidenceBadge value={95} />
      </div>
    </>
  );
}

function CitationDrawer({ row, i }: { row: HeroRow; i: number }) {
  return (
    <div className="px-3.5 py-2.5 bg-primary/5 border-t border-primary/15 animate-fade-up">
      <div className="text-[9px] font-black uppercase tracking-widest text-primary mb-1.5">
        Source citation
      </div>
      <div className="text-[10px] text-foreground mb-1 leading-snug">
        {row.founderStmt.replace(/"/g, "")}
      </div>
      <div className="text-[9px] text-blue-600 underline">
        techcrunch.com/2025/founder-interview-{row.name.toLowerCase()}
      </div>
      <div className="text-[9px] text-muted-foreground mt-1">
        Confidence {HERO_CONFIDENCES[i]}% · extracted from paragraph 3
      </div>
    </div>
  );
}

function HeroRow({
  row,
  i,
  step,
  citationRow,
  onToggle,
}: {
  row: HeroRow;
  i: number;
  step: number;
  citationRow: number | null;
  onToggle: (i: number) => void;
}) {
  const done = i < step;
  const active = i === step && step < TOTAL;
  return (
    <div
      className={`border-b border-border last:border-0 ${done ? "cursor-pointer" : ""} ${active ? "bg-primary/[0.03]" : citationRow === i ? "bg-primary/5" : "bg-card"} transition-colors duration-300`}
      onClick={() => done && onToggle(i)}
    >
      <div className="grid h-10 items-center" style={GRID_STYLE}>
        <div className="px-2.5 text-[11px] font-bold text-foreground">{row.name}</div>
        <div className="px-2.5 text-[11px] text-muted-foreground">{row.industry}</div>
        {done ? <EnrichedCells row={row} i={i} /> : <LoadingCells active={active} />}
      </div>
      {citationRow === i && <CitationDrawer row={row} i={i} />}
    </div>
  );
}

function ProgressBar({ step }: { step: number }) {
  const pct = Math.round((Math.min(step, TOTAL) / TOTAL) * 100);
  return (
    <div className="px-3 py-2 bg-secondary border-t border-border">
      <div className="flex justify-between text-[9px] font-bold text-muted-foreground mb-1.5">
        <span>Enrichment progress</span>
        <span>
          {Math.min(step, TOTAL)}/{TOTAL} rows · Est. cost{" "}
          <strong className="text-foreground">$0.03</strong>
        </span>
      </div>
      <div className="h-1 bg-border rounded-full overflow-hidden">
        <div
          className="h-full bg-primary rounded-full transition-all duration-700 ease-out"
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

export function EnrichmentMockHero() {
  const [step, setStep] = useState(0);
  const [citationRow, setCitationRow] = useState<number | null>(null);

  useEffect(() => {
    const delay = step < TOTAL ? 1200 : 3500;
    const t = setTimeout(() => {
      if (step < TOTAL) {
        setStep((s) => s + 1);
      } else {
        setStep(0);
        setCitationRow(null);
      }
    }, delay);
    return () => clearTimeout(t);
  }, [step]);

  const toggle = (i: number) => setCitationRow((prev) => (prev === i ? null : i));

  return (
    <div className="rounded-xl border border-border shadow-2xl overflow-x-auto text-left w-full">
      <div className="min-w-[480px]">
        <TitleBar step={step} />
        <ColumnHeaders />
        {HERO_ROWS.map((row, i) => (
          <HeroRow key={row.name} row={row} i={i} step={step} citationRow={citationRow} onToggle={toggle} />
        ))}
        <ProgressBar step={step} />
      </div>
    </div>
  );
}
