import { useState } from "react";
import { useBilling, useMe } from "@/hooks";
import { ProfileCard } from "./_components/profile-card";
import { PersonalInfoSection } from "./_components/personal-info-section";
import { SubscriptionSummary } from "./_components/billing/subscription-summary";
import { TierSelector } from "./_components/billing/tier-selector";
import { UpgradeDialog } from "./_components/billing/upgrade-dialog";

function BillingSection({ onManagePortal }: { onManagePortal: () => void }) {
  const { tiers, subscription, subscribe, upgrade } = useBilling();
  const [pendingUpgradeTierId, setPendingUpgradeTierId] = useState<string | null>(null);

  const tiersData = tiers.data ?? [];
  const hasActiveSubscription = !!subscription.data?.tier;
  const currentTier = tiersData.find((t) => t.id === subscription.data?.tier);
  const pendingUpgradeTier = tiersData.find((t) => t.id === pendingUpgradeTierId);

  const handleUpgradeConfirm = () => {
    if (!pendingUpgradeTierId) return;
    upgrade.mutate(pendingUpgradeTierId, { onSuccess: () => setPendingUpgradeTierId(null) });
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          Billing
        </h2>
        {hasActiveSubscription && (
          <p className="text-xs text-muted-foreground">
            To downgrade,{" "}
            <button
              onClick={onManagePortal}
              className="underline underline-offset-2 hover:text-foreground transition-colors"
            >
              manage your subscription
            </button>
          </p>
        )}
      </div>
      <SubscriptionSummary
        subscription={subscription.data}
        tiers={tiersData}
        onManagePortal={onManagePortal}
      />
      <TierSelector
        tiers={tiersData}
        currentTierId={subscription.data?.tier ?? null}
        hasActiveSubscription={hasActiveSubscription}
        onSubscribe={(tierId) => subscribe.mutate({ tier_id: tierId })}
        onUpgrade={setPendingUpgradeTierId}
        onManagePortal={onManagePortal}
        isPending={subscribe.isPending}
      />
      {pendingUpgradeTierId && currentTier && pendingUpgradeTier && subscription.data && (
        <UpgradeDialog
          currentTier={currentTier}
          newTier={pendingUpgradeTier}
          subscription={subscription.data}
          isPending={upgrade.isPending}
          onConfirm={handleUpgradeConfirm}
          onClose={() => setPendingUpgradeTierId(null)}
        />
      )}
    </div>
  );
}

export function AccountPage() {
  const me = useMe();
  const { subscription, tiers, portal } = useBilling();

  const handleManagePortal = () => portal.mutate(window.location.href);

  if (me.isLoading || subscription.isLoading || tiers.isLoading) {
    return <div className="text-xs text-muted-foreground py-4">Loading…</div>;
  }

  const currentTier = tiers.data?.find((t) => t.id === subscription.data?.tier);

  return (
    <div className="max-w-3xl space-y-6">
      <ProfileCard
        user={me.data}
        tierName={currentTier?.display_name ?? subscription.data?.tier}
      />
      <PersonalInfoSection user={me.data} />
      <BillingSection onManagePortal={handleManagePortal} />
    </div>
  );
}
