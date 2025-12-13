import axios, { AxiosError } from 'axios';
import type { InitiateLoginRequest, InitiateLoginResponse } from '../types';

interface ApiErrorResponse {
  message?: string;
  error?: string;
}

function extractErrorMessage(error: AxiosError<ApiErrorResponse>): string {
  return (
    error.response?.data?.message ||
    error.response?.data?.error ||
    error.message ||
    'Failed to initiate login'
  );
}

class AuthApiClient {
  private baseURL: string;

  constructor() {
    this.baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  }

  async initiateLogin(email: string): Promise<InitiateLoginResponse> {
    try {
      const response = await axios.post<InitiateLoginResponse>(
        `${this.baseURL}/auth/sessions`,
        { email } as InitiateLoginRequest,
        { headers: { 'Content-Type': 'application/json' }, withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error));
      }
      throw error;
    }
  }
}

export const authApi = new AuthApiClient();
