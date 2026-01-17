export const API_ENDPOINTS = {
	SIGNED_URL: "/api/v1/enrichment-signed-url",
	JOBS: "/api/v1/jobs",
	JOB_START: (jobId: string) => `/api/v1/jobs/${jobId}/start`,
	JOB_PROGRESS: (jobId: string) => `/api/v1/jobs/${jobId}/progress`,
	JOB_RESULTS: (jobId: string) => `/api/v1/jobs/${jobId}/results`,
	JOB_CANCEL: (jobId: string) => `/api/v1/jobs/${jobId}/cancel`,
} as const;

export const BASE_URL =
	process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
