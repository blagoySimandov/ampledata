import { Bot, CheckCircle, FileText, Search, Sparkles } from "lucide-react";

export const MOCK_ROWS = [
  { company: "Acme Corp", website: "acme.com", ceo: "Jane Smith", founded: "2005", revenue: "$4.2M" },
  { company: "TechStart", website: "techstart.io", ceo: "David Lee", founded: "2019", revenue: "$1.8M" },
  { company: "DataFlow AI", website: "dataflow.ai", ceo: "Sara Kim", founded: "2015", revenue: "$12M" },
  { company: "CloudBase", website: "cloudbase.co", ceo: "Tom Allen", founded: "2018", revenue: "$6.5M" },
] as const;

export const HERO_SOURCE_HEADERS = ["Company", "Website"] as const;
export const HERO_ENRICHED_HEADERS = ["CEO ✶", "Founded ✶", "Revenue ✶"] as const;

export const HERO_ROWS = [
  { name: "Retool", industry: "Dev Tools", founderStmt: '"We\'re doubling down on AI agents for internal tools."', retracted: "-", remote: "Hybrid" },
  { name: "Vercel", industry: "Infrastructure", founderStmt: '"Pricing for the next billion developers."', retracted: "-", remote: "Remote-first" },
  { name: "Notion", industry: "Productivity", founderStmt: '"AI should feel like a thought, not a feature."', retracted: "-", remote: "Hybrid" },
  { name: "Linear", industry: "Project Mgmt", founderStmt: '"Software quality is a product decision."', retracted: "-", remote: "Remote-first" },
  { name: "Loom", industry: "Async Video", founderStmt: "-", retracted: "-", remote: "Remote-first" },
];

export const HERO_COLS = [
  "Company",
  "Industry",
  "Founder's pricing stance",
  "Paper retracted?",
  "Remote policy",
] as const;

export const HERO_CONFIDENCES = [97, 99, 91, 88, 95] as const;

export const PIPELINE_NODES = [
  { icon: FileText, label: "Upload file", desc: "Drop in CSV or JSON", color: "bg-slate-100 text-slate-700 border-slate-200" },
  { icon: Search, label: "Find sources", desc: "We search the web for each row", color: "bg-blue-50 text-blue-700 border-blue-200" },
  { icon: Bot, label: "Extract answers", desc: "AI fills in the missing fields", color: "bg-violet-50 text-violet-700 border-violet-200" },
  { icon: CheckCircle, label: "Quality check", desc: "Confidence scores help you verify", color: "bg-emerald-50 text-emerald-700 border-emerald-200" },
  { icon: Sparkles, label: "Ready to use", desc: "Export enriched rows instantly", color: "bg-primary/10 text-primary border-primary/30" },
] as const;

export const STEPS = [
  {
    title: "Upload your dataset",
    description: "Drag and drop a CSV or JSON file into AmpleData. Your data is securely stored and ready for enrichment.",
    image: "https://images.unsplash.com/photo-1504868584819-f8e8b4b6d7e3?auto=format&fit=crop&w=600&q=80",
  },
  {
    title: "Define enrichment fields",
    description: "Tell AmpleData which columns to use as context and what new information you want to extract for each row.",
    image: "https://images.unsplash.com/photo-1432888498266-38ffec3eaf0a?auto=format&fit=crop&w=600&q=80",
  },
  {
    title: "Get enriched data",
    description: "AmpleData runs the enrichment jobs and populates your new columns automatically, with sources and confidence scores.",
    image: "https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=600&q=80",
  },
] as const;

export const BENEFITS = [
  "No coding or scripting required, anyone can enrich a dataset",
  "Works with CSV or JSON files of various sizes",
  "Transparent confidence scores so you can trust the output",
  "Source URLs for every enriched value for easy verification",
  "Run enrichments on demand or schedule them automatically",
] as const;

export const PIPELINE_ANIMATION_INTERVAL_MS = 1100;

export const FAQS = [
  {
    q: "How is this different from Clay?",
    a: "Clay is built around a contacts database, great for lead gen. AmpleData is built around any list, any question, any domain. Companies and people are valid lists; they're just not the only list. And we cost about 1/10th as much per cell.",
  },
  {
    q: "Can I trust the data?",
    a: "Every cell is cited. Click any cell to see the source URL, the extracted snippet, and the reasoning. Every cell also carries a confidence score. You're not asked to trust the AI, you're asked to trust your own ability to verify what it did.",
  },
  {
    q: "What if a column is wrong?",
    a: "Edit the column's natural-language definition and re-run just that column. Only that column gets re-charged. Nothing else changes.",
  },
  {
    q: "What models do you use?",
    a: "Gemini and Groq under the hood. You can also bring your own API keys to pay even less and keep usage under your own account.",
  },
  {
    q: "Is my data private?",
    a: "Your inputs are not used for training. Data is stored only long enough to complete the enrichment job and let you export results. We don't sell or share your data.",
  },
  {
    q: "Do you have an API?",
    a: "Not yet, webhook support is coming. If this is blocking you, email us and you'll be first to know.",
  },
] as const;

export const LANDING_FEATURES = [
  {
    num: "01",
    title: "Every cell is cited",
    body: "Not 'AI-generated.' Researched. Click any cell to see the source URL, the extracted snippet, and the reasoning. If a source is wrong, you'll know which one.",
  },
  {
    num: "02",
    title: "Confidence scores you can act on",
    body: "AmpleData tells you when it's sure and when it isn't. Sort by confidence, spot-check the weak cells, and re-run a column with a refined prompt.",
  },
  {
    num: "03",
    title: "Per-cell pricing, not per-seat",
    body: "Pay only for what you enrich. No minimum, no per-seat license, no annual contract. A fraction of what comparable tools charge.",
  },
] as const;

export const COMPARISONS = [
  {
    label: "vs. ChatGPT / manual research",
    body: "AmpleData runs in parallel across rows, cites every answer, scores its own confidence, and doesn't hallucinate URLs. It's the difference between asking a friend and hiring an analyst.",
  },
  {
    label: "vs. Clay",
    body: "AmpleData isn't built around a contacts database. It's built around any list, any question, any domain. And it costs about 1/10th as much per cell.",
  },
] as const;

export const TIERS = [
  {
    id: "free",
    name: "Free",
    price: 0,
    description: "Try AmpleData with no commitment.",
    features: ["100 cells enriched", "No overage", "Web search enrichment", "CSV support"],
    highlighted: false,
    badge: null,
  },
  {
    id: "starter",
    name: "Starter",
    price: 29,
    description: "Perfect for individuals and small projects.",
    features: ["1,000 cells / month", "$0.025 per extra cell", "Web search enrichment", "Email support"],
    highlighted: false,
    badge: null,
  },
  {
    id: "pro",
    name: "Pro",
    price: 99,
    description: "For growing teams with higher enrichment needs.",
    features: ["5,000 cells / month", "$0.018 per extra cell", "Priority support", "Bulk operations"],
    highlighted: true,
    badge: "Most popular",
  },
  {
    id: "enterprise",
    name: "Enterprise",
    price: 299,
    description: "High-volume enrichment for data-driven orgs.",
    features: ["25,000 cells / month", "$0.01 per extra cell", "Dedicated support", "Custom integrations"],
    highlighted: false,
    badge: null,
  },
] as const;
