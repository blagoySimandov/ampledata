import {
  decodeSourceList,
  decodeSourceDetail,
  decodeJobProgress,
  decodeRowsProgress,
} from "./decoder";
import type {
  SourceListResponse,
  SourceDetail,
  JobProgressResponse,
  RowsProgressResponse,
  SignedURLRequest,
  SignedURLResponse,
  SelectKeyRequest,
  SelectKeyResponse,
  EnrichRequest,
  EnrichResponse,
  TierResponse,
  SubscriptionStatusResponse,
  CreateSubscriptionRequest,
  CreateCheckoutResponse,
  PortalSessionResponse,
  UserResponse,
  SourceDataResponse,
} from "./types";
import { ENDPOINTS } from "./endpoints";

export class ApiClient {
  private baseUrl = "/api/v1";
  private getToken: () => Promise<string | null>;

  constructor(getToken: () => Promise<string | null>) {
    this.getToken = getToken;
  }

  private buildUrl(
    endpoint: string,
    params?: Record<string, string | number | undefined>,
  ): string {
    const url = new URL(endpoint, "http://localhost");
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          url.searchParams.append(key, String(value));
        }
      });
    }
    return url.pathname + url.search;
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit,
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const token = await this.getToken();
    const headers = new Headers(options?.headers);
    if (token) headers.set("Authorization", `Bearer ${token}`);

    const config: RequestInit = {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...Object.fromEntries(headers),
      },
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || response.statusText);
    }

    const text = await response.text();
    return text ? (JSON.parse(text) as T) : ({} as T);
  }

  public async getSources(offset = 0, limit = 50): Promise<SourceListResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.SOURCES_LIST, { offset, limit });
    const data = await this.request<SourceListResponse>(endpoint);
    return decodeSourceList(data);
  }

  public async getSource(sourceId: string): Promise<SourceDetail> {
    const endpoint = this.buildUrl(ENDPOINTS.SOURCES_DETAIL(sourceId));
    const data = await this.request<SourceDetail>(endpoint);
    return decodeSourceDetail(data);
  }

  public async getSourceData(sourceId: string): Promise<SourceDataResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.SOURCES_DATA(sourceId));
    return this.request<SourceDataResponse>(endpoint);
  }

  public async enrich(
    sourceId: string,
    req: EnrichRequest,
  ): Promise<EnrichResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.SOURCES_ENRICH(sourceId));
    return this.request<EnrichResponse>(endpoint, {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  public async getJobProgress(jobId: string): Promise<JobProgressResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.JOBS_PROGRESS(jobId));
    const data = await this.request<JobProgressResponse>(endpoint);
    return decodeJobProgress(data);
  }

  public async getJobRows(
    jobId: string,
    offset = 0,
    limit = 50,
    stage = "all",
    sort = "updated_at_desc",
  ): Promise<RowsProgressResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.JOBS_ROWS(jobId), {
      offset,
      limit,
      stage,
      sort,
    });
    const data = await this.request<RowsProgressResponse>(endpoint);
    return decodeRowsProgress(data);
  }

  public async cancelJob(jobId: string): Promise<{ message: string }> {
    const endpoint = this.buildUrl(ENDPOINTS.JOBS_CANCEL(jobId));
    return this.request<{ message: string }>(endpoint, {
      method: "POST",
    });
  }

  public async getSignedUrl(req: SignedURLRequest): Promise<SignedURLResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.ENRICHMENT_SIGNED_URL);
    return this.request<SignedURLResponse>(endpoint, {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  public async uploadFile(url: string, file: File): Promise<void> {
    const response = await fetch(url, {
      method: "PUT",
      body: file,
      headers: { "Content-Type": file.type || "text/csv" },
    });
    if (!response.ok)
      throw new Error(`File upload failed: ${response.statusText}`);
  }

  public async selectKey(req: SelectKeyRequest): Promise<SelectKeyResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.SELECT_KEY);
    return this.request<SelectKeyResponse>(endpoint, {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  public async listTiers(): Promise<TierResponse[]> {
    return this.request<TierResponse[]>(ENDPOINTS.TIERS);
  }

  public async getSubscriptionStatus(): Promise<SubscriptionStatusResponse> {
    return this.request<SubscriptionStatusResponse>(ENDPOINTS.SUBSCRIPTION);
  }

  public async createSubscriptionCheckout(
    req: CreateSubscriptionRequest,
  ): Promise<CreateCheckoutResponse> {
    return this.request<CreateCheckoutResponse>(ENDPOINTS.SUBSCRIBE, {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  public async cancelSubscription(): Promise<{ message: string }> {
    return this.request<{ message: string }>(ENDPOINTS.SUBSCRIPTION_CANCEL, {
      method: "POST",
    });
  }

  public async createPortalSession(
    returnUrl: string,
  ): Promise<PortalSessionResponse> {
    const endpoint = this.buildUrl(ENDPOINTS.SUBSCRIPTION_PORTAL, {
      return_url: returnUrl,
    });
    return this.request<PortalSessionResponse>(endpoint, { method: "POST" });
  }

  public async getMe(): Promise<UserResponse> {
    return this.request<UserResponse>(ENDPOINTS.ME);
  }
}
