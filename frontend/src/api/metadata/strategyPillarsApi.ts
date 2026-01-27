import { httpClient } from '../core/httpClient';
import type {
  StrategyPillarsConfiguration,
  StrategyPillar,
  CreateStrategyPillarRequest,
  UpdateStrategyPillarRequest,
  FitType,
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
  fitType?: FitType;
}

export interface UpdateFitConfigurationRequest {
  fitScoringEnabled: boolean;
  fitCriteria: string;
  fitType: FitType;
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

function withOptimisticLocking(version: number) {
  return { headers: { 'If-Match': `"${version}"` } };
}

const PILLARS_BASE_URL = '/api/v1/meta-model/strategy-pillars';

export const strategyPillarsApi = {
  async getConfiguration(includeInactive = true): Promise<StrategyPillarsConfigurationWithVersion> {
    const response = await httpClient.get<StrategyPillarsConfiguration>(`${PILLARS_BASE_URL}?includeInactive=${includeInactive}`);
    return { ...response.data, version: parseETag(response.headers['etag']) };
  },

  async batchUpdate(request: BatchUpdateRequest, version: number): Promise<BatchUpdateResponse> {
    const response = await httpClient.patch<BatchUpdateResponse>(PILLARS_BASE_URL, request, withOptimisticLocking(version));
    return response.data;
  },

  async createPillar(request: CreateStrategyPillarRequest): Promise<StrategyPillar> {
    const response = await httpClient.post<StrategyPillar>(PILLARS_BASE_URL, request);
    return response.data;
  },

  async updatePillar(id: string, request: UpdateStrategyPillarRequest, version: number): Promise<StrategyPillar> {
    const response = await httpClient.put<StrategyPillar>(`${PILLARS_BASE_URL}/${id}`, request, withOptimisticLocking(version));
    return response.data;
  },

  async deletePillar(id: string): Promise<void> {
    await httpClient.delete(`${PILLARS_BASE_URL}/${id}`);
  },

  async updateFitConfiguration(id: string, request: UpdateFitConfigurationRequest, version: number): Promise<StrategyPillar> {
    const response = await httpClient.put<StrategyPillar>(`${PILLARS_BASE_URL}/${id}/fit-configuration`, request, withOptimisticLocking(version));
    return response.data;
  },
};

export default strategyPillarsApi;
