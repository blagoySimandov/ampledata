import type { ApiClient, EnrichRequest } from "@/api";
import { useQueryClient, useMutation } from "@tanstack/react-query";

export function useEnrich(api: ApiClient, sourceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: EnrichRequest) => api.enrich(sourceId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["source", sourceId] });
      queryClient.invalidateQueries({ queryKey: ["sources"] });
    },
  });
}
