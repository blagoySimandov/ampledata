// src/api/client.ts
import {
  decodeSourceList,
  decodeSourceDetail,
  decodeJobProgress,
  decodeRowsProgress,
  decodeEnrichmentResults,
} from "./decoder";
import type {
  SourceListResponse,
  SourceDetail,
  JobProgressResponse,
  RowsProgressResponse,
  EnrichmentResult,
  SignedURLRequest,
  SignedURLResponse,
  SelectKeyRequest,
  SelectKeyResponse,
  EnrichRequest,
  EnrichResponse,
} from "./types";
import { ENDPOINTS } from "./endpoints";

export class ApiClient {
  private baseUrl = "/api/v1";

  private getToken(): string {
    return "mock-token";
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit,
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const headers = new Headers(options?.headers);
    headers.set("Authorization", `Bearer ${this.getToken()}`);

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
    const data = await this.request<unknown>(
      `${ENDPOINTS.SOURCES_LIST}?offset=${offset}&limit=${limit}`,
    );
    return decodeSourceList(data);
  }

  public async getSource(sourceId: string): Promise<SourceDetail> {
    const data = await this.request<unknown>(
      ENDPOINTS.SOURCES_DETAIL(sourceId),
    );
    return decodeSourceDetail(data);
  }

  public async getSourceData(
    sourceId: string,
  ): Promise<import("./types").SourceDataResponse> {
    return this.request<import("./types").SourceDataResponse>(
      ENDPOINTS.SOURCES_DATA(sourceId),
    );
  }

  public async enrich(
    sourceId: string,
    req: EnrichRequest,
  ): Promise<EnrichResponse> {
    return this.request<EnrichResponse>(ENDPOINTS.SOURCES_ENRICH(sourceId), {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  public async getJobProgress(jobId: string): Promise<JobProgressResponse> {
    const data = await this.request<unknown>(ENDPOINTS.JOBS_PROGRESS(jobId));
    return decodeJobProgress(data);
  }

  public async getJobRows(
    jobId: string,
    offset = 0,
    limit = 50,
    stage = "all",
    sort = "updated_at_desc",
  ): Promise<RowsProgressResponse> {
    const data = await this.request<unknown>(
      `${ENDPOINTS.JOBS_ROWS(jobId)}?offset=${offset}&limit=${limit}&stage=${stage}&sort=${sort}`,
    );
    return decodeRowsProgress(data);
  }

  public async getJobResults(
    jobId: string,
    start = 0,
    limit = 0,
  ): Promise<EnrichmentResult[]> {
    const data = await this.request<unknown>(
      `${ENDPOINTS.JOBS_RESULTS(jobId)}?start=${start}&limit=${limit}`,
    );
    return decodeEnrichmentResults(data);
  }

  public async cancelJob(jobId: string): Promise<{ message: string }> {
    return this.request<{ message: string }>(ENDPOINTS.JOBS_CANCEL(jobId), {
      method: "POST",
    });
  }

  public async getSignedUrl(req: SignedURLRequest): Promise<SignedURLResponse> {
    return this.request<SignedURLResponse>(ENDPOINTS.ENRICHMENT_SIGNED_URL, {
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
    return this.request<SelectKeyResponse>(ENDPOINTS.SELECT_KEY, {
      method: "POST",
      body: JSON.stringify(req),
    });
  }
}
