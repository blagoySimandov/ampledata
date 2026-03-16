import type { TierResponse } from "@/api";
import { PlanTierCard } from "./plan-tier-card";

interface Props {
  tiers: TierResponse[];
  currentTierId: string | null;
  hasActiveSubscription: boolean;
  onSubscribe: (tierId: string) => void;
  onUpgrade: (tierId: string) => void;
  onManagePortal: () => void;
  isPending?: boolean;
}

export function TierSelector({
  tiers,
  currentTierId,
  hasActiveSubscription,
  onSubscribe,
  onUpgrade,
  onManagePortal,
  isPending,
}: Props) {
  const currentIdx = tiers.findIndex((t) => t.id === currentTierId);

  return (
    <div className="flex flex-col sm:flex-row gap-3">
      {tiers.map((tier, idx) => (
        <PlanTierCard
          key={tier.id}
          tier={tier}
          isCurrent={tier.id === currentTierId}
          isUpgrade={hasActiveSubscription && currentIdx !== -1 && idx > currentIdx}
          hasActiveSubscription={hasActiveSubscription}
          onSubscribe={onSubscribe}
          onUpgrade={onUpgrade}
          onManagePortal={onManagePortal}
          isPending={isPending}
        />
      ))}
    </div>
  );
}
