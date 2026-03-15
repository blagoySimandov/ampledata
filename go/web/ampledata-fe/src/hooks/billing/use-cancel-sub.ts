import type { ApiClient } from "@/api";
import { useQueryClient, useMutation } from "@tanstack/react-query";

export function useCancelSubscription(api: ApiClient) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.cancelSubscription(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
    },
  });
}
