import type { ApiClient } from "@/api";
import { useQuery } from "@tanstack/react-query";

export function useListTemplates(api: ApiClient) {
  return useQuery({
    queryKey: ["templates"],
    queryFn: () => api.listTemplates(),
  });
}
