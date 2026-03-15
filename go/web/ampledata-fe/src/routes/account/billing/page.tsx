import { useBilling } from "@/hooks";
import { SubscriptionSummary } from "./_components/subscription-summary";
import { TierSelector } from "./_components/tier-selector";

export function BillingPage() {
  const { tiers, subscription, subscribe, portal } = useBilling();

  function handleManagePortal() {
    portal.mutate(window.location.href);
  }

  if (tiers.isLoading || subscription.isLoading) {
    return <div className="text-xs text-muted-foreground py-4">Loading...</div>;
  }

  const tiersData = tiers.data ?? [];
  const hasActiveSubscription = !!subscription.data?.tier;

  return (
    <div className="max-w-3xl space-y-6">
      <SubscriptionSummary
        subscription={subscription.data}
        tiers={tiersData}
        onManagePortal={handleManagePortal}
      />
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Plans
          </h2>
          {hasActiveSubscription && (
            <p className="text-xs text-muted-foreground">
              To switch plans,{" "}
              <button
                onClick={handleManagePortal}
                className="underline underline-offset-2 hover:text-foreground transition-colors"
              >
                manage your subscription
              </button>
            </p>
          )}
        </div>
        <TierSelector
          tiers={tiersData}
          currentTierId={subscription.data?.tier ?? null}
          hasActiveSubscription={hasActiveSubscription}
          onSubscribe={(tierId) => subscribe.mutate({ tier_id: tierId })}
          onManagePortal={handleManagePortal}
          isPending={subscribe.isPending}
        />
      </div>
    </div>
  );
}
