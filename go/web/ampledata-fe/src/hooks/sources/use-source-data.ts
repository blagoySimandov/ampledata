import type { ApiClient } from "@/api";
import { useQuery } from "@tanstack/react-query";

export function useSourceData(api: ApiClient, sourceId: string) {
  return useQuery({
    queryKey: ["source-data", sourceId],
    queryFn: () => api.getSourceData(sourceId),
    staleTime: Infinity, // Source data doesn't change
  });
}
