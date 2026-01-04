export const WORKOS_CLIENT_ID = import.meta.env.VITE_WORKOS_CLIENT_ID || '';
export const WORKOS_API_HOSTNAME = import.meta.env.VITE_WORKOS_API_HOSTNAME || 'api.workos.com';
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const AUTH_ROUTES = {
  LOGIN: '/login',
  CALLBACK: '/auth/callback',
} as const;

export const APP_ROUTES = {
  HOME: '/',
  ENRICHMENT: '/enrichment',
} as const;

export const API_ENDPOINTS = {
  ENRICH: '/api/v1/enrich',
  JOB_PROGRESS: (jobId: string) => `/api/v1/jobs/${jobId}/progress`,
  CANCEL_JOB: (jobId: string) => `/api/v1/jobs/${jobId}/cancel`,
  JOB_RESULTS: (jobId: string) => `/api/v1/jobs/${jobId}/results`,
  UPLOAD_SIGNED_URL: '/api/v1/enrichment-signed-url',
} as const;

export const HTTP_HEADERS = {
  AUTHORIZATION: 'Authorization',
  CONTENT_TYPE: 'Content-Type',
} as const;

export const CONTENT_TYPES = {
  JSON: 'application/json',
} as const;

export const AUTH_TOKENS = {
  BEARER_PREFIX: 'Bearer ',
} as const;

export const HTTP_STATUS = {
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  INTERNAL_SERVER_ERROR: 500,
} as const;

export const UI_MESSAGES = {
  LOADING: 'Loading...',
  UNAUTHORIZED: 'Unauthorized',
  SIGN_IN: 'Sign In',
  SIGN_OUT: 'Sign Out',
  PROFILE: 'Profile',
  ACCOUNT: 'Account',
  SETTINGS: 'Settings',
} as const;

export const QUERY_KEYS = {
  JOB_PROGRESS: (jobId: string) => ['job', 'progress', jobId],
  JOB_RESULTS: (jobId: string) => ['job', 'results', jobId],
} as const;

export const DEFAULT_AVATAR = 'https://api.dicebear.com/7.x/initials/svg?seed=';

