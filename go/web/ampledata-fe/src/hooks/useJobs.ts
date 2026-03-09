// src/hooks/useJobs.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiClient } from '../api';
import type { SignedURLRequest, StartJobRequest, SelectKeyRequest } from '../api';

export function useListJobs(api: ApiClient, offset: number = 0, limit: number = 50) {
  return useQuery({
    queryKey: ['jobs', offset, limit],
    queryFn: () => api.getJobs(offset, limit),
  });
}

export function useJobProgress(api: ApiClient, jobId: string, refetchInterval: number | false = 5000) {
  return useQuery({
    queryKey: ['job-progress', jobId],
    queryFn: () => api.getJobProgress(jobId),
    refetchInterval,
  });
}

export function useJobRows(
  api: ApiClient,
  jobId: string,
  offset: number = 0,
  limit: number = 50,
  stage: string = 'all',
  sort: string = 'updated_at_desc',
  refetchInterval: number | false = 5000
) {
  return useQuery({
    queryKey: ['job-rows', jobId, offset, limit, stage, sort],
    queryFn: () => api.getJobRows(jobId, offset, limit, stage, sort),
    refetchInterval,
  });
}

export function useJobResults(api: ApiClient, jobId: string, start: number = 0, limit: number = 0) {
  return useQuery({
    queryKey: ['job-results', jobId, start, limit],
    queryFn: () => api.getJobResults(jobId, start, limit),
  });
}

export function useCancelJob(api: ApiClient) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (jobId: string) => api.cancelJob(jobId),
    onSuccess: (_, jobId) => {
      // Invalidate relevant queries so the UI updates to show the job as cancelled
      queryClient.invalidateQueries({ queryKey: ['job-progress', jobId] });
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
    },
  });
}

export function useGetSignedUrl(api: ApiClient) {
  return useMutation({
    mutationFn: (req: SignedURLRequest) => api.getSignedUrl(req),
  });
}

export function useUploadFile(api: ApiClient) {
  return useMutation({
    mutationFn: ({ url, file }: { url: string; file: File }) => api.uploadFile(url, file),
  });
}

export function useSelectKey(api: ApiClient) {
  return useMutation({
    mutationFn: (req: SelectKeyRequest) => api.selectKey(req),
  });
}

export function useStartJob(api: ApiClient) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ jobId, req }: { jobId: string; req: StartJobRequest }) => 
      api.startJob(jobId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
    },
  });
}
