import { useApi } from "../use-api";
import { useCancelSubscription } from "./use-cancel-sub";
import { useListTiers } from "./use-list-tiers";
import { usePortalSession } from "./use-portal-session";
import { useSubscribe } from "./use-subscribe";
import { useSubscription } from "./use-subscription";

export function useBilling() {
  const api = useApi();
  const tiers = useListTiers(api);
  const subscription = useSubscription(api);
  const subscribe = useSubscribe(api);
  const cancel = useCancelSubscription(api);
  const portal = usePortalSession(api);
  return { tiers, subscription, subscribe, cancel, portal };
}
