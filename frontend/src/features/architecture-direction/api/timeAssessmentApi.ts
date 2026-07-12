import { httpClient } from '../../../api/core/httpClient';
import { ApiError, type CapabilityId, type ComponentId } from '../../../api/types';
import { followLink } from '../../../utils/hateoas';
import type {
  AssessRealizationRequest,
  TimeAssessment,
  TimeAssessmentRollupsResponse,
  TimeAssessmentsResponse,
} from '../types';

const assessmentPath = (capabilityId: CapabilityId | string, componentId: ComponentId | string) =>
  `/api/v1/capabilities/${capabilityId}/components/${componentId}/time-assessment`;

export const timeAssessmentApi = {
  async getByCapabilityIds(capabilityIds: (CapabilityId | string)[]): Promise<TimeAssessmentsResponse> {
    const response = await httpClient.get<TimeAssessmentsResponse>(
      `/api/v1/time-assessments?capabilityIds=${capabilityIds.join(',')}`,
    );
    return response.data;
  },

  async getRollups(componentIds: (ComponentId | string)[]): Promise<TimeAssessmentRollupsResponse> {
    const response = await httpClient.get<TimeAssessmentRollupsResponse>(
      `/api/v1/time-assessments/rollups?componentIds=${componentIds.join(',')}`,
    );
    return response.data;
  },

  async getOne(capabilityId: CapabilityId | string, componentId: ComponentId | string): Promise<TimeAssessment | null> {
    try {
      const response = await httpClient.get<TimeAssessment>(assessmentPath(capabilityId, componentId));
      return response.data;
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 404) return null;
      throw err;
    }
  },

  async assess(
    capabilityId: CapabilityId | string,
    componentId: ComponentId | string,
    request: AssessRealizationRequest,
  ): Promise<TimeAssessment> {
    const response = await httpClient.put<TimeAssessment>(assessmentPath(capabilityId, componentId), request);
    return response.data;
  },

  async remove(assessment: TimeAssessment): Promise<void> {
    await httpClient.delete(followLink(assessment, 'delete'));
  },
};
