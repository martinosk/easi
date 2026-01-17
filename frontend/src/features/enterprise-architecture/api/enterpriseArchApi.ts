import { httpClient } from '../../../api/core/httpClient';
import type { CapabilityId, StrategicFitAnalysis } from '../../../api/types';
import type {
  EnterpriseCapability,
  EnterpriseCapabilityId,
  EnterpriseCapabilityLink,
  EnterpriseCapabilityLinkId,
  EnterpriseStrategicImportanceId,
  StrategicImportance,
  CreateEnterpriseCapabilityRequest,
  UpdateEnterpriseCapabilityRequest,
  LinkCapabilityRequest,
  SetStrategicImportanceRequest,
  UpdateStrategicImportanceRequest,
  EnterpriseCapabilitiesListResponse,
  DomainCapabilityLinkStatus,
  CapabilityLinkStatusResponse,
  MaturityAnalysisResponse,
  MaturityGapDetail,
} from '../types';

export const enterpriseArchApi = {
  async getAll(): Promise<EnterpriseCapability[]> {
    const response = await httpClient.get<EnterpriseCapabilitiesListResponse>('/api/v1/enterprise-capabilities');
    return response.data.data;
  },

  async getById(id: EnterpriseCapabilityId): Promise<EnterpriseCapability> {
    const response = await httpClient.get<EnterpriseCapability>(`/api/v1/enterprise-capabilities/${id}`);
    return response.data;
  },

  async create(request: CreateEnterpriseCapabilityRequest): Promise<EnterpriseCapability> {
    const response = await httpClient.post<EnterpriseCapability>('/api/v1/enterprise-capabilities', request);
    return response.data;
  },

  async update(id: EnterpriseCapabilityId, request: UpdateEnterpriseCapabilityRequest): Promise<EnterpriseCapability> {
    const response = await httpClient.put<EnterpriseCapability>(`/api/v1/enterprise-capabilities/${id}`, request);
    return response.data;
  },

  async delete(id: EnterpriseCapabilityId): Promise<void> {
    await httpClient.delete(`/api/v1/enterprise-capabilities/${id}`);
  },

  async getLinks(enterpriseCapabilityId: EnterpriseCapabilityId): Promise<EnterpriseCapabilityLink[]> {
    const response = await httpClient.get<{ data: EnterpriseCapabilityLink[] }>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/links`
    );
    return response.data.data;
  },

  async linkDomainCapability(enterpriseCapabilityId: EnterpriseCapabilityId, request: LinkCapabilityRequest): Promise<EnterpriseCapabilityLink> {
    const response = await httpClient.post<EnterpriseCapabilityLink>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/links`,
      request
    );
    return response.data;
  },

  async unlinkDomainCapability(enterpriseCapabilityId: EnterpriseCapabilityId, linkId: EnterpriseCapabilityLinkId): Promise<void> {
    await httpClient.delete(`/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/links/${linkId}`);
  },

  async getStrategicImportance(enterpriseCapabilityId: EnterpriseCapabilityId): Promise<StrategicImportance[]> {
    const response = await httpClient.get<{ data: StrategicImportance[] }>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/strategic-importance`
    );
    return response.data.data;
  },

  async setStrategicImportance(
    enterpriseCapabilityId: EnterpriseCapabilityId,
    request: SetStrategicImportanceRequest
  ): Promise<StrategicImportance> {
    const response = await httpClient.post<StrategicImportance>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/strategic-importance`,
      request
    );
    return response.data;
  },

  async updateStrategicImportance(
    enterpriseCapabilityId: EnterpriseCapabilityId,
    importanceId: EnterpriseStrategicImportanceId,
    request: UpdateStrategicImportanceRequest
  ): Promise<StrategicImportance> {
    const response = await httpClient.put<StrategicImportance>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/strategic-importance/${importanceId}`,
      request
    );
    return response.data;
  },

  async removeStrategicImportance(enterpriseCapabilityId: EnterpriseCapabilityId, importanceId: EnterpriseStrategicImportanceId): Promise<void> {
    await httpClient.delete(`/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/strategic-importance/${importanceId}`);
  },

  async getDomainCapabilityLinkStatus(domainCapabilityId: CapabilityId): Promise<DomainCapabilityLinkStatus> {
    const response = await httpClient.get<DomainCapabilityLinkStatus>(
      `/api/v1/domain-capabilities/${domainCapabilityId}/enterprise-capability`
    );
    return response.data;
  },

  async getLinkStatus(capabilityId: string): Promise<CapabilityLinkStatusResponse> {
    const response = await httpClient.get<CapabilityLinkStatusResponse>(
      `/api/v1/domain-capabilities/${capabilityId}/enterprise-link-status`
    );
    return response.data;
  },

  async getBatchLinkStatus(capabilityIds: string[]): Promise<CapabilityLinkStatusResponse[]> {
    const response = await httpClient.get<{ data: CapabilityLinkStatusResponse[] }>(
      `/api/v1/domain-capabilities/enterprise-link-status?capabilityIds=${capabilityIds.join(',')}`
    );
    return response.data.data;
  },

  async setTargetMaturity(enterpriseCapabilityId: EnterpriseCapabilityId, targetMaturity: number): Promise<void> {
    await httpClient.put(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/target-maturity`,
      { targetMaturity }
    );
  },

  async getMaturityAnalysisCandidates(sortBy?: string): Promise<MaturityAnalysisResponse> {
    const params = sortBy ? `?sortBy=${sortBy}` : '';
    const response = await httpClient.get<MaturityAnalysisResponse>(
      `/api/v1/enterprise-capabilities/maturity-analysis${params}`
    );
    return response.data;
  },

  async getMaturityGapDetail(enterpriseCapabilityId: EnterpriseCapabilityId): Promise<MaturityGapDetail> {
    const response = await httpClient.get<MaturityGapDetail>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/maturity-gap`
    );
    return response.data;
  },

  async getStrategicFitAnalysis(pillarId: string): Promise<StrategicFitAnalysis> {
    const response = await httpClient.get<StrategicFitAnalysis>(
      `/api/v1/strategic-fit-analysis/${pillarId}`
    );
    return response.data;
  },
};
