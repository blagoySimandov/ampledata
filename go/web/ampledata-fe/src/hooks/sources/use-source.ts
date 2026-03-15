import type { SourceDetail, ApiClient } from "@/api";
import { useQuery } from "@tanstack/react-query";

function hasRunningJob(source: SourceDetail | undefined): boolean {
  return (
    source?.jobs.some(
      (j) => j.status === "RUNNING" || j.status === "PENDING",
    ) ?? false
  );
}

export function useSource(api: ApiClient, sourceId: string) {
  return useQuery({
    queryKey: ["source", sourceId],
    queryFn: () => api.getSource(sourceId),
    refetchInterval: (query) =>
      hasRunningJob(query.state.data) ? 5000 : false,
  });
}
