import type { ApiClient } from "@/api";
import { useQuery } from "@tanstack/react-query";

export function useListTiers(api: ApiClient) {
  return useQuery({
    queryKey: ["tiers"],
    queryFn: () => api.listTiers(),
    staleTime: Infinity,
  });
}
