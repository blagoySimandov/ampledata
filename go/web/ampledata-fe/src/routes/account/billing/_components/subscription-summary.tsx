import type { SubscriptionStatusResponse, TierResponse } from "@/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { formatDate, formatDollars } from "@/lib/formatters";
import { ExternalLink, Zap } from "lucide-react";

interface Props {
  subscription: SubscriptionStatusResponse | undefined;
  tiers: TierResponse[];
  onManagePortal: () => void;
}

function NoSubscriptionState() {
  return (
    <div className="rounded-lg border border-dashed px-5 py-7 text-center space-y-1.5">
      <p className="text-sm font-semibold">No active subscription</p>
      <p className="text-xs text-muted-foreground">
        Choose a plan below to start enriching your data.
      </p>
    </div>
  );
}

function TokenUsageBar({ used, included }: { used: number; included: number }) {
  const pct = Math.min((used / included) * 100, 100);
  return (
    <div className="space-y-1.5">
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>
          Cells enriched this period{" "}
          <span className="text-muted-foreground/60">(1 cell = 1 token)</span>
        </span>
        <span className="font-medium tabular-nums">
          {used.toLocaleString()} / {included.toLocaleString()} &middot;{" "}
          {pct.toFixed(0)}%
        </span>
      </div>
      <Progress value={pct} className="h-1.5" />
    </div>
  );
}

function ActiveSubscriptionCard({
  subscription,
  currentTierName,
  currentTierPrice,
  onManagePortal,
}: {
  subscription: SubscriptionStatusResponse;
  currentTierName: string;
  currentTierPrice: string | null;
  onManagePortal: () => void;
}) {
  return (
    <Card className="gap-0">
      <CardContent className="pt-5 space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex flex-col gap-1.5">
            <Badge variant="default" className="w-fit gap-1.5 text-xs">
              <Zap className="size-3" />
              Active
            </Badge>
            <span className="text-xl font-black tracking-tight">
              {currentTierName}
            </span>
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              {currentTierPrice && <span>{currentTierPrice} / month</span>}
              {currentTierPrice && <span>&middot;</span>}
              <span>Renews {formatDate(subscription.current_period_end)}</span>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={onManagePortal}
            className="shrink-0 gap-1.5"
          >
            Manage subscription
            <ExternalLink className="size-3" />
          </Button>
        </div>
        <TokenUsageBar
          used={subscription.tokens_used}
          included={subscription.tokens_included}
        />
      </CardContent>
    </Card>
  );
}

export function SubscriptionSummary({ subscription, tiers, onManagePortal }: Props) {
  if (!subscription?.tier) return <NoSubscriptionState />;

  const currentTier = tiers.find((t) => t.id === subscription.tier);

  return (
    <ActiveSubscriptionCard
      subscription={subscription}
      currentTierName={currentTier?.display_name ?? subscription.tier}
      currentTierPrice={
        currentTier ? formatDollars(currentTier.monthly_price_cents) : null
      }
      onManagePortal={onManagePortal}
    />
  );
}
