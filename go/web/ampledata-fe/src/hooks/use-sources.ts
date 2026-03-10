// src/hooks/use-sources.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiClient } from '../api';
import type { EnrichRequest, SourceDetail } from '../api';

export function useListSources(api: ApiClient, offset = 0, limit = 50) {
  return useQuery({
    queryKey: ['sources', offset, limit],
    queryFn: () => api.getSources(offset, limit),
  });
}

export function useSource(api: ApiClient, sourceId: string) {
  return useQuery({
    queryKey: ['source', sourceId],
    queryFn: () => api.getSource(sourceId),
    refetchInterval: (query) => hasRunningJob(query.state.data) ? 5000 : false,
  });
}

function hasRunningJob(source: SourceDetail | undefined): boolean {
  return source?.jobs.some(j => j.status === 'RUNNING' || j.status === 'PENDING') ?? false;
}

export function useEnrich(api: ApiClient, sourceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: EnrichRequest) => api.enrich(sourceId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['source', sourceId] });
      queryClient.invalidateQueries({ queryKey: ['sources'] });
    },
  });
}
