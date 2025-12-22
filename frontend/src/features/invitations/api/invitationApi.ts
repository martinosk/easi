import { httpClient } from '../../../api/core/httpClient';
import { fetchAllPaginated } from '../../../api/core/pagination';
import type { Invitation, CreateInvitationRequest, UpdateInvitationRequest } from '../types';

export const invitationApi = {
  async getAll(): Promise<Invitation[]> {
    return fetchAllPaginated<Invitation>('/api/v1/invitations');
  },

  async getById(id: string): Promise<Invitation> {
    const response = await httpClient.get<Invitation>(`/api/v1/invitations/${id}`);
    return response.data;
  },

  async create(request: CreateInvitationRequest): Promise<Invitation> {
    const response = await httpClient.post<Invitation>('/api/v1/invitations', request);
    return response.data;
  },

  async update(id: string, request: UpdateInvitationRequest): Promise<Invitation> {
    const response = await httpClient.patch<Invitation>(`/api/v1/invitations/${id}`, request);
    return response.data;
  },

  async revoke(id: string): Promise<Invitation> {
    return invitationApi.update(id, { status: 'revoked' });
  },
};
