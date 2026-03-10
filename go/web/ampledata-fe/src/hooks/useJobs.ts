// src/hooks/useJobs.ts
import { useQuery, useMutation, useQueryClient, useQueries } from '@tanstack/react-query';
import { ApiClient } from '../api';
import type { SignedURLRequest, SelectKeyRequest, SourceJobSummary } from '../api';

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
  offset = 0,
  limit = 50,
  stage = 'all',
  sort = 'updated_at_desc',
  refetchInterval: number | false = 5000
) {
  return useQuery({
    queryKey: ['job-rows', jobId, offset, limit, stage, sort],
    queryFn: () => api.getJobRows(jobId, offset, limit, stage, sort),
    refetchInterval,
  });
}

export function useJobResults(api: ApiClient, jobId: string, start = 0, limit = 0) {
  return useQuery({
    queryKey: ['job-results', jobId, start, limit],
    queryFn: () => api.getJobResults(jobId, start, limit),
  });
}

export function useAllJobsRows(api: ApiClient, jobs: SourceJobSummary[]) {
  return useQueries({
    queries: jobs.map((job) => ({
      queryKey: ['job-rows', job.job_id, 0, 10000, 'all', 'updated_at_desc'],
      queryFn: () => api.getJobRows(job.job_id, 0, 10000, 'all', 'updated_at_desc'),
      refetchInterval: job.status === 'RUNNING' || job.status === 'PENDING' ? 5000 : false,
    })),
  });
}

export function useCancelJob(api: ApiClient) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (jobId: string) => api.cancelJob(jobId),
    onSuccess: (_, jobId) => {
      queryClient.invalidateQueries({ queryKey: ['job-progress', jobId] });
      queryClient.invalidateQueries({ queryKey: ['sources'] });
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
