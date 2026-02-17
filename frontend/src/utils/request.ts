import axios, {
  type AxiosInstance,
  type AxiosRequestConfig,
  type AxiosResponse,
  type AxiosError,
} from "axios";
import { getToken, setToken, getRefreshToken, removeToken, removeRefreshToken } from "./storage";
import { navigateTo } from "./navigation";

export interface UnifiedResponse<T = any> {
  code: number;
  data: T;
  msg: string;
}

class RequestClient {
  private axiosInstance: AxiosInstance;
  private isRefreshing = false;
  private refreshSubscribers: Array<(token: string) => void> = [];

  constructor(config?: AxiosRequestConfig) {
    this.axiosInstance = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || "/api",
      timeout: 30000,
      ...config,
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor
    this.axiosInstance.interceptors.request.use(
      (config) => {
        // Add Authorization header
        const token = getToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      },
    );

    // Response interceptor
    this.axiosInstance.interceptors.response.use(
      (response: AxiosResponse<UnifiedResponse>) => {
        const { code, data, msg } = response.data;

        // Success response
        if (code === 0) {
          return data;
        }

        // Business error - don't show message here, let caller handle it
        return Promise.reject(new Error(msg || "Request failed"));
      },
      async (error: AxiosError<UnifiedResponse>) => {
        // Handle 401 - Token expired
        if (error.response?.status === 401) {
          return this.handleTokenRefresh(error);
        }

        // Handle other errors - don't show message here, let caller handle it
        const msg =
          error.response?.data?.msg || error.message || "Network error";
        return Promise.reject(new Error(msg));
      },
    );
  }

  private async handleTokenRefresh(error: AxiosError): Promise<any> {
    const originalRequest = error.config;
    if (!originalRequest) {
      return Promise.reject(error);
    }

    // If already refreshing, queue the request
    if (this.isRefreshing) {
      return new Promise((resolve) => {
        this.refreshSubscribers.push((token: string) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          resolve(this.axiosInstance(originalRequest));
        });
      });
    }

    this.isRefreshing = true;

    try {
      const refreshToken = getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token");
      }

      // Call refresh token API
      const response = await axios.post<
        UnifiedResponse<{ access_token: string }>
      >(`${this.axiosInstance.defaults.baseURL}/auth/refresh`, {
        refresh_token: refreshToken,
      });

      if (response.data.code === 0) {
        const newToken = response.data.data.access_token;
        setToken(newToken);

        // Retry all queued requests
        this.refreshSubscribers.forEach((callback) => callback(newToken));
        this.refreshSubscribers = [];

        // Retry original request
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return this.axiosInstance(originalRequest);
      } else {
        throw new Error("Token refresh failed");
      }
    } catch (refreshError) {
      // Refresh failed, redirect to login
      removeToken();
      removeRefreshToken();
      navigateTo("/login");
      return Promise.reject(refreshError);
    } finally {
      this.isRefreshing = false;
    }
  }

  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.axiosInstance.get(url, config);
  }

  post<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.axiosInstance.post(url, data, config);
  }

  put<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.axiosInstance.put(url, data, config);
  }

  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.axiosInstance.delete(url, config);
  }
}

// Export singleton instance and class
export const request = new RequestClient();
export { RequestClient };
export default request;
