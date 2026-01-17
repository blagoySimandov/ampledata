import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@workos-inc/authkit-react";
import { API_ENDPOINTS, BASE_URL } from "./constants";

export type ColumnType = "string" | "number" | "boolean" | "date";

export interface ColumnMetadata {
	name: string;
	type: ColumnType;
	description?: string;
}

export interface SignedURLRequest {
	contentType: string;
	length: number;
}

export interface SignedURLResponse {
	url: string;
	jobId: string;
}

export interface StartJobRequest {
	key_column: string;
	columns_metadata: ColumnMetadata[];
	entity_type?: string;
}

export interface StartJobResponse {
	job_id: string;
	message: string;
}

export type JobStatus =
	| "PENDING"
	| "RUNNING"
	| "PAUSED"
	| "CANCELLED"
	| "COMPLETED";

export type RowStage =
	| "PENDING"
	| "SERP_FETCHED"
	| "DECISION_MADE"
	| "CRAWLED"
	| "ENRICHED"
	| "COMPLETED"
	| "FAILED"
	| "CANCELLED";

export interface JobSummary {
	job_id: string;
	status: JobStatus;
	total_rows: number;
	file_path: string;
	created_at: string;
	started_at?: string;
}

export interface JobListResponse {
	jobs: JobSummary[];
	total_count: number;
}

export interface FieldConfidenceInfo {
	score: number;
	reason: string;
}

export interface EnrichmentResult {
	key: string;
	extracted_data: Record<string, unknown>;
	confidence?: Record<string, FieldConfidenceInfo>;
	sources: string[];
	error?: string;
}

export type EnrichmentResultsResponse = EnrichmentResult[];

export interface JobProgressResponse {
	job_id: string;
	total_rows: number;
	rows_by_stage: Record<RowStage, number>;
	started_at: string;
	status: JobStatus;
}

export interface CancelJobResponse {
	message: string;
}

async function fetchAPI<T>(
	endpoint: string,
	options?: RequestInit,
	accessToken?: string | (() => Promise<string | undefined>)
): Promise<T> {
	const headers: Record<string, string> = {
		"Content-Type": "application/json",
		...(options?.headers as Record<string, string>),
	};

	let token: string | undefined;
	if (typeof accessToken === "function") {
		token = await accessToken();
	} else {
		token = accessToken;
	}

	if (token) {
		headers["Authorization"] = `Bearer ${token}`;
	}

	const response = await fetch(`${BASE_URL}${endpoint}`, {
		...options,
		headers,
	});

	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`API error: ${response.status} - ${errorText}`);
	}

	return response.json();
}

async function uploadFile(signedUrl: string, file: File): Promise<void> {
	const response = await fetch(signedUrl, {
		method: "PUT",
		headers: {
			"Content-Type": file.type,
		},
		body: file,
	});

	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`Upload error: ${response.status} - ${errorText}`);
	}
}

export function useRequestSignedURL() {
	const { getAccessToken } = useAuth();
	return useMutation({
		mutationFn: async (
			data: SignedURLRequest
		): Promise<SignedURLResponse> => {
			return fetchAPI<SignedURLResponse>(
				API_ENDPOINTS.SIGNED_URL,
				{
					method: "POST",
					body: JSON.stringify(data),
				},
				getAccessToken
			);
		},
	});
}

export function useUploadFile() {
	return useMutation({
		mutationFn: async ({
			signedUrl,
			file,
		}: {
			signedUrl: string;
			file: File;
		}): Promise<void> => {
			return uploadFile(signedUrl, file);
		},
	});
}

export function useListJobs(params?: { offset?: number; limit?: number }) {
	const { getAccessToken } = useAuth();
	return useQuery({
		queryKey: ["jobs", params],
		queryFn: async (): Promise<JobListResponse> => {
			const searchParams = new URLSearchParams();
			if (params?.offset !== undefined) {
				searchParams.append("offset", params.offset.toString());
			}
			if (params?.limit !== undefined) {
				searchParams.append("limit", params.limit.toString());
			}
			const queryString = searchParams.toString();
			const endpoint = queryString
				? `${API_ENDPOINTS.JOBS}?${queryString}`
				: API_ENDPOINTS.JOBS;
			return fetchAPI<JobListResponse>(
				endpoint,
				undefined,
				getAccessToken
			);
		},
	});
}

export function useStartJob() {
	const queryClient = useQueryClient();
	const { getAccessToken } = useAuth();
	return useMutation({
		mutationFn: async ({
			jobId,
			data,
		}: {
			jobId: string;
			data: StartJobRequest;
		}): Promise<StartJobResponse> => {
			return fetchAPI<StartJobResponse>(
				API_ENDPOINTS.JOB_START(jobId),
				{
					method: "POST",
					body: JSON.stringify(data),
				},
				getAccessToken
			);
		},
		onSuccess: (_, variables) => {
			queryClient.invalidateQueries({ queryKey: ["jobs"] });
			queryClient.invalidateQueries({
				queryKey: ["job-progress", variables.jobId],
			});
		},
	});
}

export function useJobProgress(
	jobId: string,
	options?: { enabled?: boolean; refetchInterval?: number }
) {
	const { getAccessToken } = useAuth();
	return useQuery({
		queryKey: ["job-progress", jobId],
		queryFn: async (): Promise<JobProgressResponse> => {
			return fetchAPI<JobProgressResponse>(
				API_ENDPOINTS.JOB_PROGRESS(jobId),
				undefined,
				getAccessToken
			);
		},
		enabled: options?.enabled !== false && !!jobId,
		refetchInterval: options?.refetchInterval,
	});
}

export function useJobResults(
	jobId: string,
	params?: { start?: number; limit?: number },
	options?: { enabled?: boolean }
) {
	const { getAccessToken } = useAuth();
	return useQuery({
		queryKey: ["job-results", jobId, params],
		queryFn: async (): Promise<EnrichmentResult[]> => {
			const searchParams = new URLSearchParams();
			if (params?.start !== undefined) {
				searchParams.append("start", params.start.toString());
			}
			if (params?.limit !== undefined) {
				searchParams.append("limit", params.limit.toString());
			}
			const queryString = searchParams.toString();
			const endpoint = queryString
				? `${API_ENDPOINTS.JOB_RESULTS(jobId)}?${queryString}`
				: API_ENDPOINTS.JOB_RESULTS(jobId);
			return fetchAPI<EnrichmentResult[]>(
				endpoint,
				undefined,
				getAccessToken
			);
		},
		enabled: options?.enabled !== false && !!jobId,
	});
}

export function useCancelJob() {
	const queryClient = useQueryClient();
	const { getAccessToken } = useAuth();
	return useMutation({
		mutationFn: async (jobId: string): Promise<CancelJobResponse> => {
			return fetchAPI<CancelJobResponse>(
				API_ENDPOINTS.JOB_CANCEL(jobId),
				{
					method: "POST",
				},
				getAccessToken
			);
		},
		onSuccess: (_, jobId) => {
			queryClient.invalidateQueries({ queryKey: ["jobs"] });
			queryClient.invalidateQueries({
				queryKey: ["job-progress", jobId],
			});
		},
	});
}
