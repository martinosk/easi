import { httpClient } from '../../../api/core';
import { followLink } from '../../../utils/hateoas';
import type {
  ValueStream,
  ValueStreamId,
  CreateValueStreamRequest,
  UpdateValueStreamRequest,
  ValueStreamsResponse,
} from '../../../api/types';

export const valueStreamsApi = {
  async getAll(): Promise<ValueStreamsResponse> {
    const response = await httpClient.get<ValueStreamsResponse>('/api/v1/value-streams');
    return response.data;
  },

  async getById(id: ValueStreamId): Promise<ValueStream> {
    const response = await httpClient.get<ValueStream>(`/api/v1/value-streams/${id}`);
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
};
