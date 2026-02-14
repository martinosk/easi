import { httpClient } from '../../../api/core/httpClient';
import { followLink } from '../../../utils/hateoas';
import type {
  ValueStream,
  ValueStreamId,
  ValueStreamDetail,
  ValueStreamStage,
  CreateValueStreamRequest,
  UpdateValueStreamRequest,
  CreateStageRequest,
  UpdateStageRequest,
  ReorderStagesRequest,
  StageCapabilityMapping,
  ValueStreamsResponse,
} from '../../../api/types';

export const valueStreamsApi = {
  async getAll(): Promise<ValueStreamsResponse> {
    const response = await httpClient.get<ValueStreamsResponse>('/api/v1/value-streams');
    return response.data;
  },

  async getById(id: ValueStreamId): Promise<ValueStreamDetail> {
    const response = await httpClient.get<ValueStreamDetail>(`/api/v1/value-streams/${id}`);
    return response.data;
  },

  async create(request: CreateValueStreamRequest): Promise<ValueStream> {
    const response = await httpClient.post<ValueStream>('/api/v1/value-streams', request);
    return response.data;
  },

  async update(valueStream: ValueStream, request: UpdateValueStreamRequest): Promise<ValueStream> {
    const response = await httpClient.put<ValueStream>(followLink(valueStream, 'edit'), request);
    return response.data;
  },

  async delete(valueStream: ValueStream): Promise<void> {
    await httpClient.delete(followLink(valueStream, 'delete'));
  },

  async addStage(valueStream: ValueStream, request: CreateStageRequest): Promise<ValueStreamDetail> {
    const response = await httpClient.post<ValueStreamDetail>(followLink(valueStream, 'x-add-stage'), request);
    return response.data;
  },

  async updateStage(stage: ValueStreamStage, request: UpdateStageRequest): Promise<ValueStreamDetail> {
    const response = await httpClient.put<ValueStreamDetail>(followLink(stage, 'edit'), request);
    return response.data;
  },

  async deleteStage(stage: ValueStreamStage): Promise<ValueStreamDetail> {
    const response = await httpClient.delete<ValueStreamDetail>(followLink(stage, 'delete'));
    return response.data;
  },

  async reorderStages(valueStream: ValueStream, request: ReorderStagesRequest): Promise<ValueStreamDetail> {
    const response = await httpClient.put<ValueStreamDetail>(followLink(valueStream, 'x-reorder-stages'), request);
    return response.data;
  },

  async addStageCapability(stage: ValueStreamStage, capabilityId: string): Promise<ValueStreamDetail> {
    const response = await httpClient.post<ValueStreamDetail>(
      followLink(stage, 'x-add-capability'),
      { capabilityId },
    );
    return response.data;
  },

  async removeStageCapability(mapping: StageCapabilityMapping): Promise<ValueStreamDetail> {
    const response = await httpClient.delete<ValueStreamDetail>(followLink(mapping, 'delete'));
    return response.data;
  },
};
