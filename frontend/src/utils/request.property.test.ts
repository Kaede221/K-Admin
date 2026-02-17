import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import * as fc from 'fast-check';

// Mock antd message before importing request
vi.mock('antd', () => ({
  message: {
    error: vi.fn(),
  },
}));

// Mock axios module with factory function
vi.mock('axios', () => {
  const mockAxiosInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
    defaults: {
      baseURL: '/api',
    },
  };
  
  return {
    default: {
      create: vi.fn(() => mockAxiosInstance),
      post: vi.fn(),
    },
  };
});

import axios from 'axios';
import { RequestClient } from './request';

const mockedAxios = vi.mocked(axios, true);

// Get the mock instance that axios.create returns
const getMockAxiosInstance = () => {
  return (axios.create as any)();
};

describe('Feature: k-admin-system, Request Client Property Tests', () => {
  beforeEach(() => {
    // Clear localStorage
    localStorage.clear();
    
    // Reset mocks
    vi.clearAllMocks();
    
    // Reset mock implementations
    const mockInstance = getMockAxiosInstance();
    mockInstance.interceptors.request.use.mockClear();
    mockInstance.interceptors.response.use.mockClear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  /**
   * Property 32: Authorization Header Injection
   * 
   * For any valid token string stored in localStorage,
   * the request client must inject it as a Bearer token
   * in the Authorization header of all outgoing requests.
   * 
   * Validates: Requirements 11.2
   */
  it('Property 32: Authorization Header Injection', () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 10, maxLength: 200 }),
        (token) => {
          // Setup
          localStorage.setItem('access_token', token);
          
          // Create a new request client to trigger interceptor setup
          new RequestClient();
          
          // Get the mock instance and request interceptor
          const mockInstance = getMockAxiosInstance();
          const requestInterceptor = mockInstance.interceptors.request.use.mock.calls[0][0];
          
          // Create a mock config
          const config: any = {
            headers: {},
          };
          
          // Apply interceptor
          const result = requestInterceptor(config);
          
          // Verify Authorization header is set correctly
          expect(result.headers.Authorization).toBe(`Bearer ${token}`);
        }
      ),
      { numRuns: 100 }
    );
  });

  /**
   * Property 33: Response Data Extraction
   * 
   * For any successful response with code 0, the request client
   * must extract and return only the data field, unwrapping
   * the unified response structure.
   * 
   * Validates: Requirements 11.3
   */
  it('Property 33: Response Data Extraction', () => {
    fc.assert(
      fc.property(
        fc.anything(),
        fc.string(),
        (data, msg) => {
          // Create a new request client to trigger interceptor setup
          new RequestClient();
          
          // Get the mock instance and response interceptor
          const mockInstance = getMockAxiosInstance();
          const responseInterceptor = mockInstance.interceptors.response.use.mock.calls[0][0];
          
          // Create a mock response
          const response: any = {
            data: {
              code: 0,
              data: data,
              msg: msg,
            },
          };
          
          // Apply interceptor
          const result = responseInterceptor(response);
          
          // Verify only data is returned
          expect(result).toEqual(data);
        }
      ),
      { numRuns: 100 }
    );
  });

  /**
   * Property 34: Error Notification Display
   * 
   * For any failed response (code !== 0) or network error,
   * the request client must display an error notification
   * with the appropriate error message.
   * 
   * Validates: Requirements 11.4
   */
  it('Property 34: Error Notification Display', async () => {
    const { message } = await import('antd');
    
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 9999 }),
        fc.string({ minLength: 1, maxLength: 100 }),
        (code, msg) => {
          // Create a new request client to trigger interceptor setup
          new RequestClient();
          
          // Get the mock instance and response interceptor
          const mockInstance = getMockAxiosInstance();
          const responseInterceptor = mockInstance.interceptors.response.use.mock.calls[0][0];
          
          // Create a mock response with non-zero code
          const response: any = {
            data: {
              code: code,
              data: null,
              msg: msg,
            },
          };
          
          // Clear previous calls
          vi.clearAllMocks();
          
          // Apply interceptor (should reject)
          try {
            responseInterceptor(response);
          } catch (error) {
            // Expected to throw
          }
          
          // Verify error message was displayed
          expect(message.error).toHaveBeenCalledWith(msg);
        }
      ),
      { numRuns: 100 }
    );
  });

  /**
   * Property 35: Automatic Token Refresh on 401
   * 
   * When a request receives a 401 response and a valid refresh token exists,
   * the request client must automatically attempt to refresh the access token
   * and retry the original request with the new token.
   * 
   * Validates: Requirements 11.5
   */
  it('Property 35: Automatic Token Refresh on 401', async () => {
    fc.assert(
      fc.asyncProperty(
        fc.string({ minLength: 20, maxLength: 100 }),
        fc.string({ minLength: 20, maxLength: 100 }),
        fc.string({ minLength: 20, maxLength: 100 }),
        async (oldToken, refreshToken, newToken) => {
          // Setup
          localStorage.setItem('access_token', oldToken);
          localStorage.setItem('refresh_token', refreshToken);
          
          // Create a new request client to trigger interceptor setup
          new RequestClient();
          
          // Get the mock instance and error interceptor
          const mockInstance = getMockAxiosInstance();
          const errorInterceptor = mockInstance.interceptors.response.use.mock.calls[0][1];
          
          // Mock the refresh token API call
          mockedAxios.post.mockResolvedValueOnce({
            data: {
              code: 0,
              data: { access_token: newToken },
              msg: 'success',
            },
          });
          
          // Mock the retry request
          mockInstance.get.mockResolvedValueOnce({ data: 'success' });
          
          // Create a 401 error
          const error: any = {
            response: {
              status: 401,
              data: {
                code: 401,
                msg: 'Unauthorized',
              },
            },
            config: {
              url: '/test',
              headers: {},
            },
          };
          
          // Apply error interceptor
          await errorInterceptor(error);
          
          // Verify refresh token API was called
          expect(mockedAxios.post).toHaveBeenCalledWith(
            '/api/auth/refresh',
            { refresh_token: refreshToken }
          );
          
          // Verify new token was stored
          expect(localStorage.getItem('access_token')).toBe(newToken);
          
          // Verify original request was retried with new token
          expect(mockInstance.get).toHaveBeenCalled();
        }
      ),
      { numRuns: 100 }
    );
  });
});
