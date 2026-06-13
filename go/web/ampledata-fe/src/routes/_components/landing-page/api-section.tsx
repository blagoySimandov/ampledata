import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { API_FEATURES, API_CODE_SAMPLE } from "./constants";

function ApiFeature({ title, body }: { title: string; body: string }) {
  return (
    <div className="flex gap-3">
      <span className="text-primary text-base shrink-0 mt-0.5">✓</span>
      <div>
        <div className="text-sm font-bold text-foreground mb-0.5">{title}</div>
        <div className="text-sm text-muted-foreground leading-relaxed">{body}</div>
      </div>
    </div>
  );
}

function TerminalDot({ className }: { className: string }) {
  return <span className={`w-3 h-3 rounded-full ${className}`} />;
}

function highlightTokens(text: string) {
  const parts = text.split(/(\s-{1,2}[A-Za-z][\w-]*|\b(?:curl|export|echo)\b)/g);
  return parts.map((p, i) => {
    if (/^\s-{1,2}[A-Za-z]/.test(p)) return <span key={i} className="text-primary/70 font-semibold">{p}</span>;
    if (/^(?:curl|export|echo)$/.test(p)) return <span key={i} className="font-bold text-foreground">{p}</span>;
    return p;
  });
}

function highlightSegment(text: string) {
  const parts = text.split(/("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')/g);
  return parts.map((p, i) =>
    /^["']/.test(p) ? <span key={i} className="text-primary">{p}</span> : <span key={i}>{highlightTokens(p)}</span>,
  );
}

function highlightLine(line: string, key: number, newline: boolean) {
  const tail = newline ? "\n" : "";
  if (line.trimStart().startsWith("#")) {
    return <span key={key} className="text-muted-foreground italic">{line + tail}</span>;
  }
  return <span key={key}>{highlightSegment(line)}{tail}</span>;
}

function HighlightedCode({ source }: { source: string }) {
  const lines = source.split("\n");
  return <code>{lines.map((l, i) => highlightLine(l, i, i < lines.length - 1))}</code>;
}

function CodeCard() {
  return (
    <div className="min-w-0 rounded-xl border border-border bg-card overflow-hidden shadow-[0_8px_30px_oklch(0_0_0/0.08)]">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-secondary/50">
        <TerminalDot className="bg-primary/40" />
        <TerminalDot className="bg-primary/25" />
        <TerminalDot className="bg-primary/15" />
        <span className="ml-2 text-[11px] font-bold uppercase tracking-widest text-muted-foreground">
          enrich.sh
        </span>
      </div>
      <pre className="px-5 py-4 overflow-x-auto text-[12.5px] leading-relaxed text-foreground/80 font-mono">
        <HighlightedCode source={API_CODE_SAMPLE} />
      </pre>
    </div>
  );
}

function ApiCopy() {
  return (
    <div>
      <div className="inline-flex items-center gap-2 bg-primary/8 text-primary rounded-full px-3 py-1 text-[11px] font-black uppercase tracking-widest mb-5">
        For developers
      </div>
      <h2 className="text-[clamp(26px,4vw,38px)] font-black tracking-tight text-foreground mb-3.5 leading-[1.15]">
        Enrich straight from your own code
      </h2>
      <p className="text-base text-muted-foreground leading-relaxed mb-6">
        Generate an API key and drive the same enrichment engine over HTTP, no UI in the loop. Wire it
        into a cron job, a pipeline, or your own product.
      </p>
      <div className="flex flex-col gap-4 mb-8">
        {API_FEATURES.map((f) => (
          <ApiFeature key={f.title} title={f.title} body={f.body} />
        ))}
      </div>
      <div className="flex items-center gap-4 flex-wrap">
        <Button asChild className="h-auto px-6 py-3 text-base font-bold rounded-lg">
          <Link to="/login">Get your API key &rarr;</Link>
        </Button>
        <Link to="/docs" className="text-sm font-bold text-primary hover:underline">
          Read the API docs &rarr;
        </Link>
      </div>
    </div>
  );
}

export function ApiSection() {
  return (
    <section id="api" className="py-14 md:py-20 px-4 md:px-6 bg-secondary/40 border-t border-border">
      <div className="max-w-[1280px] mx-auto grid md:grid-cols-2 gap-10 md:gap-16 items-center">
        <ApiCopy />
        <CodeCard />
      </div>
    </section>
  );
}
