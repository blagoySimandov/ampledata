import { useQuery } from "@tanstack/react-query";
import { useApi } from "./use-api";

export function useMe() {
  const api = useApi();
  return useQuery({
    queryKey: ["me"],
    queryFn: () => api.getMe(),
  });
}
