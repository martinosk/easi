import { httpClient } from '../../../api/core';
import type {
  BusinessDomainId,
  CapabilityId,
  StrategyImportance,
  StrategyImportanceId,
  SetStrategyImportanceRequest,
  UpdateStrategyImportanceRequest,
  CollectionResponse,
} from '../../../api/types';

export const strategyImportanceApi = {
  async getByDomainAndCapability(
    domainId: BusinessDomainId,
    capabilityId: CapabilityId
  ): Promise<StrategyImportance[]> {
    const response = await httpClient.get<CollectionResponse<StrategyImportance>>(
      `/api/v1/business-domains/${domainId}/capabilities/${capabilityId}/importance`
    );
    return response.data.data || [];
  },

  async getByDomain(domainId: BusinessDomainId): Promise<StrategyImportance[]> {
    const response = await httpClient.get<CollectionResponse<StrategyImportance>>(
      `/api/v1/business-domains/${domainId}/importance`
    );
    return response.data.data || [];
  },

  async getByCapability(capabilityId: CapabilityId): Promise<StrategyImportance[]> {
    const response = await httpClient.get<CollectionResponse<StrategyImportance>>(
      `/api/v1/capabilities/${capabilityId}/importance`
    );
    return response.data.data || [];
  },

  async setImportance(
    domainId: BusinessDomainId,
    capabilityId: CapabilityId,
    request: SetStrategyImportanceRequest
  ): Promise<StrategyImportance> {
    const response = await httpClient.post<StrategyImportance>(
      `/api/v1/business-domains/${domainId}/capabilities/${capabilityId}/importance`,
      request
    );
    return response.data;
  },

  async updateImportance(
    domainId: BusinessDomainId,
    capabilityId: CapabilityId,
    importanceId: StrategyImportanceId,
    request: UpdateStrategyImportanceRequest
  ): Promise<StrategyImportance> {
    const response = await httpClient.put<StrategyImportance>(
      `/api/v1/business-domains/${domainId}/capabilities/${capabilityId}/importance/${importanceId}`,
      request
    );
    return response.data;
  },

  async removeImportance(
    domainId: BusinessDomainId,
    capabilityId: CapabilityId,
    importanceId: StrategyImportanceId
  ): Promise<void> {
    await httpClient.delete(
      `/api/v1/business-domains/${domainId}/capabilities/${capabilityId}/importance/${importanceId}`
    );
  },
};

export default strategyImportanceApi;
