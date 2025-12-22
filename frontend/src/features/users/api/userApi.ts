import axios, { AxiosError } from 'axios';
import type { User, UpdateUserRequest, UsersListResponse } from '../types';

interface ApiErrorResponse {
  message?: string;
  error?: string;
}

function extractErrorMessage(error: AxiosError<ApiErrorResponse>, fallback: string): string {
  return (
    error.response?.data?.message ||
    error.response?.data?.error ||
    error.message ||
    fallback
  );
}

class UserApiClient {
  private baseURL: string;

  constructor() {
    this.baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  }

  async listUsers(
    limit: number = 50,
    after?: string,
    statusFilter?: string,
    roleFilter?: string
  ): Promise<UsersListResponse> {
    try {
      const params = new URLSearchParams();
      params.append('limit', limit.toString());
      if (after) params.append('after', after);
      if (statusFilter) params.append('status', statusFilter);
      if (roleFilter) params.append('role', roleFilter);

      const response = await axios.get<UsersListResponse>(
        `${this.baseURL}/api/v1/users?${params.toString()}`,
        { withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to fetch users'));
      }
      throw error;
    }
  }

  async getUser(id: string): Promise<User> {
    try {
      const response = await axios.get<User>(
        `${this.baseURL}/api/v1/users/${id}`,
        { withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to fetch user'));
      }
      throw error;
    }
  }

  async updateUser(id: string, request: UpdateUserRequest): Promise<User> {
    try {
      const response = await axios.patch<User>(
        `${this.baseURL}/api/v1/users/${id}`,
        request,
        { headers: { 'Content-Type': 'application/json' }, withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to update user'));
      }
      throw error;
    }
  }
}

export const userApi = new UserApiClient();
