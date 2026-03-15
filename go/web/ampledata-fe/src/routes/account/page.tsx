import { useBilling, useMe } from "@/hooks";
import { ProfileCard } from "./_components/profile-card";
import { PersonalInfoSection } from "./_components/personal-info-section";
import { SubscriptionSummary } from "./billing/_components/subscription-summary";
import { TierSelector } from "./billing/_components/tier-selector";

function BillingSection({
  onManagePortal,
}: {
  onManagePortal: () => void;
}) {
  const { tiers, subscription, subscribe } = useBilling();

  const tiersData = tiers.data ?? [];
  const hasActiveSubscription = !!subscription.data?.tier;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          Billing
        </h2>
        {hasActiveSubscription && (
          <p className="text-xs text-muted-foreground">
            To switch plans,{" "}
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
        onManagePortal={onManagePortal}
        isPending={subscribe.isPending}
      />
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
