import { httpClient } from '../core/httpClient';
import type {
  AIConfigurationResponse,
  UpdateAIConfigRequest,
  TestConnectionResponse,
} from './types';

const BASE = '/api/v1/assistant-config';

export const assistantConfigApi = {
  async getConfig(): Promise<AIConfigurationResponse> {
    const response = await httpClient.get<AIConfigurationResponse>(BASE);
    return response.data;
  },

  async updateConfig(request: UpdateAIConfigRequest): Promise<AIConfigurationResponse> {
    const response = await httpClient.put<AIConfigurationResponse>(BASE, request);
    return response.data;
  },

  async testConnection(): Promise<TestConnectionResponse> {
    const response = await httpClient.post<TestConnectionResponse>(`${BASE}/connection-tests`);
    return response.data;
  },
};
