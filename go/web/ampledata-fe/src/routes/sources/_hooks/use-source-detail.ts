import { useApi, useSource } from "@/hooks";
import { useState } from "react";

export function useSourceDetail(sourceId: string) {
  const api = useApi();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const { data: source, isLoading, isError } = useSource(api, sourceId);

  const mostRecentJob = source?.jobs[0];
  const activeJobId = selectedJobId ?? mostRecentJob?.job_id ?? null;
  const activeJob = source?.jobs.find((j) => j.job_id === activeJobId);

  return {
    source,
    isLoading,
    isError,
    sidebarOpen,
    setSidebarOpen,
    selectedJobId,
    setSelectedJobId,
    mostRecentJob,
    activeJobId,
    activeJob,
  };
}
