import axios from 'axios';
import type { AxiosError, AxiosRequestConfig } from 'axios';
import type { Invitation, CreateInvitationRequest, UpdateInvitationRequest, InvitationsListResponse } from '../types';

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

class InvitationApiClient {
  private baseURL: string;

  constructor() {
    this.baseURL = import.meta.env.VITE_API_URL ?? '';
  }

  private async request<T>(config: AxiosRequestConfig, errorMessage: string): Promise<T> {
    try {
      const response = await axios.request<T>({
        ...config,
        baseURL: this.baseURL,
        withCredentials: true,
      });
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, errorMessage));
      }
      throw error;
    }
  }

  async createInvitation(request: CreateInvitationRequest): Promise<Invitation> {
    return this.request<Invitation>({
      method: 'POST',
      url: '/api/v1/invitations',
      data: request,
      headers: { 'Content-Type': 'application/json' },
    }, 'Failed to create invitation');
  }

  async listInvitations(limit: number = 50, after?: string): Promise<InvitationsListResponse> {
    const params = new URLSearchParams();
    params.append('limit', limit.toString());
    if (after) params.append('after', after);

    return this.request<InvitationsListResponse>({
      method: 'GET',
      url: `/api/v1/invitations?${params.toString()}`,
    }, 'Failed to fetch invitations');
  }

  async getInvitation(id: string): Promise<Invitation> {
    return this.request<Invitation>({
      method: 'GET',
      url: `/api/v1/invitations/${id}`,
    }, 'Failed to fetch invitation');
  }

  async updateInvitation(id: string, request: UpdateInvitationRequest): Promise<Invitation> {
    return this.request<Invitation>({
      method: 'PATCH',
      url: `/api/v1/invitations/${id}`,
      data: request,
      headers: { 'Content-Type': 'application/json' },
    }, 'Failed to update invitation');
  }

  async revokeInvitation(id: string): Promise<Invitation> {
    return this.updateInvitation(id, { status: 'revoked' });
  }
}

export const invitationApi = new InvitationApiClient();
