import { Bot, CheckCircle, FileText, Search, Sparkles } from "lucide-react";

export const MOCK_ROWS = [
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
] as const;

export const HERO_SOURCE_HEADERS = ["Company", "Website"] as const;
export const HERO_ENRICHED_HEADERS = ["CEO ✦", "Founded ✦", "Revenue ✦"] as const;

export const PIPELINE_NODES = [
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

export const STEPS = [
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
] as const;

export const BENEFITS = [
  "No coding or scripting required — anyone can enrich a dataset",
  "Works with CSV or JSON files of various sizes",
  "Transparent confidence scores so you can trust the output",
  "Source URLs for every enriched value for easy verification",
  "Run enrichments on demand or schedule them automatically",
] as const;

export const PIPELINE_ANIMATION_INTERVAL_MS = 1100;
