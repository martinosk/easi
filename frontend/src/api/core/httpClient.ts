import axios, { type AxiosError, type AxiosInstance } from 'axios';
import { useUserStore } from '../../store/userStore';
import { ApiError } from '../types';

let isRedirectingToLogin = false;

function extractResponseMessage(data: unknown): string | null {
  if (!data || typeof data !== 'object') return null;

  const errorData = data as { message?: string; error?: string; details?: Record<string, string> };
  if (errorData.message) return errorData.message;
  if (errorData.error) return errorData.error;

  const detailMessages = errorData.details ? Object.values(errorData.details).filter(Boolean) : [];
  return detailMessages.length > 0 ? detailMessages.join(', ') : 'An error occurred';
}

function extractErrorMessage(error: AxiosError): string {
  const responseMessage = extractResponseMessage(error.response?.data);
  return responseMessage ?? error.message ?? 'An unknown error occurred';
}

function createHttpClient(baseURL: string = import.meta.env.VITE_API_URL || 'http://localhost:8080'): AxiosInstance {
  const client = axios.create({
    baseURL,
    headers: {
      'Content-Type': 'application/json',
    },
    withCredentials: true,
  });

  client.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      const statusCode = error.response?.status || 500;

      if (statusCode === 401 && !window.location.pathname.endsWith('/login') && !isRedirectingToLogin) {
        isRedirectingToLogin = true;
        useUserStore.getState().clearUser();
        const basePath = import.meta.env.BASE_URL || '/';
        const returnUrl = encodeURIComponent(window.location.pathname + window.location.search);
        window.location.href = `${basePath}login?returnUrl=${returnUrl}`;
        return Promise.reject(error);
      }

      const message = extractErrorMessage(error);
      throw new ApiError(message, statusCode, error.response?.data);
    }
  );

  return client;
}

export const httpClient = createHttpClient();

export function resetLoginRedirectFlag(): void {
  isRedirectingToLogin = false;
}
