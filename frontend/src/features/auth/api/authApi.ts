import axios, { AxiosError } from 'axios';
import type { InitiateLoginRequest, InitiateLoginResponse } from '../types';

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
        {
          headers: {
            'Content-Type': 'application/json',
          },
          withCredentials: true,
        }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError<{ message?: string; error?: string }>;
        const errorMessage =
          axiosError.response?.data?.message ||
          axiosError.response?.data?.error ||
          axiosError.message ||
          'Failed to initiate login';
        throw new Error(errorMessage);
      }
      throw error;
    }
  }
}

export const authApi = new AuthApiClient();
