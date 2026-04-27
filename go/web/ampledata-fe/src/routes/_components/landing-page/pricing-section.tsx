import { Check } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { TIERS } from "./constants";

function TierBadge({ badge }: { badge: string }) {
  return (
    <div className="absolute -top-3.5 inset-x-0 flex justify-center">
      <span className="bg-primary text-primary-foreground text-[9px] font-black px-3 py-1 rounded-full tracking-wide whitespace-nowrap">
        {badge}
      </span>
    </div>
  );
}

function TierCard({ tier }: { tier: (typeof TIERS)[number] }) {
  return (
    <div
      className={cn(
        "bg-card rounded-xl px-5 py-6 flex flex-col relative",
        tier.highlighted
          ? "border-2 border-primary shadow-[0_8px_28px_oklch(0.553_0.195_38.402/0.15)]"
          : "border border-border shadow-sm",
      )}
    >
      {tier.badge && <TierBadge badge={tier.badge} />}
      <div
        className={cn(
          "text-[13px] font-black uppercase tracking-wide mb-1.5",
          tier.highlighted ? "text-primary" : "text-foreground",
        )}
      >
        {tier.name}
      </div>
      <div className="flex items-baseline gap-1 mb-1">
        <span className="text-4xl font-black text-foreground tracking-tight">
          {tier.price === 0 ? "Free" : `$${tier.price}`}
        </span>
        {tier.price > 0 && (
          <span className="text-xs text-muted-foreground">/mo</span>
        )}
      </div>
      <p className="text-sm text-muted-foreground leading-relaxed mb-5">{tier.description}</p>
      <ul className="flex-1 flex flex-col gap-2 mb-5">
        {tier.features.map((f) => (
          <li key={f} className="flex items-start gap-2 text-sm text-muted-foreground">
            <Check
              className={cn("size-3.5 mt-0.5 shrink-0", tier.highlighted ? "text-primary" : "text-emerald-600")}
            />
            {f}
          </li>
        ))}
      </ul>
      <Button variant={tier.highlighted ? "default" : "outline"} className="w-full h-auto py-2.5 text-sm font-semibold rounded-lg" asChild>
        <Link to="/login">Get started</Link>
      </Button>
    </div>
  );
}

export function PricingSection() {
  return (
    <section id="pricing" className="py-14 md:py-20 bg-primary/[0.03] border-t border-border">
      <div className="max-w-[1280px] mx-auto px-4 md:px-6">
        <div className="text-center mb-14">
          <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-3 py-1 text-[11px] font-black uppercase tracking-widest mb-4">
            ⚡ Pricing
          </div>
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-2.5">
            Choose the plan that fits your data
          </h2>
          <p className="text-base text-muted-foreground">
            All plans include AI-powered enrichment. Upgrade or downgrade at any time.
          </p>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {TIERS.map((tier) => (
            <TierCard key={tier.id} tier={tier} />
          ))}
        </div>
        {/* TODO: re-enable once BYOK LLM key support is launched */}
        {/* <p className="text-center text-sm text-muted-foreground mt-7">
          Bring your own LLM keys to pay even less. &nbsp;·&nbsp; No contracts. Cancel any time.
        </p> */}
      </div>
    </section>
  );
}
