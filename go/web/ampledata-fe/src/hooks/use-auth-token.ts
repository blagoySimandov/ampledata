import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@workos-inc/authkit-react";

export function useAuthToken() {
  const { getAccessToken } = useAuth();
  return useQuery({
    queryKey: ["me"],
    queryFn: () => getAccessToken(),
  });
}
