import type { ApiClient } from "@/api";
import { useQuery } from "@tanstack/react-query";

export function useListSources(api: ApiClient, offset = 0, limit = 50) {
  return useQuery({
    queryKey: ["sources", offset, limit],
    queryFn: () => api.getSources(offset, limit),
  });
}
