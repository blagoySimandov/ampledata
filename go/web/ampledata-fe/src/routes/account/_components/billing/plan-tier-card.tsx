import { Check, ExternalLink, ArrowUp } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { TierResponse } from "@/api";
import { formatDollars, formatOveragePrice } from "@/lib/formatters";

interface Props {
  tier: TierResponse;
  isCurrent: boolean;
  isUpgrade: boolean;
  hasActiveSubscription: boolean;
  onSubscribe: (tierId: string) => void;
  onUpgrade: (tierId: string) => void;
  onManagePortal: () => void;
  isPending?: boolean;
}

function TierFeatures({ tier }: { tier: TierResponse }) {
  return (
    <ul className="space-y-1.5 mt-3">
      <li className="flex items-center gap-2 text-xs">
        <Check className="size-3 text-primary shrink-0" />
        {tier.included_tokens.toLocaleString()} cells enriched / month
      </li>
      <li className="flex items-center gap-2 text-xs">
        <Check className="size-3 text-primary shrink-0" />
        {formatOveragePrice(tier.overage_price_cents_decimal)} per extra token
      </li>
    </ul>
  );
}

function TierCta({
  tier,
  isCurrent,
  isUpgrade,
  hasActiveSubscription,
  onSubscribe,
  onUpgrade,
  onManagePortal,
  isPending,
}: Props) {
  if (!hasActiveSubscription) {
    return (
      <Button className="w-full mt-4" size="sm" onClick={() => onSubscribe(tier.id)} disabled={isPending}>
        Subscribe
      </Button>
    );
  }

  if (isCurrent) return null;

  if (isUpgrade) {
    return (
      <Button className="w-full mt-4 gap-1.5" size="sm" onClick={() => onUpgrade(tier.id)} disabled={isPending}>
        <ArrowUp className="size-3" />
        Upgrade to {tier.display_name}
      </Button>
    );
  }

  return (
    <Button className="w-full mt-4 gap-1.5" variant="outline" size="sm" onClick={onManagePortal}>
      Switch plan
      <ExternalLink className="size-3" />
    </Button>
  );
}

export function PlanTierCard(props: Props) {
  const { tier, isCurrent } = props;

  return (
    <Card className={cn("gap-0 flex-1", isCurrent && "ring-2 ring-primary")}>
      <CardHeader className="pb-0">
        {isCurrent && (
          <Badge variant="default" className="w-fit mb-1">
            Current plan
          </Badge>
        )}
        <CardTitle className="text-sm font-black">
          {tier.display_name}
        </CardTitle>
        <div className="flex items-baseline gap-1 mt-1">
          <span className="text-2xl font-black">
            {formatDollars(tier.monthly_price_cents)}
          </span>
          <span className="text-xs text-muted-foreground">/ month</span>
        </div>
        <CardDescription className="mt-1">
          {tier.included_tokens.toLocaleString()} cells / month
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-0">
        <TierFeatures tier={tier} />
        <TierCta {...props} />
      </CardContent>
    </Card>
  );
}
