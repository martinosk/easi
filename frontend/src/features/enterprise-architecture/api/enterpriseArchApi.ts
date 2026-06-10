import { httpClient } from '../../../api/core/httpClient';
import type { StrategicFitAnalysis } from '../../../api/types';
import type {
  CompositionResponse,
  CreateEnterpriseCapabilityRequest,
  EnterpriseCapabilitiesListResponse,
  EnterpriseCapability,
  EnterpriseCapabilityId,
  EnterpriseStrategicImportanceId,
  MaturityAnalysisResponse,
  MaturityGapDetail,
  SetStrategicImportanceRequest,
  StrategicImportance,
  TimeSuggestionsResponse,
  UpdateEnterpriseCapabilityRequest,
  UpdateStrategicImportanceRequest,
} from '../types';

function strategicImportanceUrl(
  capabilityId: EnterpriseCapabilityId,
  importanceId?: EnterpriseStrategicImportanceId,
): string {
  const base = `/api/v1/enterprise-capabilities/${capabilityId}/strategic-importance`;
  return importanceId ? `${base}/${importanceId}` : base;
}

export const enterpriseArchApi = {
  async getAll(): Promise<EnterpriseCapability[]> {
    const response = await httpClient.get<EnterpriseCapabilitiesListResponse>('/api/v1/enterprise-capabilities');
    return response.data.data;
  },

  async getById(id: EnterpriseCapabilityId): Promise<EnterpriseCapability> {
    const response = await httpClient.get<EnterpriseCapability>(`/api/v1/enterprise-capabilities/${id}`);
    return response.data;
  },

  async getComposition(id: EnterpriseCapabilityId): Promise<CompositionResponse> {
    const response = await httpClient.get<CompositionResponse>(`/api/v1/enterprise-capabilities/${id}/composition`);
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

  async getStrategicImportance(enterpriseCapabilityId: EnterpriseCapabilityId): Promise<StrategicImportance[]> {
    const response = await httpClient.get<{ data: StrategicImportance[] }>(
      strategicImportanceUrl(enterpriseCapabilityId),
    );
    return response.data.data;
  },

  async setStrategicImportance(
    enterpriseCapabilityId: EnterpriseCapabilityId,
    request: SetStrategicImportanceRequest,
  ): Promise<StrategicImportance> {
    const response = await httpClient.post<StrategicImportance>(
      strategicImportanceUrl(enterpriseCapabilityId),
      request,
    );
    return response.data;
  },

  async updateStrategicImportance(
    enterpriseCapabilityId: EnterpriseCapabilityId,
    importanceId: EnterpriseStrategicImportanceId,
    request: UpdateStrategicImportanceRequest,
  ): Promise<StrategicImportance> {
    const response = await httpClient.put<StrategicImportance>(
      strategicImportanceUrl(enterpriseCapabilityId, importanceId),
      request,
    );
    return response.data;
  },

  async removeStrategicImportance(
    enterpriseCapabilityId: EnterpriseCapabilityId,
    importanceId: EnterpriseStrategicImportanceId,
  ): Promise<void> {
    await httpClient.delete(strategicImportanceUrl(enterpriseCapabilityId, importanceId));
  },

  async setTargetMaturity(enterpriseCapabilityId: EnterpriseCapabilityId, targetMaturity: number): Promise<void> {
    await httpClient.put(`/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/target-maturity`, {
      targetMaturity,
    });
  },

  async getMaturityAnalysisCandidates(sortBy?: string): Promise<MaturityAnalysisResponse> {
    const params = sortBy ? `?sortBy=${sortBy}` : '';
    const response = await httpClient.get<MaturityAnalysisResponse>(
      `/api/v1/enterprise-capabilities/maturity-analysis${params}`,
    );
    return response.data;
  },

  async getMaturityGapDetail(enterpriseCapabilityId: EnterpriseCapabilityId): Promise<MaturityGapDetail> {
    const response = await httpClient.get<MaturityGapDetail>(
      `/api/v1/enterprise-capabilities/${enterpriseCapabilityId}/maturity-gap`,
    );
    return response.data;
  },

  async getStrategicFitAnalysis(pillarId: string): Promise<StrategicFitAnalysis> {
    const response = await httpClient.get<StrategicFitAnalysis>(`/api/v1/strategic-fit-analysis/${pillarId}`);
    return response.data;
  },

  async getTimeSuggestions(filters?: {
    capabilityId?: string;
    componentId?: string;
  }): Promise<TimeSuggestionsResponse> {
    const params = new URLSearchParams();
    if (filters?.capabilityId) {
      params.append('capabilityId', filters.capabilityId);
    }
    if (filters?.componentId) {
      params.append('componentId', filters.componentId);
    }
    const queryString = params.toString();
    const url = queryString ? `/api/v1/time-suggestions?${queryString}` : '/api/v1/time-suggestions';
    const response = await httpClient.get<TimeSuggestionsResponse>(url);
    return response.data;
  },
};
