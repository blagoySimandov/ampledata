import type { SubscriptionStatusResponse, TierResponse } from "@/api";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useListTiers } from "@/hooks/billing/use-list-tiers";
import { useSubscription } from "@/hooks/billing/use-subscription";
import { useApi } from "@/hooks/use-api";
import { formatDate, formatOveragePrice, usePct } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import { Link } from "@tanstack/react-router";
import { Zap } from "lucide-react";

function urgencyColor(pct: number) {
  if (pct >= 90) return "text-destructive";
  if (pct >= 75) return "text-amber-500 dark:text-amber-400";
  return "text-muted-foreground";
}

function MiniBar({ pct }: { pct: number }) {
  return (
    <div className="h-1 w-full rounded-full bg-muted overflow-hidden">
      <div
        className={cn(
          "h-full rounded-full transition-all",
          pct >= 90 ? "bg-destructive" : "bg-primary",
        )}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

function PopoverDetails({
  subscription,
  tier,
}: {
  subscription: SubscriptionStatusResponse;
  tier: TierResponse | undefined;
}) {
  const { tokens_used, tokens_included, current_period_end } = subscription;
  const remaining = Math.max(tokens_included - tokens_used, 0);
  const pct = usePct(tokens_used, tokens_included);

  return (
    <div className="space-y-3 p-1">
      <div className="space-y-0.5">
        <p className="text-xs font-semibold">
          {subscription.tier ? "Free cells included in your plan" : "Free plan credits"}
        </p>
        {current_period_end && (
          <p className="text-xs text-muted-foreground">
            Resets {formatDate(current_period_end)}
          </p>
        )}
      </div>
      <div className="space-y-1.5">
        <MiniBar pct={pct} />
        <div className="flex justify-between text-xs tabular-nums">
          <span className="text-muted-foreground">
            {tokens_used.toLocaleString()} used
          </span>
          <span className={cn("font-semibold", urgencyColor(pct))}>
            {remaining.toLocaleString()} left
          </span>
        </div>
      </div>
      {tier ? (
        <p className="text-xs text-muted-foreground border-t pt-2">
          {pct >= 100 ? (
            <>Per-cell billing is now active at {formatOveragePrice(tier.overage_price_cents_decimal)}/cell.</>
          ) : (
            <>After your free cells run out, you're billed at {formatOveragePrice(tier.overage_price_cents_decimal)}/cell.</>
          )}
        </p>
      ) : (
        <p className="text-xs text-muted-foreground border-t pt-2">
          Upgrade to a plan to get more cells.
        </p>
      )}
      <Link
        to="/account"
        className="block text-xs text-primary font-medium hover:underline"
      >
        {subscription.tier ? "Manage billing →" : "View plans →"}
      </Link>
    </div>
  );
}

function CreditsPill({
  subscription,
  tier,
}: {
  subscription: SubscriptionStatusResponse;
  tier: TierResponse | undefined;
}) {
  const { tokens_used, tokens_included } = subscription;
  const remaining = Math.max(tokens_included - tokens_used, 0);
  const pct = usePct(tokens_used, tokens_included);

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          className={cn(
            "flex flex-col gap-1 rounded-lg border px-3 py-1.5 w-36",
            "bg-background hover:bg-accent transition-colors cursor-pointer",
            "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          )}
        >
          <div className="flex items-center justify-between gap-1.5">
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Zap className="size-3 shrink-0" />
              Cells
            </span>
            <span
              className={cn(
                "text-xs font-semibold tabular-nums",
                urgencyColor(pct),
              )}
            >
              {remaining.toLocaleString()} left
            </span>
          </div>
          <MiniBar pct={pct} />
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-60">
        <PopoverDetails subscription={subscription} tier={tier} />
      </PopoverContent>
    </Popover>
  );
}

export function CreditsWidget() {
  const api = useApi();
  const { data: subscription } = useSubscription(api);
  const { data: tiers } = useListTiers(api);

  if (!subscription) return null;

  const tier = tiers?.find((t) => t.id === subscription.tier);

  return <CreditsPill subscription={subscription} tier={tier} />;
}
