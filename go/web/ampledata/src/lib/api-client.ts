import {
  API_BASE_URL,
  HTTP_HEADERS,
  CONTENT_TYPES,
  AUTH_TOKENS,
  HTTP_STATUS,
} from "@/constants";
import type { ApiResponse, ApiError } from "@/types";

interface RequestOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: unknown;
}

export class ApiClient {
  private baseUrl: string;
  private getAccessToken?: () => Promise<string | null>;

  constructor(
    baseUrl: string = API_BASE_URL,
    getAccessToken?: () => Promise<string | null>
  ) {
    this.baseUrl = baseUrl;
    this.getAccessToken = getAccessToken;
  }

  private async getAuthHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {
      [HTTP_HEADERS.CONTENT_TYPE]: CONTENT_TYPES.JSON,
    };

    if (this.getAccessToken) {
      const token = await this.getAccessToken();
      if (token) {
        headers[HTTP_HEADERS.AUTHORIZATION] = `${AUTH_TOKENS.BEARER_PREFIX}${token}`;
      }
    }

    return headers;
  }

  private async handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
    if (!response.ok) {
      const error: ApiError = {
        message: response.statusText,
        status: response.status,
      };

      if (response.status === HTTP_STATUS.UNAUTHORIZED) {
        error.message = "Authentication required";
      } else if (response.status === HTTP_STATUS.FORBIDDEN) {
        error.message = "Access forbidden";
      }

      try {
        const errorData = await response.json();
        error.message = errorData.message || error.message;
        error.code = errorData.code;
      } catch {
        // Response body is not JSON
      }

      return { error };
    }

    try {
      const data = await response.json();
      return { data };
    } catch {
      return { data: undefined as T };
    }
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    const headers = await this.getAuthHeaders();

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: "GET",
        headers,
      });

      return this.handleResponse<T>(response);
    } catch (error) {
      return {
        error: {
          message: error instanceof Error ? error.message : "Network error",
        },
      };
    }
  }

  async post<T>(endpoint: string, data?: unknown): Promise<ApiResponse<T>> {
    const headers = await this.getAuthHeaders();

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: "POST",
        headers,
        body: data ? JSON.stringify(data) : undefined,
      });

      return this.handleResponse<T>(response);
    } catch (error) {
      return {
        error: {
          message: error instanceof Error ? error.message : "Network error",
        },
      };
    }
  }

  async put<T>(endpoint: string, data?: unknown): Promise<ApiResponse<T>> {
    const headers = await this.getAuthHeaders();

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: "PUT",
        headers,
        body: data ? JSON.stringify(data) : undefined,
      });

      return this.handleResponse<T>(response);
    } catch (error) {
      return {
        error: {
          message: error instanceof Error ? error.message : "Network error",
        },
      };
    }
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    const headers = await this.getAuthHeaders();

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: "DELETE",
        headers,
      });

      return this.handleResponse<T>(response);
    } catch (error) {
      return {
        error: {
          message: error instanceof Error ? error.message : "Network error",
        },
      };
    }
  }
}

export const createApiClient = (
  getAccessToken?: () => Promise<string | null>
) => {
  return new ApiClient(API_BASE_URL, getAccessToken);
};
