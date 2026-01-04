import { useQuery, useMutation, UseQueryOptions, UseMutationOptions } from "@tanstack/react-query";
import { useAuth } from "@workos-inc/authkit-react";
import { createApiClient } from "@/lib/api-client";
import { API_ENDPOINTS, QUERY_KEYS } from "@/constants";
import type { ApiResponse } from "@/types";

export function useApiClient() {
  const { getAccessToken } = useAuth();
  return createApiClient(getAccessToken);
}

export function useJobProgress(
  jobId: string,
  options?: Omit<UseQueryOptions<ApiResponse<unknown>>, "queryKey" | "queryFn">
) {
  const apiClient = useApiClient();

  return useQuery({
    queryKey: QUERY_KEYS.JOB_PROGRESS(jobId),
    queryFn: () => apiClient.get(API_ENDPOINTS.JOB_PROGRESS(jobId)),
    enabled: !!jobId,
    ...options,
  });
}

export function useJobResults(
  jobId: string,
  options?: Omit<UseQueryOptions<ApiResponse<unknown>>, "queryKey" | "queryFn">
) {
  const apiClient = useApiClient();

  return useQuery({
    queryKey: QUERY_KEYS.JOB_RESULTS(jobId),
    queryFn: () => apiClient.get(API_ENDPOINTS.JOB_RESULTS(jobId)),
    enabled: !!jobId,
    ...options,
  });
}

export function useEnrichMutation(
  options?: UseMutationOptions<ApiResponse<{ jobId: string }>, Error, unknown>
) {
  const apiClient = useApiClient();

  return useMutation({
    mutationFn: (data: unknown) => apiClient.post(API_ENDPOINTS.ENRICH, data),
    ...options,
  });
}

export function useCancelJobMutation(
  options?: UseMutationOptions<ApiResponse<unknown>, Error, string>
) {
  const apiClient = useApiClient();

  return useMutation({
    mutationFn: (jobId: string) => apiClient.post(API_ENDPOINTS.CANCEL_JOB(jobId)),
    ...options,
  });
}

export function useUploadSignedUrlMutation(
  options?: UseMutationOptions<ApiResponse<{ jobId: string }>, Error, unknown>
) {
  const apiClient = useApiClient();

  return useMutation({
    mutationFn: (data: unknown) => apiClient.post(API_ENDPOINTS.UPLOAD_SIGNED_URL, data),
    ...options,
  });
}
