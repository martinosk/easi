import { httpClient } from '../core/httpClient';
import type {
  StrategyPillarsConfiguration,
  StrategyPillar,
  CreateStrategyPillarRequest,
  UpdateStrategyPillarRequest,
} from '../types';

export interface StrategyPillarsConfigurationWithVersion extends StrategyPillarsConfiguration {
  version: number;
}

export interface PillarChange {
  operation: 'add' | 'update' | 'remove';
  id?: string;
  name?: string;
  description?: string;
  fitScoringEnabled?: boolean;
  fitCriteria?: string;
}

export interface UpdateFitConfigurationRequest {
  fitScoringEnabled: boolean;
  fitCriteria: string;
}

export interface BatchUpdateRequest {
  changes: PillarChange[];
}

export interface BatchUpdateResponse {
  data: StrategyPillar[];
  _links: Record<string, string>;
}

function parseETag(etag: string | undefined): number {
  if (!etag) return 0;
  const match = etag.match(/^"?(\d+)"?$/);
  return match ? parseInt(match[1], 10) : 0;
}

export const strategyPillarsApi = {
  async getConfiguration(includeInactive = true): Promise<StrategyPillarsConfigurationWithVersion> {
    const response = await httpClient.get<StrategyPillarsConfiguration>(
      `/api/v1/meta-model/strategy-pillars?includeInactive=${includeInactive}`
    );
    const version = parseETag(response.headers['etag']);
    return { ...response.data, version };
  },

  async batchUpdate(request: BatchUpdateRequest, version: number): Promise<BatchUpdateResponse> {
    const response = await httpClient.patch<BatchUpdateResponse>(
      '/api/v1/meta-model/strategy-pillars',
      request,
      {
        headers: {
          'If-Match': `"${version}"`,
        },
      }
    );
    return response.data;
  },

  async createPillar(request: CreateStrategyPillarRequest): Promise<StrategyPillar> {
    const response = await httpClient.post<StrategyPillar>(
      '/api/v1/meta-model/strategy-pillars',
      request
    );
    return response.data;
  },

  async updatePillar(id: string, request: UpdateStrategyPillarRequest, version: number): Promise<StrategyPillar> {
    const response = await httpClient.put<StrategyPillar>(
      `/api/v1/meta-model/strategy-pillars/${id}`,
      request,
      {
        headers: {
          'If-Match': `"${version}"`,
        },
      }
    );
    return response.data;
  },

  async deletePillar(id: string): Promise<void> {
    await httpClient.delete(`/api/v1/meta-model/strategy-pillars/${id}`);
  },

  async updateFitConfiguration(
    id: string,
    request: UpdateFitConfigurationRequest,
    version: number
  ): Promise<StrategyPillar> {
    const response = await httpClient.put<StrategyPillar>(
      `/api/v1/meta-model/strategy-pillars/${id}/fit-configuration`,
      request,
      {
        headers: {
          'If-Match': `"${version}"`,
        },
      }
    );
    return response.data;
  },
};

export default strategyPillarsApi;
