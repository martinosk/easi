import { httpClient, fetchAllPaginated } from '../../../api/core';
import type {
  Component,
  ComponentId,
  CreateComponentRequest,
} from '../../../api/types';

export const componentsApi = {
  async getAll(): Promise<Component[]> {
    return fetchAllPaginated<Component>('/api/v1/components');
  },

  async getById(id: ComponentId): Promise<Component> {
    const response = await httpClient.get<Component>(`/api/v1/components/${id}`);
    return response.data;
  },

  async create(request: CreateComponentRequest): Promise<Component> {
    const response = await httpClient.post<Component>('/api/v1/components', request);
    return response.data;
  },

  async update(id: ComponentId, request: CreateComponentRequest): Promise<Component> {
    const response = await httpClient.put<Component>(`/api/v1/components/${id}`, request);
    return response.data;
  },

  async delete(id: ComponentId): Promise<void> {
    await httpClient.delete(`/api/v1/components/${id}`);
  },
};

export default componentsApi;
