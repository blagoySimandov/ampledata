import { useEffect, useState } from "react";
import { Link } from "@tanstack/react-router";
import logo from "../../../../assets/ampledata-logo.png";
import { Button } from "@/components/ui/button";
import { ContactFormWidget } from "@/components/widgets";
import { EnrichmentMockHero } from "./enrichment-mock-hero";
import { PricingSection } from "./pricing-section";
import { FaqSection } from "./faq-section";
import { LANDING_FEATURES, COMPARISONS } from "./constants";

function NavBar() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 12);
    window.addEventListener("scroll", handler, { passive: true });
    return () => window.removeEventListener("scroll", handler);
  }, []);

  return (
    <nav
      className={`sticky top-0 z-50 transition-all duration-250 backdrop-blur-md ${scrolled ? "bg-background/95 border-b border-border" : "bg-background/80 border-b border-transparent"}`}
    >
      <div className="max-w-[1160px] mx-auto px-4 md:px-6 flex items-center justify-between h-[clamp(60px,6vw,84px)]">
        <Link to="/">
          <img src={logo} alt="AmpleData" className="h-[clamp(34px,3vw,44px)] w-auto" />
        </Link>
        <div className="flex items-center gap-2">
          <button
            onClick={() => document.getElementById("pricing")?.scrollIntoView({ behavior: "smooth" })}
            className="px-3 py-1.5 text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors rounded-md"
          >
            Pricing
          </button>
          <Button asChild className="h-auto px-5 py-2.5 text-sm font-semibold rounded-lg">
            <Link to="/login">Try it out now &rarr;</Link>
          </Button>
        </div>
      </div>
    </nav>
  );
}

function HeroSection() {
  return (
    <section className="relative overflow-hidden bg-background text-center py-14 md:py-24 px-4 md:px-6">
      <div className="absolute -top-28 -right-16 w-[560px] h-[560px] rounded-full bg-primary/6 blur-[80px] pointer-events-none animate-[float_13s_ease-in-out_infinite]" />
      <div className="absolute -bottom-20 -left-20 w-[400px] h-[400px] rounded-full bg-primary/5 blur-[70px] pointer-events-none animate-[float-reverse_10s_ease-in-out_infinite]" />
      <div className="max-w-[860px] mx-auto relative">
        <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-4 py-1.5 text-xs font-bold mb-6">
          ✶ AI research worker for your spreadsheet
        </div>
        <h1 className="text-[clamp(36px,6vw,72px)] font-black tracking-[-0.03em] text-foreground mb-5 leading-[1.08] text-balance">
          Stop pasting rows
          <br />
          into ChatGPT.
        </h1>
        <p className="text-[clamp(15px,2vw,22px)] text-muted-foreground max-w-[680px] mx-auto mb-3 leading-relaxed text-pretty">
          AmpleData researches every row of your spreadsheet, with{" "}
          <strong className="text-foreground font-bold">citations</strong>,{" "}
          <strong className="text-foreground font-bold">confidence scores</strong>, and{" "}
          <strong className="text-foreground font-bold">conflict resolution</strong> across sources.
        </p>
        <div className="flex gap-3 justify-center flex-wrap mt-8 mb-10 md:mb-14">
          {/* TODO: replace with sample list demo once built, for now redirect to login */}
          <Button size="lg" asChild className="h-auto px-7 py-3.5 text-base font-bold rounded-lg shadow-[0_4px_14px_oklch(0.553_0.195_38.402/0.3)]">
            <Link to="/login">Try it out now &rarr;</Link>
          </Button>
          <Button
            size="lg"
            variant="outline"
            className="h-auto px-6 py-3.5 text-base font-semibold rounded-lg"
            onClick={() => document.getElementById("how")?.scrollIntoView({ behavior: "smooth" })}
          >
            See how it works
          </Button>
        </div>
        <div className="max-w-[740px] mx-auto min-w-0 w-full">
          <EnrichmentMockHero />
          <p className="text-[11px] text-muted-foreground mt-2.5 font-semibold">
            Click any enriched cell to see its source citation
          </p>
        </div>
      </div>
    </section>
  );
}

function WhatItDoesSection() {
  return (
    <section id="how" className="py-14 md:py-20 px-4 md:px-6 bg-secondary/40 border-t border-border">
      <div className="max-w-[860px] mx-auto">
        <div className="inline-flex items-center gap-2 bg-primary/8 text-primary rounded-full px-3 py-1 text-[11px] font-black uppercase tracking-widest mb-5">
          What it does
        </div>
        <p className="text-[clamp(17px,2vw,21px)] text-foreground leading-[1.7] text-pretty">
          Give AmpleData a list, companies, products, papers, jobs, properties, anything with a name.
          Describe the columns you want filled in, in plain English:{" "}
          <em className="text-primary">"the founder's most recent public statement on pricing,"</em>{" "}
          <em className="text-primary">"whether this paper has been retracted,"</em>{" "}
          <em className="text-primary">"the company's stated stance on remote work."</em> AmpleData
          dispatches web search and crawl, extracts structured answers with an LLM, resolves conflicts
          across sources, and returns a cleaned-up sheet you can trust, every cell linked back to where
          the answer came from.
        </p>
      </div>
    </section>
  );
}

function DemoFeatureBullet({ title, desc }: { title: string; desc: string }) {
  return (
    <div className="flex gap-3 p-4 bg-secondary rounded-xl border border-border">
      <span className="text-primary text-base shrink-0 mt-0.5">✓</span>
      <div>
        <div className="text-sm font-bold text-foreground mb-0.5">{title}</div>
        <div className="text-sm text-muted-foreground leading-relaxed">{desc}</div>
      </div>
    </div>
  );
}

function DemoSection() {
  return (
    <section id="demo" className="py-14 md:py-20 px-4 md:px-6 bg-background">
      <div className="max-w-[1280px] mx-auto grid md:grid-cols-2 gap-10 md:gap-16 items-center">
        <div>
          <div className="inline-flex items-center gap-2 bg-primary/8 text-primary rounded-full px-3 py-1 text-[11px] font-black uppercase tracking-widest mb-5">
            Live enrichment
          </div>
          <h2 className="text-[clamp(26px,4vw,38px)] font-black tracking-tight text-foreground mb-3.5 leading-[1.15]">
            Watch cells fill in, live, in your browser
          </h2>
          <p className="text-base text-muted-foreground leading-relaxed mb-6">
            Each row gets dispatched to the web, sources get crawled, an LLM extracts the answer,
            conflicts are resolved, and the result lands in the cell, with a source URL and confidence
            score attached.
          </p>
          <div className="flex flex-col gap-2.5">
            <DemoFeatureBullet
              title="Every cell is cited"
              desc="Click any cell to see the source URL, extracted snippet, and reasoning."
            />
            <DemoFeatureBullet
              title="Confidence you can act on"
              desc="Sort by confidence. Spot-check weak cells. Re-run with a refined prompt."
            />
            <DemoFeatureBullet
              title="Per-cell pricing"
              desc="Pay only for cells you enrich. No seat licenses, no minimums, no annual contracts."
            />
          </div>
        </div>
        <div className="min-w-0">
          <EnrichmentMockHero />
        </div>
      </div>
    </section>
  );
}

function FeaturesSection() {
  return (
    <section className="py-14 md:py-20 px-4 md:px-6 bg-primary/[0.03] border-t border-border">
      <div className="max-w-[1280px] mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-[clamp(26px,4vw,38px)] font-black tracking-tight text-foreground mb-2.5">
            Three things that matter
          </h2>
          <p className="text-base text-muted-foreground">Not the longest feature list. The most distinctive one.</p>
        </div>
        <div className="grid md:grid-cols-3 gap-5">
          {LANDING_FEATURES.map((f) => (
            <div key={f.num} className="bg-card border border-border rounded-xl p-7 flex flex-col gap-3.5">
              <div className="text-[11px] font-black tracking-widest uppercase text-primary">{f.num}</div>
              <h3 className="text-xl font-extrabold text-foreground leading-snug">{f.title}</h3>
              <p className="text-base text-muted-foreground leading-relaxed">{f.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function ComparisonSection() {
  return (
    <section className="py-14 md:py-20 px-4 md:px-6 bg-background border-t border-border">
      <div className="max-w-[1080px] mx-auto">
        <div className="text-center mb-11">
          <div className="inline-flex items-center gap-2 bg-primary/8 text-primary rounded-full px-3 py-1 text-[11px] font-black uppercase tracking-widest mb-4">
            Comparisons
          </div>
          <h2 className="text-[clamp(24px,3.5vw,34px)] font-black tracking-tight text-foreground">
            How does it compare?
          </h2>
        </div>
        <div className="grid md:grid-cols-2 gap-4">
          {COMPARISONS.map((c) => (
            <div key={c.label} className="bg-secondary border border-border rounded-xl p-6">
              <div className="text-[11px] font-black uppercase tracking-widest text-primary mb-2.5">{c.label}</div>
              <p className="text-base text-muted-foreground leading-relaxed">{c.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function CtaSection() {
  return (
    <section className="bg-primary py-14 md:py-20 px-4 md:px-6 text-center">
      <p className="text-xs font-black uppercase tracking-widest text-primary-foreground/60 mb-3.5">
        AmpleData
      </p>
      <h2 className="text-[clamp(24px,4vw,40px)] font-black text-primary-foreground mb-3.5 tracking-tight text-balance">
        Built for people who do research
        <br />
        one row at a time.
      </h2>
      <p className="text-base text-primary-foreground/75 mb-8 leading-relaxed max-w-lg mx-auto">
        Upload a list. Define your columns. Get back cited, confidence-scored data in minutes.
      </p>
      {/* TODO: replace with sample list demo once built, for now redirect to login */}
      <Button size="lg" variant="secondary" className="h-auto px-8 py-4 text-base font-bold rounded-lg shadow-[0_4px_14px_rgba(0,0,0,0.15)]" asChild>
        <Link to="/login">Try it out now &rarr;</Link>
      </Button>
    </section>
  );
}

function LandingFooter() {
  return (
    <footer className="bg-background border-t border-border py-6 px-4 md:px-6">
      <div className="max-w-[1080px] mx-auto flex items-center justify-between flex-wrap gap-3">
        <img src={logo} alt="AmpleData" className="h-7 w-auto" />
        <p className="text-xs text-muted-foreground">
          Made by a person who got tired of pasting rows into ChatGPT.
        </p>
        <div className="flex gap-4 text-xs text-muted-foreground">
          <Link to="/privacy-policy" className="hover:text-foreground transition-colors">
            Privacy
          </Link>
          <Link to="/terms" className="hover:text-foreground transition-colors">
            Terms
          </Link>
        </div>
      </div>
    </footer>
  );
}

export function LandingPage() {
  return (
    <div className="landing-page min-h-screen bg-background flex flex-col">
      <NavBar />
      <HeroSection />
      <WhatItDoesSection />
      <DemoSection />
      <FeaturesSection />
      <ComparisonSection />
      <PricingSection />
      <FaqSection />
      <ContactFormWidget variant="landing" />
      <CtaSection />
      <LandingFooter />
    </div>
  );
}
