import { httpClient } from '../../../api/core/httpClient';
import type { EditGrant, CreateEditGrantRequest } from '../types';

export const editGrantApi = {
  async getMyGrants(): Promise<EditGrant[]> {
    const response = await httpClient.get<{ data: EditGrant[] }>('/api/v1/edit-grants');
    return response.data.data || [];
  },

  async getById(id: string): Promise<EditGrant> {
    const response = await httpClient.get<EditGrant>(`/api/v1/edit-grants/${id}`);
    return response.data;
  },

  async getForArtifact(artifactType: string, artifactId: string): Promise<EditGrant[]> {
    const response = await httpClient.get<{ data: EditGrant[] }>(
      `/api/v1/edit-grants/artifact/${artifactType}/${artifactId}`
    );
    return response.data.data || [];
  },

  async create(request: CreateEditGrantRequest): Promise<EditGrant> {
    const response = await httpClient.post<EditGrant>('/api/v1/edit-grants', {
      ...request,
      scope: request.scope ?? 'write',
    });
    return response.data;
  },

  async revoke(id: string): Promise<void> {
    await httpClient.delete(`/api/v1/edit-grants/${id}`);
  },
};
