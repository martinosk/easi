import { httpClient } from '../../../api/core/httpClient';
import type {
  AcquiredEntity,
  AcquiredEntityId,
  AcquiredEntitiesResponse,
  CreateAcquiredEntityRequest,
  UpdateAcquiredEntityRequest,
  Vendor,
  VendorId,
  VendorsResponse,
  CreateVendorRequest,
  UpdateVendorRequest,
  InternalTeam,
  InternalTeamId,
  InternalTeamsResponse,
  CreateInternalTeamRequest,
  UpdateInternalTeamRequest,
  OriginRelationship,
  OriginRelationshipId,
} from '../../../api/types';

export const originEntitiesApi = {
  acquiredEntities: {
    async getAll(): Promise<AcquiredEntity[]> {
      const response = await httpClient.get<AcquiredEntitiesResponse>('/api/v1/acquired-entities');
      return response.data.data;
    },

    async getById(id: AcquiredEntityId): Promise<AcquiredEntity> {
      const response = await httpClient.get<AcquiredEntity>(`/api/v1/acquired-entities/${id}`);
      return response.data;
    },

    async create(request: CreateAcquiredEntityRequest): Promise<AcquiredEntity> {
      const response = await httpClient.post<AcquiredEntity>('/api/v1/acquired-entities', request);
      return response.data;
    },

    async update(id: AcquiredEntityId, request: UpdateAcquiredEntityRequest): Promise<AcquiredEntity> {
      const response = await httpClient.put<AcquiredEntity>(`/api/v1/acquired-entities/${id}`, request);
      return response.data;
    },

    async delete(id: AcquiredEntityId): Promise<void> {
      await httpClient.delete(`/api/v1/acquired-entities/${id}`);
    },

    async linkComponent(componentId: string, acquiredEntityId: AcquiredEntityId): Promise<OriginRelationship> {
      const response = await httpClient.post<OriginRelationship>(
        `/api/v1/components/${componentId}/origin/acquired-via`,
        { acquiredEntityId }
      );
      return response.data;
    },

    async unlinkComponent(relationshipId: OriginRelationshipId): Promise<void> {
      await httpClient.delete(`/api/v1/origin-relationships/acquired-via/${relationshipId}`);
    },
  },

  vendors: {
    async getAll(): Promise<Vendor[]> {
      const response = await httpClient.get<VendorsResponse>('/api/v1/vendors');
      return response.data.data;
    },

    async getById(id: VendorId): Promise<Vendor> {
      const response = await httpClient.get<Vendor>(`/api/v1/vendors/${id}`);
      return response.data;
    },

    async create(request: CreateVendorRequest): Promise<Vendor> {
      const response = await httpClient.post<Vendor>('/api/v1/vendors', request);
      return response.data;
    },

    async update(id: VendorId, request: UpdateVendorRequest): Promise<Vendor> {
      const response = await httpClient.put<Vendor>(`/api/v1/vendors/${id}`, request);
      return response.data;
    },

    async delete(id: VendorId): Promise<void> {
      await httpClient.delete(`/api/v1/vendors/${id}`);
    },

    async linkComponent(componentId: string, vendorId: VendorId): Promise<OriginRelationship> {
      const response = await httpClient.post<OriginRelationship>(
        `/api/v1/components/${componentId}/origin/purchased-from`,
        { vendorId }
      );
      return response.data;
    },

    async unlinkComponent(relationshipId: OriginRelationshipId): Promise<void> {
      await httpClient.delete(`/api/v1/origin-relationships/purchased-from/${relationshipId}`);
    },
  },

  internalTeams: {
    async getAll(): Promise<InternalTeam[]> {
      const response = await httpClient.get<InternalTeamsResponse>('/api/v1/internal-teams');
      return response.data.data;
    },

    async getById(id: InternalTeamId): Promise<InternalTeam> {
      const response = await httpClient.get<InternalTeam>(`/api/v1/internal-teams/${id}`);
      return response.data;
    },

    async create(request: CreateInternalTeamRequest): Promise<InternalTeam> {
      const response = await httpClient.post<InternalTeam>('/api/v1/internal-teams', request);
      return response.data;
    },

    async update(id: InternalTeamId, request: UpdateInternalTeamRequest): Promise<InternalTeam> {
      const response = await httpClient.put<InternalTeam>(`/api/v1/internal-teams/${id}`, request);
      return response.data;
    },

    async delete(id: InternalTeamId): Promise<void> {
      await httpClient.delete(`/api/v1/internal-teams/${id}`);
    },

    async linkComponent(componentId: string, internalTeamId: InternalTeamId): Promise<OriginRelationship> {
      const response = await httpClient.post<OriginRelationship>(
        `/api/v1/components/${componentId}/origin/built-by`,
        { internalTeamId }
      );
      return response.data;
    },

    async unlinkComponent(relationshipId: OriginRelationshipId): Promise<void> {
      await httpClient.delete(`/api/v1/origin-relationships/built-by/${relationshipId}`);
    },
  },
};
