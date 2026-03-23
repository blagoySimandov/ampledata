import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { ArrowRight, CheckCircle, Database, Search, Sparkles, Zap } from "lucide-react";
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
            <span className="text-primary">without writing a single line of code</span>
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mb-10">
            Upload a CSV or JSON file, define what information you need, and let AmpleData automatically search the web and extract the data for you. No coding required.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" asChild className="text-base px-8">
              <Link to="/login">
                Start for free <ArrowRight className="ml-2 size-5" />
              </Link>
            </Button>
          </div>
          {/* Hero image placeholder */}
          <div className="mt-16 w-full max-w-5xl rounded-2xl overflow-hidden shadow-2xl border border-border">
            <img
              src="https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=1200&q=80"
              alt="Data enrichment dashboard"
              className="w-full h-auto object-cover"
            />
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="bg-secondary/30 py-24">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
              Everything you need to enrich your data
            </h2>
            <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
              AmpleData combines web search, AI extraction, and structured
              output to fill in the blanks in your datasets automatically.
            </p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="bg-card rounded-2xl p-8 border border-border shadow-sm flex flex-col gap-4"
              >
                <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
                  <feature.icon className="size-6 text-primary" />
                </div>
                <h3 className="text-xl font-bold text-foreground">
                  {feature.title}
                </h3>
                <p className="text-muted-foreground leading-relaxed">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="bg-background py-24">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
              How it works
            </h2>
            <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
              Get from raw dataset to enriched data in three simple steps.
            </p>
          </div>
          <div className="grid md:grid-cols-3 gap-8 relative">
            {steps.map((step, index) => (
              <div key={step.title} className="flex flex-col items-center text-center gap-4">
                <div className="w-14 h-14 rounded-full bg-primary flex items-center justify-center text-primary-foreground font-black text-xl shadow-md">
                  {index + 1}
                </div>
                <h3 className="text-xl font-bold text-foreground">{step.title}</h3>
                <p className="text-muted-foreground leading-relaxed">{step.description}</p>
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

const features = [
  {
    icon: Database,
    title: "Upload any dataset",
    description:
      "Import your CSV or JSON files in seconds. AmpleData automatically detects columns and prepares your data for enrichment.",
  },
  {
    icon: Search,
    title: "AI-powered web search",
    description:
      "Our engine searches the web for each row in your dataset, crawls the top results, and extracts exactly the information you need.",
  },
  {
    icon: Zap,
    title: "Structured results with confidence",
    description:
      "Get enriched columns with confidence scores and source URLs so you always know where the data came from and how reliable it is.",
  },
];

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
