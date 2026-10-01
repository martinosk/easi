import { httpClient } from '../../../api/core/httpClient';
import type { HomeResponse } from '../types';

export const homeApi = {
  async getHome(): Promise<HomeResponse> {
    const response = await httpClient.get<HomeResponse>('/api/v1/home');
    return response.data;
  },
};
