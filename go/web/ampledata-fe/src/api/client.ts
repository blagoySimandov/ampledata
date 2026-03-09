// src/api/client.ts

import {
  decodeJobList,
  decodeJobProgress,
  decodeRowsProgress,
  decodeEnrichmentResults,
} from './decoder';
import type {
  JobListResponse,
  JobProgressResponse,
  RowsProgressResponse,
  EnrichmentResult,
} from './types';

export class ApiClient {
  private baseUrl = '/api/v1';

  // Mock auth token as requested
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

    const config: RequestInit = {
      ...options,
      headers,
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      let errorMessage = response.statusText;
      try {
        const errorData = await response.json();
        if (errorData.message) {
          errorMessage = errorData.message;
        }
      } catch (e) {
        // Fallback to text if JSON parsing fails
      }
      throw new Error(errorMessage);
    }

    // Some endpoints might return empty bodies for 200/204
    const text = await response.text();
    if (!text) return {} as T;
    
    return JSON.parse(text) as T;
  }

  public async getJobs(offset: number = 0, limit: number = 50): Promise<JobListResponse> {
    const data = await this.request<unknown>(`/jobs?offset=${offset}&limit=${limit}`);
    return decodeJobList(data);
  }

  public async getJobProgress(jobId: string): Promise<JobProgressResponse> {
    const data = await this.request<unknown>(`/jobs/${jobId}/progress`);
    return decodeJobProgress(data);
  }

  public async getJobRows(
    jobId: string,
    offset: number = 0,
    limit: number = 50,
    stage: string = 'all',
    sort: string = 'updated_at_desc'
  ): Promise<RowsProgressResponse> {
    const data = await this.request<unknown>(
      `/jobs/${jobId}/rows?offset=${offset}&limit=${limit}&stage=${stage}&sort=${sort}`
    );
    return decodeRowsProgress(data);
  }

  public async getJobResults(
    jobId: string,
    start: number = 0,
    limit: number = 0
  ): Promise<EnrichmentResult[]> {
    const data = await this.request<unknown>(`/jobs/${jobId}/results?start=${start}&limit=${limit}`);
    return decodeEnrichmentResults(data);
  }

  public async cancelJob(jobId: string): Promise<{ message: string }> {
    return this.request<{ message: string }>(`/jobs/${jobId}/cancel`, {
      method: 'POST',
    });
  }
}
