import { ArrowRight, Check, Zap } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

const TIERS = [
  {
    id: "starter",
    name: "Starter",
    price: 29,
    tokens: 1000,
    overagePrice: "0.025",
    description: "Perfect for individuals and small projects.",
    features: [
      "1,000 cells enriched / month",
      "$0.025 per extra cell",
      "Web search enrichment",
      "CSV & JSON file support",
      "Email support",
    ],
    highlighted: false,
    badge: null,
  },
  {
    id: "pro",
    name: "Pro",
    price: 99,
    tokens: 5000,
    overagePrice: "0.018",
    description: "For growing teams with higher enrichment needs.",
    features: [
      "5,000 cells enriched / month",
      "$0.018 per extra cell",
      "Web search enrichment",
      "CSV & JSON file support",
      "Priority support",
      "Bulk operations",
    ],
    highlighted: true,
    badge: "Most popular",
  },
  {
    id: "enterprise",
    name: "Enterprise",
    price: 299,
    tokens: 25000,
    overagePrice: "0.01",
    description: "High-volume enrichment for data-driven organizations.",
    features: [
      "25,000 cells enriched / month",
      "$0.01 per extra cell",
      "Web search enrichment",
      "CSV & JSON file support",
      "Dedicated support",
      "Bulk operations",
      "Custom integrations",
    ],
    highlighted: false,
    badge: null,
  },
] as const;

export function PricingSection() {
  return (
    <section id="pricing" className="bg-background py-24">
      <div className="container mx-auto px-4">
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-4 py-1.5 text-sm font-medium mb-6">
            <Zap className="size-4" />
            <span>Simple, transparent pricing</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
            Choose the plan that fits your data
          </h2>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            All plans include full access to AI-powered enrichment. Upgrade or
            downgrade at any time. Pay only for what you use.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto items-start">
          {TIERS.map((tier) => (
            <Card
              key={tier.id}
              className={cn(
                "relative flex flex-col gap-0",
                tier.highlighted &&
                  "ring-2 ring-primary shadow-lg shadow-primary/10",
              )}
            >
              {tier.badge && (
                <div className="absolute -top-3.5 inset-x-0 flex justify-center">
                  <Badge className="px-3 py-1 text-xs font-semibold">
                    {tier.badge}
                  </Badge>
                </div>
              )}
              <CardHeader className="pb-4">
                <CardTitle
                  className={cn(
                    "text-lg font-black",
                    tier.highlighted && "text-primary",
                  )}
                >
                  {tier.name}
                </CardTitle>
                <div className="flex items-baseline gap-1 mt-2">
                  <span className="text-4xl font-black text-foreground">
                    ${tier.price}
                  </span>
                  <span className="text-sm text-muted-foreground">/ month</span>
                </div>
                <CardDescription className="mt-2 text-sm leading-relaxed">
                  {tier.description}
                </CardDescription>
              </CardHeader>

              <CardContent className="flex flex-col flex-1 pt-0">
                <ul className="space-y-2.5 mb-8">
                  {tier.features.map((feature) => (
                    <li key={feature} className="flex items-start gap-2.5">
                      <Check
                        className={cn(
                          "size-4 shrink-0 mt-0.5",
                          tier.highlighted
                            ? "text-primary"
                            : "text-muted-foreground",
                        )}
                      />
                      <span className="text-sm text-muted-foreground">
                        {feature}
                      </span>
                    </li>
                  ))}
                </ul>

                <Button
                  className="w-full mt-auto"
                  variant={tier.highlighted ? "default" : "outline"}
                  asChild
                >
                  <Link to="/login">
                    Get started <ArrowRight className="ml-2 size-4" />
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>

        <p className="text-center text-sm text-muted-foreground mt-10">
          All prices in USD. No contracts. Cancel anytime.
        </p>
      </div>
    </section>
  );
}
