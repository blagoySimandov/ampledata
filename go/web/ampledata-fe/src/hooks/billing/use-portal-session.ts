import type { ApiClient } from "@/api";
import { useMutation } from "@tanstack/react-query";

export function usePortalSession(api: ApiClient) {
  return useMutation({
    mutationFn: (returnUrl: string) => api.createPortalSession(returnUrl),
    onSuccess: ({ url }) => {
      window.location.href = url;
    },
  });
}
