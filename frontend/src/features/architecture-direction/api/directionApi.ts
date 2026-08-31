import { httpClient } from '../../../api/core/httpClient';
import type { EnterpriseCapabilityId } from '../../../api/types';
import type {
  CaptureDirectionRequest,
  CompositionPreviewResponse,
  Direction,
  ECDirectionResponse,
  SourceCandidatesQuery,
  SourceCandidatesResponse,
  UpdateDirectionRequest,
} from '../types';

const path = (ecId: EnterpriseCapabilityId, suffix = '') =>
  `/api/v1/enterprise-capabilities/${ecId}/direction${suffix}`;

export const directionApi = {
  async getForEnterpriseCapability(id: EnterpriseCapabilityId, href?: string): Promise<ECDirectionResponse> {
    const response = await httpClient.get<ECDirectionResponse>(href ?? path(id));
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

  async revert(id: EnterpriseCapabilityId): Promise<Direction> {
    const response = await httpClient.post<Direction>(path(id, '/revert'));
    return response.data;
  },

  async addSource(id: EnterpriseCapabilityId, capabilityId: string): Promise<Direction> {
    const response = await httpClient.post<Direction>(path(id, '/sources'), { capabilityId });
    return response.data;
  },

  async removeSource(id: EnterpriseCapabilityId, capabilityId: string): Promise<void> {
    await httpClient.delete(path(id, `/sources/${capabilityId}`));
  },

  async previewComposition(
    id: EnterpriseCapabilityId,
    sourceCapabilityIds: string[],
  ): Promise<CompositionPreviewResponse> {
    const response = await httpClient.post<CompositionPreviewResponse>(path(id, '/composition-preview'), {
      sourceCapabilityIds,
    });
    return response.data;
  },

  async searchSourceCandidates(
    id: EnterpriseCapabilityId,
    query: SourceCandidatesQuery,
  ): Promise<SourceCandidatesResponse> {
    const params = new URLSearchParams({ q: query.q, ecId: id });
    if (query.domainId) params.set('domainId', query.domainId);
    if (query.limit) params.set('limit', String(query.limit));
    const response = await httpClient.get<SourceCandidatesResponse>(
      `/api/v1/capabilities/source-candidates?${params.toString()}`,
    );
    return response.data;
  },
};
