import { httpClient, fetchAllPaginated } from '../../../api/core';
import type {
  Relation,
  RelationId,
  CreateRelationRequest,
} from '../../../api/types';

export const relationsApi = {
  async getAll(): Promise<Relation[]> {
    return fetchAllPaginated<Relation>('/api/v1/relations');
  },

  async getById(id: RelationId): Promise<Relation> {
    const response = await httpClient.get<Relation>(`/api/v1/relations/${id}`);
    return response.data;
  },

  async create(request: CreateRelationRequest): Promise<Relation> {
    const response = await httpClient.post<Relation>('/api/v1/relations', request);
    return response.data;
  },

  async update(id: RelationId, request: Partial<CreateRelationRequest>): Promise<Relation> {
    const response = await httpClient.put<Relation>(`/api/v1/relations/${id}`, request);
    return response.data;
  },

  async delete(id: RelationId): Promise<void> {
    await httpClient.delete(`/api/v1/relations/${id}`);
  },
};

export default relationsApi;
