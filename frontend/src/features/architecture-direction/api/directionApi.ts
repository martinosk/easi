import { httpClient } from '../../../api/core/httpClient';
import type { EnterpriseCapabilityId } from '../../../api/types';
import type {
  CaptureDirectionRequest,
  Direction,
  ECDirectionResponse,
  UpdateDirectionRequest,
} from '../types';

const path = (ecId: EnterpriseCapabilityId, suffix = '') =>
  `/api/v1/enterprise-capabilities/${ecId}/direction${suffix}`;

export const directionApi = {
  async getForEnterpriseCapability(id: EnterpriseCapabilityId): Promise<ECDirectionResponse> {
    const response = await httpClient.get<ECDirectionResponse>(path(id));
    return response.data;
  },

  async capture(id: EnterpriseCapabilityId, request: CaptureDirectionRequest): Promise<Direction> {
    const response = await httpClient.post<Direction>(path(id), request);
    return response.data;
  },

  async update(id: EnterpriseCapabilityId, request: UpdateDirectionRequest): Promise<Direction> {
    const response = await httpClient.put<Direction>(path(id), request);
    return response.data;
  },

  async propose(id: EnterpriseCapabilityId): Promise<Direction> {
    const response = await httpClient.post<Direction>(path(id, '/propose'));
    return response.data;
  },

  async agree(id: EnterpriseCapabilityId): Promise<Direction> {
    const response = await httpClient.post<Direction>(path(id, '/agree'));
    return response.data;
  },

  async reject(id: EnterpriseCapabilityId): Promise<Direction> {
    const response = await httpClient.post<Direction>(path(id, '/reject'));
    return response.data;
  },
};
