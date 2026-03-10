// src/api/client.ts

import {
  decodeSourceList,
  decodeSourceDetail,
  decodeJobProgress,
  decodeRowsProgress,
  decodeEnrichmentResults,
} from './decoder';
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
} from './types';

export class ApiClient {
  private baseUrl = '/api/v1';

  private getToken(): string {
    return 'mock-token';
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const headers = new Headers(options?.headers);
    if (!headers.has('Content-Type') && (!options?.method || options.method === 'GET' || options.method === 'POST')) {
      headers.set('Content-Type', 'application/json');
    }

    headers.set('Authorization', `Bearer ${this.getToken()}`);

    const config: RequestInit = { ...options, headers };
    const response = await fetch(url, config);

    if (!response.ok) {
      let errorMessage = response.statusText;
      try {
        const errorData = await response.json();
        if (errorData.message) errorMessage = errorData.message;
      } catch { /* fallback to statusText */ }
      throw new Error(errorMessage);
    }

    const text = await response.text();
    if (!text) return {} as T;
    return JSON.parse(text) as T;
  }

  public async getSources(offset = 0, limit = 50): Promise<SourceListResponse> {
    const data = await this.request<unknown>(`/sources?offset=${offset}&limit=${limit}`);
    return decodeSourceList(data);
  }

  public async getSource(sourceId: string): Promise<SourceDetail> {
    const data = await this.request<unknown>(`/sources/${sourceId}`);
    return decodeSourceDetail(data);
  }

  public async getSourceData(sourceId: string): Promise<import('./types').SourceDataResponse> {
    // Assuming simple JSON decode is fine here or we can define a decoder. We'll cast for now.
    return this.request<import('./types').SourceDataResponse>(`/sources/${sourceId}/data`);
  }

  public async enrich(sourceId: string, req: EnrichRequest): Promise<EnrichResponse> {
    return this.request<EnrichResponse>(`/sources/${sourceId}/enrich`, {
      method: 'POST',
      body: JSON.stringify(req),
    });
  }

  public async getJobProgress(jobId: string): Promise<JobProgressResponse> {
    const data = await this.request<unknown>(`/jobs/${jobId}/progress`);
    return decodeJobProgress(data);
  }

  public async getJobRows(
    jobId: string,
    offset = 0,
    limit = 50,
    stage = 'all',
    sort = 'updated_at_desc'
  ): Promise<RowsProgressResponse> {
    const data = await this.request<unknown>(
      `/jobs/${jobId}/rows?offset=${offset}&limit=${limit}&stage=${stage}&sort=${sort}`
    );
    return decodeRowsProgress(data);
  }

  public async getJobResults(jobId: string, start = 0, limit = 0): Promise<EnrichmentResult[]> {
    const data = await this.request<unknown>(`/jobs/${jobId}/results?start=${start}&limit=${limit}`);
    return decodeEnrichmentResults(data);
  }

  public async cancelJob(jobId: string): Promise<{ message: string }> {
    return this.request<{ message: string }>(`/jobs/${jobId}/cancel`, { method: 'POST' });
  }

  public async getSignedUrl(req: SignedURLRequest): Promise<SignedURLResponse> {
    return this.request<SignedURLResponse>('/enrichment-signed-url', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  }

  public async uploadFile(url: string, file: File): Promise<void> {
    const response = await fetch(url, {
      method: 'PUT',
      body: file,
      headers: { 'Content-Type': file.type || 'text/csv' },
    });
    if (!response.ok) throw new Error(`File upload failed: ${response.statusText}`);
  }

  public async selectKey(req: SelectKeyRequest): Promise<SelectKeyResponse> {
    return this.request<SelectKeyResponse>('/select-key', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  }
}
