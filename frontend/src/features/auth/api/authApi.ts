import { httpClient } from '../../../api/core';
import type { InitiateLoginRequest, InitiateLoginResponse, CurrentSessionResponse } from '../types';

export const authApi = {
  async initiateLogin(email: string, returnUrl?: string): Promise<InitiateLoginResponse> {
    const request: InitiateLoginRequest = { email };
    if (returnUrl) {
      request.returnUrl = returnUrl;
    }
    const response = await httpClient.post<InitiateLoginResponse>(
      '/api/v1/auth/sessions',
      request
    );
    return response.data;
  },

  async getCurrentSession(): Promise<CurrentSessionResponse> {
    const response = await httpClient.get<CurrentSessionResponse>('/api/v1/auth/sessions/current');
    return response.data;
  },

  async logout(): Promise<void> {
    await httpClient.delete('/api/v1/auth/sessions/current');
  },
};
