import axios, { AxiosError } from 'axios';
import type { Invitation, CreateInvitationRequest, InvitationsListResponse } from '../types';

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
    this.baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  }

  async createInvitation(request: CreateInvitationRequest): Promise<Invitation> {
    try {
      const response = await axios.post<Invitation>(
        `${this.baseURL}/api/v1/invitations`,
        request,
        { headers: { 'Content-Type': 'application/json' }, withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to create invitation'));
      }
      throw error;
    }
  }

  async listInvitations(limit: number = 50, after?: string): Promise<InvitationsListResponse> {
    try {
      const params = new URLSearchParams();
      params.append('limit', limit.toString());
      if (after) params.append('after', after);

      const response = await axios.get<InvitationsListResponse>(
        `${this.baseURL}/api/v1/invitations?${params.toString()}`,
        { withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to fetch invitations'));
      }
      throw error;
    }
  }

  async getInvitation(id: string): Promise<Invitation> {
    try {
      const response = await axios.get<Invitation>(
        `${this.baseURL}/api/v1/invitations/${id}`,
        { withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to fetch invitation'));
      }
      throw error;
    }
  }

  async revokeInvitation(id: string): Promise<Invitation> {
    try {
      const response = await axios.post<Invitation>(
        `${this.baseURL}/api/v1/invitations/${id}/revoke`,
        {},
        { headers: { 'Content-Type': 'application/json' }, withCredentials: true }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(extractErrorMessage(error, 'Failed to revoke invitation'));
      }
      throw error;
    }
  }
}

export const invitationApi = new InvitationApiClient();
