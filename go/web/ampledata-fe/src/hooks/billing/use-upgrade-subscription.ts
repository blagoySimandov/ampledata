import type { ApiClient } from "@/api";
import { useQueryClient, useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

export function useUpgradeSubscription(api: ApiClient) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (tierId: string) => api.upgradeSubscription(tierId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
      toast.success("Subscription upgraded successfully");
    },
    onError: (e) =>
      toast.error(e instanceof Error ? e.message : "Upgrade failed"),
  });
}
