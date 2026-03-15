import type { ApiClient, CreateSubscriptionRequest } from "@/api";
import { useMutation } from "@tanstack/react-query";

type CreateSubscriptionMutationReq = Pick<CreateSubscriptionRequest, "tier_id">;
export function useSubscribe(api: ApiClient) {
  return useMutation({
    mutationFn: (req: CreateSubscriptionMutationReq) =>
      api.createSubscriptionCheckout({
        tier_id: req.tier_id,
        success_url: window.location.href,
        cancel_url: window.location.href,
      }),
    onSuccess: (data) => {
      window.location.href = data.checkout_url;
    },
  });
}
