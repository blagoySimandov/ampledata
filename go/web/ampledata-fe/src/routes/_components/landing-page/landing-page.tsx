import { Link } from "@tanstack/react-router";
import { ArrowRight, CheckCircle, Sparkles } from "lucide-react";
import logo from "../../../../assets/ampledata-logo.png";
import { Button } from "@/components/ui/button";
import { BENEFITS, STEPS } from "./constants";
import { ContactSection } from "./contact-section";
import { EnrichmentMockHero } from "./enrichment-mock-hero";
import { EnrichmentPipelineGraph } from "./enrichment-pipeline-graph";
import { PricingSection } from "./pricing-section";

export function LandingPage() {
  return (
    <div className="min-h-screen bg-background flex flex-col">
      <nav className="sticky top-0 z-10 w-full border-b bg-background/95 backdrop-blur border-border shadow-sm">
        <div className="container mx-auto px-4 h-20 flex items-center justify-between">
          <Link to="/">
            <img src={logo} alt="AmpleData" className="h-12 w-auto" />
          </Link>
          <div className="flex items-center gap-3">
            <Button
              variant="ghost"
              onClick={() =>
                document
                  .getElementById("pricing")
                  ?.scrollIntoView({ behavior: "smooth" })
              }
            >
              Pricing
            </Button>
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
          <div className="w-full max-w-3xl">
            <EnrichmentMockHero />
          </div>
        </div>
      </section>

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
            {STEPS.map((step, index) => (
              <div
                key={step.title}
                className="flex flex-col items-center text-center gap-4"
              >
                <div className="w-14 h-14 rounded-full bg-primary flex items-center justify-center text-primary-foreground font-black text-xl shadow-md">
                  {index + 1}
                </div>
                <h3 className="text-xl font-bold text-foreground">{step.title}</h3>
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

      <section className="bg-primary/5 py-24">
        <div className="container mx-auto px-4">
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-6">
                Built for data teams who move fast
              </h2>
              <ul className="space-y-4">
                {BENEFITS.map((benefit) => (
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

      <PricingSection />

      <ContactSection />

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
