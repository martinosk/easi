import { httpClient } from '../../../api/core/httpClient';
import { fetchAllPaginated } from '../../../api/core/pagination';
import type { User, UpdateUserRequest } from '../types';

export const userApi = {
  async getAll(statusFilter?: string, roleFilter?: string): Promise<User[]> {
    return fetchAllPaginated<User>('/api/v1/users', {
      queryParams: {
        status: statusFilter,
        role: roleFilter,
      },
    });
  },

  async getById(id: string): Promise<User> {
    const response = await httpClient.get<User>(`/api/v1/users/${id}`);
    return response.data;
  },

  async update(id: string, request: UpdateUserRequest): Promise<User> {
    const response = await httpClient.patch<User>(`/api/v1/users/${id}`, request);
    return response.data;
  },
};
