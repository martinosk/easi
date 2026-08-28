import { httpClient } from '../../../api/core/httpClient';
import type { CapabilityId } from '../../../api/types';
import { followLink } from '../../../utils/hateoas';
import type {
  AddJourneyMilestoneRequest,
  CapabilityJourney,
  CapabilityJourneyHistoryResponse,
  CapabilityJourneyResponse,
  CapabilityJourneysBulkResponse,
  CaptureJourneyRequest,
  ChangeJourneySourceApplicationsRequest,
  JourneyMilestone,
  ReorderJourneyMilestonesRequest,
  UpdateJourneyDetailsRequest,
  UpdateJourneyMilestoneRequest,
  UpdateJourneyProgressRequest,
} from '../types';

const journeyPath = (capabilityId: CapabilityId | string) => `/api/v1/capabilities/${capabilityId}/journey`;

export const journeyApi = {
  async getForCapability(capabilityId: CapabilityId | string): Promise<CapabilityJourneyResponse> {
    const response = await httpClient.get<CapabilityJourneyResponse>(journeyPath(capabilityId));
    return response.data;
  },

  async getHistory(capabilityId: CapabilityId | string): Promise<CapabilityJourneyHistoryResponse> {
    const response = await httpClient.get<CapabilityJourneyHistoryResponse>(`${journeyPath(capabilityId)}/history`);
    return response.data;
  },

  async getAll(): Promise<CapabilityJourneysBulkResponse> {
    const response = await httpClient.get<CapabilityJourneysBulkResponse>('/api/v1/capability-journeys');
    return response.data;
  },

  async capture(capabilityId: CapabilityId | string, request: CaptureJourneyRequest): Promise<CapabilityJourney> {
    const response = await httpClient.post<CapabilityJourney>(journeyPath(capabilityId), request);
    return response.data;
  },

  async start(journey: CapabilityJourney): Promise<CapabilityJourney> {
    const response = await httpClient.post<CapabilityJourney>(followLink(journey, 'x-start'));
    return response.data;
  },

  async complete(journey: CapabilityJourney): Promise<CapabilityJourney> {
    const response = await httpClient.post<CapabilityJourney>(followLink(journey, 'x-complete'));
    return response.data;
  },

  async abandon(journey: CapabilityJourney): Promise<CapabilityJourney> {
    const response = await httpClient.post<CapabilityJourney>(followLink(journey, 'x-abandon'));
    return response.data;
  },

  async updateDetails(journey: CapabilityJourney, request: UpdateJourneyDetailsRequest): Promise<CapabilityJourney> {
    const response = await httpClient.put<CapabilityJourney>(followLink(journey, 'edit'), request);
    return response.data;
  },

  async updateProgress(journey: CapabilityJourney, request: UpdateJourneyProgressRequest): Promise<CapabilityJourney> {
    const response = await httpClient.put<CapabilityJourney>(followLink(journey, 'x-progress'), request);
    return response.data;
  },

  async changeSourceApplications(
    journey: CapabilityJourney,
    request: ChangeJourneySourceApplicationsRequest,
  ): Promise<CapabilityJourney> {
    const response = await httpClient.put<CapabilityJourney>(followLink(journey, 'x-change-sources'), request);
    return response.data;
  },

  async addMilestone(journey: CapabilityJourney, request: AddJourneyMilestoneRequest): Promise<CapabilityJourney> {
    const response = await httpClient.post<CapabilityJourney>(followLink(journey, 'x-add-milestone'), request);
    return response.data;
  },

  async updateMilestone(
    milestone: JourneyMilestone,
    request: UpdateJourneyMilestoneRequest,
  ): Promise<CapabilityJourney> {
    const response = await httpClient.put<CapabilityJourney>(followLink(milestone, 'edit'), request);
    return response.data;
  },

  async removeMilestone(milestone: JourneyMilestone): Promise<void> {
    await httpClient.delete(followLink(milestone, 'delete'));
  },

  async reorderMilestones(
    journey: CapabilityJourney,
    request: ReorderJourneyMilestonesRequest,
  ): Promise<CapabilityJourney> {
    const response = await httpClient.put<CapabilityJourney>(followLink(journey, 'x-reorder-milestones'), request);
    return response.data;
  },
};
