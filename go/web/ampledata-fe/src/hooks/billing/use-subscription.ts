import type { ApiClient } from "@/api";
import { useQuery } from "@tanstack/react-query";

export function useSubscription(api: ApiClient) {
  return useQuery({
    queryKey: ["subscription"],
    queryFn: () => api.getSubscriptionStatus(),
  });
}
