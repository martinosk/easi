import { httpClient, fetchAllPaginated } from '../../../api/core';
import { followLink } from '../../../utils/hateoas';
import type {
  Component,
  ComponentId,
  CreateComponentRequest,
  AddComponentExpertRequest,
  Expert,
} from '../../../api/types';

export const componentsApi = {
  async getAll(): Promise<Component[]> {
    return fetchAllPaginated<Component>('/api/v1/components');
  },

  async getById(id: ComponentId): Promise<Component> {
    const response = await httpClient.get<Component>(`/api/v1/components/${id}`);
    return response.data;
  },

  async getBySelfLink(component: Component): Promise<Component> {
    const response = await httpClient.get<Component>(followLink(component, 'self'));
    return response.data;
  },

  async create(request: CreateComponentRequest): Promise<Component> {
    const response = await httpClient.post<Component>('/api/v1/components', request);
    return response.data;
  },

  async update(component: Component, request: CreateComponentRequest): Promise<Component> {
    const response = await httpClient.put<Component>(followLink(component, 'edit'), request);
    return response.data;
  },

  async delete(component: Component): Promise<void> {
    await httpClient.delete(followLink(component, 'delete'));
  },

  async addExpert(id: ComponentId, request: AddComponentExpertRequest): Promise<void> {
    await httpClient.post(`/api/v1/components/${id}/experts`, request);
  },

  async removeExpert(expert: Expert): Promise<void> {
    await httpClient.delete(followLink(expert, 'x-remove'));
  },

  async getExpertRoles(): Promise<string[]> {
    const response = await httpClient.get<{ roles: string[] }>('/api/v1/components/expert-roles');
    return response.data.roles;
  },
};

export default componentsApi;
