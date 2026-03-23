import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import {
  ArrowRight,
  Bot,
  CheckCircle,
  FileText,
  Search,
  Sparkles,
} from "lucide-react";
import { useEffect, useState } from "react";
import logo from "../../assets/ampledata-logo.png";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/")({
  beforeLoad: ({ context }) => {
    if (!context.auth.isLoading && context.auth.user) {
      throw redirect({ to: "/app" });
    }
  },
  component: LandingPage,
});

// ─── Mock enrichment data ─────────────────────────────────────────────────────

const MOCK_ROWS = [
  {
    company: "Acme Corp",
    website: "acme.com",
    ceo: "Jane Smith",
    founded: "2005",
    revenue: "$4.2M",
  },
  {
    company: "TechStart",
    website: "techstart.io",
    ceo: "David Lee",
    founded: "2019",
    revenue: "$1.8M",
  },
  {
    company: "DataFlow AI",
    website: "dataflow.ai",
    ceo: "Sara Kim",
    founded: "2015",
    revenue: "$12M",
  },
  {
    company: "CloudBase",
    website: "cloudbase.co",
    ceo: "Tom Allen",
    founded: "2018",
    revenue: "$6.5M",
  },
];

// ─── Hero mock component ──────────────────────────────────────────────────────

function SkeletonCell({ active }: { active?: boolean }) {
  return (
    <div
      className={`h-4 rounded w-20 ${active ? "bg-primary/20 animate-pulse" : "bg-slate-200 animate-pulse"}`}
    />
  );
}

function EnrichmentMockHero() {
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
      {/* Browser chrome */}
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

      {/* Table header */}
      <div className="grid grid-cols-5 bg-slate-50/60 border-b border-slate-100">
        {["Company", "Website"].map((h) => (
          <div
            key={h}
            className="px-3 py-2 text-[10px] font-black uppercase tracking-widest text-slate-400"
          >
            {h}
          </div>
        ))}
        {["CEO ✦", "Founded ✦", "Revenue ✦"].map((h) => (
          <div
            key={h}
            className="px-3 py-2 text-[10px] font-black uppercase tracking-widest text-primary/60"
          >
            {h}
          </div>
        ))}
      </div>

      {/* Rows */}
      {MOCK_ROWS.map((row, i) => {
        const done = i < enrichedCount;
        const active = i === enrichedCount && enrichedCount < total;
        return (
          <div
            key={row.company}
            className={`grid grid-cols-5 border-b border-slate-100 last:border-0 transition-colors duration-500 ${active ? "bg-primary/5" : "bg-white"}`}
          >
            {/* Source columns — always shown */}
            <div className="px-3 py-2.5 text-xs font-semibold text-slate-900 truncate">
              {row.company}
            </div>
            <div className="px-3 py-2.5 text-xs text-slate-500 truncate">
              {row.website}
            </div>

            {/* Enriched columns */}
            {done ? (
              <>
                <div className="px-3 py-2.5 flex items-center gap-1.5">
                  <span className="text-xs text-slate-900 truncate">
                    {row.ceo}
                  </span>
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

      {/* Progress bar */}
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

// ─── Enrichment pipeline graph ────────────────────────────────────────────────

const PIPELINE_NODES = [
  {
    icon: FileText,
    label: "Upload file",
    desc: "Drop in CSV or JSON",
    color: "bg-slate-100 text-slate-700 border-slate-200",
  },
  {
    icon: Search,
    label: "Find sources",
    desc: "We search the web for each row",
    color: "bg-blue-50 text-blue-700 border-blue-200",
  },
  {
    icon: Bot,
    label: "Extract answers",
    desc: "AI fills in the missing fields",
    color: "bg-violet-50 text-violet-700 border-violet-200",
  },
  {
    icon: CheckCircle,
    label: "Quality check",
    desc: "Confidence scores help you verify",
    color: "bg-emerald-50 text-emerald-700 border-emerald-200",
  },
  {
    icon: Sparkles,
    label: "Ready to use",
    desc: "Export enriched rows instantly",
    color: "bg-primary/10 text-primary border-primary/30",
  },
] as const;

const PIPELINE_ANIMATION_INTERVAL_MS = 1100;

function EnrichmentPipelineGraph() {
  const [activeIdx, setActiveIdx] = useState(0);

  useEffect(() => {
    const id = setInterval(
      () => setActiveIdx((prevIndex) => (prevIndex + 1) % PIPELINE_NODES.length),
      PIPELINE_ANIMATION_INTERVAL_MS,
    );
    return () => clearInterval(id);
  }, []);

  return (
    <div className="w-full max-w-[90rem] mx-auto flex flex-col lg:flex-row items-center justify-center gap-4 lg:gap-2">
      {PIPELINE_NODES.map((node, i) => {
        const Icon = node.icon;
        const isActive = i === activeIdx;
        const isPast = i < activeIdx;
        
        return [
          <div key={`${node.label}-node`} className="flex flex-col items-center w-full lg:w-[220px] xl:w-[240px] shrink-0">
            <div
              className={`w-full max-w-[320px] lg:max-w-none rounded-3xl border-2 p-8 lg:p-6 bg-card transition-all duration-500 z-10 relative ${
                isActive
                  ? "border-primary shadow-2xl scale-105 ring-4 ring-primary/20 lg:-translate-y-2"
                  : isPast
                  ? "border-primary/50 shadow-md opacity-90"
                  : "border-border shadow-sm opacity-60 grayscale-[0.2]"
              }`}
            >
              <div 
                className={`mx-auto w-20 h-20 lg:w-16 lg:h-16 rounded-2xl flex items-center justify-center transition-all duration-500 ${node.color} ${
                  isActive ? "shadow-lg scale-110" : ""
                }`}
              >
                <Icon className={`size-10 lg:size-8 transition-transform duration-500 ${isActive ? "scale-110" : ""}`} />
              </div>
              <div 
                className={`mt-5 lg:mt-4 text-center text-xl lg:text-lg font-black leading-tight transition-colors duration-300 ${
                  isActive ? "text-primary" : "text-foreground"
                }`}
              >
                {node.label}
              </div>
              <p className="mt-3 lg:mt-2 text-base lg:text-sm text-center text-muted-foreground leading-relaxed">
                {node.desc}
              </p>
            </div>
          </div>,
          i < PIPELINE_NODES.length - 1 && (
            <div key={`${node.label}-connector`} className="flex items-center justify-center py-2 lg:py-0 lg:w-6 xl:w-10 shrink-0 z-0">
              <ArrowRight 
                className={`size-10 lg:size-8 transition-all duration-500 rotate-90 lg:rotate-0 ${
                  isPast ? "text-primary opacity-100" : isActive ? "text-primary/70 animate-pulse translate-y-2 lg:translate-y-0 lg:translate-x-2" : "text-muted opacity-30"
                }`}
              />
            </div>
          )
        ];
      })}
    </div>
  );
}

// ─── Landing page ─────────────────────────────────────────────────────────────

function LandingPage() {
  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Navigation */}
      <nav className="sticky top-0 z-10 w-full border-b bg-background/95 backdrop-blur border-border shadow-sm">
        <div className="container mx-auto px-4 h-20 flex items-center justify-between">
          <Link to="/">
            <img src={logo} alt="AmpleData" className="h-12 w-auto" />
          </Link>
          <div className="flex items-center gap-3">
            <Button variant="ghost" asChild>
              <Link to="/login">Sign in</Link>
            </Button>
            <Button asChild>
              <Link to="/login">
                Get started <ArrowRight className="ml-2 size-4" />
              </Link>
            </Button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative overflow-hidden bg-background">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute -top-24 -right-24 w-[600px] h-[600px] rounded-full bg-primary/5 blur-3xl" />
          <div className="absolute -bottom-24 -left-24 w-[400px] h-[400px] rounded-full bg-primary/5 blur-3xl" />
        </div>
        <div className="container mx-auto px-4 py-24 md:py-32 flex flex-col items-center text-center relative">
          <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-4 py-1.5 text-sm font-medium mb-6">
            <Sparkles className="size-4" />
            <span>AI-powered dataset enrichment</span>
          </div>
          <h1 className="text-4xl md:text-6xl font-black tracking-tight text-foreground max-w-3xl mb-6">
            Enrich your datasets{" "}
            <span className="text-primary">
              without writing a single line of code
            </span>
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mb-10">
            Upload a CSV or JSON file, define what information you need, and let
            AmpleData automatically search the web and extract the data for you.
            No coding required.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
            <Button size="lg" asChild className="text-base px-8">
              <Link to="/login">
                Start for free <ArrowRight className="ml-2 size-5" />
              </Link>
            </Button>
          </div>
          {/* Animated mock enrichment component */}
          <div className="w-full max-w-3xl">
            <EnrichmentMockHero />
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="bg-secondary/30 py-24">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
              How it works
            </h2>
            <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
              Get from raw dataset to enriched data in three simple steps.
            </p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {steps.map((step, index) => (
              <div
                key={step.title}
                className="flex flex-col items-center text-center gap-4"
              >
                <div className="w-14 h-14 rounded-full bg-primary flex items-center justify-center text-primary-foreground font-black text-xl shadow-md">
                  {index + 1}
                </div>
                <h3 className="text-xl font-bold text-foreground">
                  {step.title}
                </h3>
                <p className="text-muted-foreground leading-relaxed">
                  {step.description}
                </p>
                <div className="w-full rounded-xl overflow-hidden border border-border shadow-sm mt-2">
                  <img
                    src={step.image}
                    alt={step.title}
                    className="w-full h-48 object-cover"
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Enrichment Pipeline Section (formerly Features) */}
      <section className="bg-background py-24">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
              Your data journey
            </h2>
            <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
              A simple visual flow that shows how AmpleData turns raw rows into
              trusted, ready-to-use data.
            </p>
          </div>
          <EnrichmentPipelineGraph />
        </div>
      </section>

      {/* Benefits Section */}
      <section className="bg-primary/5 py-24">
        <div className="container mx-auto px-4">
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-6">
                Built for data teams who move fast
              </h2>
              <ul className="space-y-4">
                {benefits.map((benefit) => (
                  <li key={benefit} className="flex items-start gap-3">
                    <CheckCircle className="size-5 text-primary mt-0.5 shrink-0" />
                    <span className="text-muted-foreground">{benefit}</span>
                  </li>
                ))}
              </ul>
            </div>
            <div className="rounded-2xl overflow-hidden shadow-xl border border-border">
              <img
                src="https://images.unsplash.com/photo-1460925895917-afdab827c52f?auto=format&fit=crop&w=800&q=80"
                alt="Data analytics"
                className="w-full h-auto object-cover"
              />
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-primary py-24">
        <div className="container mx-auto px-4 text-center">
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-primary-foreground mb-4">
            Ready to enrich your data?
          </h2>
          <p className="text-primary-foreground/80 text-lg max-w-xl mx-auto mb-10">
            Join teams using AmpleData to automate their data enrichment
            workflows and save hours of manual research.
          </p>
          <Button
            size="lg"
            variant="secondary"
            asChild
            className="text-base px-10"
          >
            <Link to="/login">
              Get started for free <ArrowRight className="ml-2 size-5" />
            </Link>
          </Button>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-background border-t border-border py-8">
        <div className="container mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4">
          <img src={logo} alt="AmpleData" className="h-8 w-auto" />
          <p className="text-muted-foreground text-sm">
            © {new Date().getFullYear()} AmpleData. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}

const steps = [
  {
    title: "Upload your dataset",
    description:
      "Drag and drop a CSV or JSON file into AmpleData. Your data is securely stored and ready for enrichment.",
    image:
      "https://images.unsplash.com/photo-1504868584819-f8e8b4b6d7e3?auto=format&fit=crop&w=600&q=80",
  },
  {
    title: "Define enrichment fields",
    description:
      "Tell AmpleData which columns to use as context and what new information you want to extract for each row.",
    image:
      "https://images.unsplash.com/photo-1432888498266-38ffec3eaf0a?auto=format&fit=crop&w=600&q=80",
  },
  {
    title: "Get enriched data",
    description:
      "AmpleData runs the enrichment jobs and populates your new columns automatically — with sources and confidence scores.",
    image:
      "https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=600&q=80",
  },
];

const benefits = [
  "No coding or scripting required — anyone can enrich a dataset",
  "Works with CSV or JSON files of various sizes",
  "Transparent confidence scores so you can trust the output",
  "Source URLs for every enriched value for easy verification",
  "Run enrichments on demand or schedule them automatically",
];
