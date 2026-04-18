import type { ApiClient } from "@/api";
import { useQueryClient, useMutation } from "@tanstack/react-query";

export function useCreateSampleSource(api: ApiClient) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.createSampleSource(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sources"] });
    },
  });
}
