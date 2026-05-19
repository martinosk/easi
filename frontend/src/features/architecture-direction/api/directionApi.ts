import { httpClient } from '../../../api/core/httpClient';
import type { EnterpriseCapabilityId } from '../../../api/types';
import type {
  CaptureDirectionRequest,
  Direction,
  DirectionId,
  ECDirectionResponse,
  UpdateDirectionRequest,
} from '../types';

export const directionApi = {
  async getForEnterpriseCapability(id: EnterpriseCapabilityId): Promise<ECDirectionResponse> {
    const response = await httpClient.get<ECDirectionResponse>(`/api/v1/enterprise-capabilities/${id}/direction`);
    return response.data;
  },

  async capture(id: EnterpriseCapabilityId, request: CaptureDirectionRequest): Promise<Direction> {
    const response = await httpClient.post<Direction>(`/api/v1/enterprise-capabilities/${id}/direction`, request);
    return response.data;
  },

  async getById(id: DirectionId): Promise<Direction> {
    const response = await httpClient.get<Direction>(`/api/v1/directions/${id}`);
    return response.data;
  },

  async update(id: DirectionId, request: UpdateDirectionRequest): Promise<Direction> {
    const response = await httpClient.put<Direction>(`/api/v1/directions/${id}`, request);
    return response.data;
  },

  async advance(id: DirectionId, target: 'proposed' | 'agreed'): Promise<Direction> {
    const response = await httpClient.post<Direction>(`/api/v1/directions/${id}/advance/${target}`);
    return response.data;
  },

  async reject(id: DirectionId): Promise<Direction> {
    const response = await httpClient.post<Direction>(`/api/v1/directions/${id}/reject`);
    return response.data;
  },
};
