import { httpClient } from '../../../api/core/httpClient';
import type { EnterpriseCapabilityId } from '../../../api/types';
import type {
  ECStandardApplicationResponse,
  SetStandardApplicationRequest,
  StandardApplication,
  StandardApplicationHistory,
} from '../types';

const path = (ecId: EnterpriseCapabilityId, suffix = '') =>
  `/api/v1/enterprise-capabilities/${ecId}/standard-application${suffix}`;

export const standardApplicationApi = {
  async getForEnterpriseCapability(id: EnterpriseCapabilityId): Promise<ECStandardApplicationResponse> {
    const response = await httpClient.get<ECStandardApplicationResponse>(path(id));
    return response.data;
  },

  async set(id: EnterpriseCapabilityId, request: SetStandardApplicationRequest): Promise<StandardApplication> {
    const response = await httpClient.put<StandardApplication>(path(id), request);
    return response.data;
  },

  async getHistory(id: EnterpriseCapabilityId): Promise<StandardApplicationHistory> {
    const response = await httpClient.get<StandardApplicationHistory>(path(id, '/history'));
    return response.data;
  },
};
