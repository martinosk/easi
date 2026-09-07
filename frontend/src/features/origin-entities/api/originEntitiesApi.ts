import { httpClient } from '../../../api/core/httpClient';
import type {
  AcquiredEntitiesResponse,
  AcquiredEntity,
  AcquiredEntityId,
  AllOriginRelationshipsResponse,
  CreateAcquiredEntityRequest,
  CreateInternalTeamRequest,
  CreateVendorRequest,
  InternalTeam,
  InternalTeamId,
  InternalTeamsResponse,
  OriginRelationship,
  UpdateAcquiredEntityRequest,
  UpdateInternalTeamRequest,
  UpdateVendorRequest,
  Vendor,
  VendorId,
  VendorsResponse,
} from '../../../api/types';
import { followLink } from '../../../utils/hateoas';
import { linkOriginComponent } from './linkOriginComponent';

export const originEntitiesApi = {
  async getAllOriginRelationships(): Promise<AllOriginRelationshipsResponse> {
    const response = await httpClient.get<AllOriginRelationshipsResponse>('/api/v1/origin-relationships');
    return response.data;
  },

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

    async update(entity: AcquiredEntity, request: UpdateAcquiredEntityRequest): Promise<AcquiredEntity> {
      const response = await httpClient.put<AcquiredEntity>(followLink(entity, 'edit'), request);
      return response.data;
    },

    async delete(id: AcquiredEntityId): Promise<void> {
      await httpClient.delete(`/api/v1/acquired-entities/${id}`);
    },

    async linkComponent(
      componentId: string,
      acquiredEntityId: AcquiredEntityId,
      notes?: string,
    ): Promise<OriginRelationship> {
      return linkOriginComponent(componentId, 'acquired-via', { acquiredEntityId, notes });
    },

    async unlinkComponent(componentId: string): Promise<void> {
      await httpClient.delete(`/api/v1/components/${componentId}/origin/acquired-via`);
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

    async update(vendor: Vendor, request: UpdateVendorRequest): Promise<Vendor> {
      const response = await httpClient.put<Vendor>(followLink(vendor, 'edit'), request);
      return response.data;
    },

    async delete(id: VendorId): Promise<void> {
      await httpClient.delete(`/api/v1/vendors/${id}`);
    },

    async linkComponent(componentId: string, vendorId: VendorId, notes?: string): Promise<OriginRelationship> {
      return linkOriginComponent(componentId, 'purchased-from', { vendorId, notes });
    },

    async unlinkComponent(componentId: string): Promise<void> {
      await httpClient.delete(`/api/v1/components/${componentId}/origin/purchased-from`);
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

    async update(team: InternalTeam, request: UpdateInternalTeamRequest): Promise<InternalTeam> {
      const response = await httpClient.put<InternalTeam>(followLink(team, 'edit'), request);
      return response.data;
    },

    async delete(id: InternalTeamId): Promise<void> {
      await httpClient.delete(`/api/v1/internal-teams/${id}`);
    },

    async linkComponent(
      componentId: string,
      internalTeamId: InternalTeamId,
      notes?: string,
    ): Promise<OriginRelationship> {
      return linkOriginComponent(componentId, 'built-by', { internalTeamId, notes });
    },

    async unlinkComponent(componentId: string): Promise<void> {
      await httpClient.delete(`/api/v1/components/${componentId}/origin/built-by`);
    },
  },
};
