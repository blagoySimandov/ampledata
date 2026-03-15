import type { TierResponse } from "@/api";
import { PlanTierCard } from "./plan-tier-card";

interface Props {
  tiers: TierResponse[];
  currentTierId: string | null;
  hasActiveSubscription: boolean;
  onSubscribe: (tierId: string) => void;
  onManagePortal: () => void;
  isPending?: boolean;
}

export function TierSelector({
  tiers,
  currentTierId,
  hasActiveSubscription,
  onSubscribe,
  onManagePortal,
  isPending,
}: Props) {
  return (
    <div className="flex flex-col sm:flex-row gap-3">
      {tiers.map((tier) => (
        <PlanTierCard
          key={tier.id}
          tier={tier}
          isCurrent={tier.id === currentTierId}
          hasActiveSubscription={hasActiveSubscription}
          onSubscribe={onSubscribe}
          onManagePortal={onManagePortal}
          isPending={isPending}
        />
      ))}
    </div>
  );
}
