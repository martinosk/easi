import { httpClient } from '../../../api/core';
import type {
  ComponentId,
  CapabilityId,
  BusinessDomainId,
  ApplicationFitScore,
  ApplicationFitScoresResponse,
  SetApplicationFitScoreRequest,
  FitComparison,
  FitComparisonsResponse,
} from '../../../api/types';

export const fitScoresApi = {
  async getByComponent(componentId: ComponentId): Promise<ApplicationFitScoresResponse> {
    const response = await httpClient.get<ApplicationFitScoresResponse>(
      `/api/v1/components/${componentId}/fit-scores`
    );
    return response.data;
  },

  async setScore(
    componentId: ComponentId,
    pillarId: string,
    request: SetApplicationFitScoreRequest
  ): Promise<ApplicationFitScore> {
    const response = await httpClient.put<ApplicationFitScore>(
      `/api/v1/components/${componentId}/fit-scores/${pillarId}`,
      request
    );
    return response.data;
  },

  async deleteScore(componentId: ComponentId, pillarId: string): Promise<void> {
    await httpClient.delete(`/api/v1/components/${componentId}/fit-scores/${pillarId}`);
  },

  async getFitComparisons(
    componentId: ComponentId,
    capabilityId: CapabilityId,
    businessDomainId: BusinessDomainId
  ): Promise<FitComparison[]> {
    const response = await httpClient.get<FitComparisonsResponse>(
      `/api/v1/components/${componentId}/fit-comparisons`,
      { params: { capabilityId, businessDomainId } }
    );
    return response.data.data;
  },
};

export default fitScoresApi;
