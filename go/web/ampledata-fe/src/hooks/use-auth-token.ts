import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@workos-inc/authkit-react";

export function useAuthToken() {
  const { getAccessToken, organizationId } = useAuth();
  return useQuery({
    queryKey: ["me", "authToken", organizationId],
    queryFn: () => getAccessToken(),
  });
}
