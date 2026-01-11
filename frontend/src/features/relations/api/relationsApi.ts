import { httpClient, fetchAllPaginated } from '../../../api/core';
import { followLink } from '../../../utils/hateoas';
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

  async update(relation: Relation, request: Partial<CreateRelationRequest>): Promise<Relation> {
    const response = await httpClient.put<Relation>(followLink(relation, 'edit'), request);
    return response.data;
  },

  async delete(relation: Relation): Promise<void> {
    await httpClient.delete(followLink(relation, 'delete'));
  },
};

export default relationsApi;
