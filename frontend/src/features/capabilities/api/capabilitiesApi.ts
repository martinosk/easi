import { httpClient } from '../../../api/core';
import type {
  Capability,
  CapabilityId,
  CapabilityDependency,
  CapabilityDependencyId,
  CapabilityRealization,
  RealizationId,
  ComponentId,
  CreateCapabilityRequest,
  UpdateCapabilityRequest,
  UpdateCapabilityMetadataRequest,
  AddCapabilityExpertRequest,
  AddCapabilityTagRequest,
  CreateCapabilityDependencyRequest,
  LinkSystemToCapabilityRequest,
  UpdateRealizationRequest,
  CollectionResponse,
} from '../../../api/types';

export const capabilitiesApi = {
  async getAll(): Promise<Capability[]> {
    const response = await httpClient.get<CollectionResponse<Capability>>('/api/v1/capabilities');
    return response.data.data || [];
  },

  async getById(id: CapabilityId): Promise<Capability> {
    const response = await httpClient.get<Capability>(`/api/v1/capabilities/${id}`);
    return response.data;
  },

  async getChildren(id: CapabilityId): Promise<Capability[]> {
    const response = await httpClient.get<CollectionResponse<Capability>>(`/api/v1/capabilities/${id}/children`);
    return response.data.data || [];
  },

  async create(request: CreateCapabilityRequest): Promise<Capability> {
    const response = await httpClient.post<Capability>('/api/v1/capabilities', request);
    return response.data;
  },

  async update(id: CapabilityId, request: UpdateCapabilityRequest): Promise<Capability> {
    const response = await httpClient.put<Capability>(`/api/v1/capabilities/${id}`, request);
    return response.data;
  },

  async updateMetadata(id: CapabilityId, request: UpdateCapabilityMetadataRequest): Promise<Capability> {
    const response = await httpClient.put<Capability>(`/api/v1/capabilities/${id}/metadata`, request);
    return response.data;
  },

  async addExpert(id: CapabilityId, request: AddCapabilityExpertRequest): Promise<void> {
    await httpClient.post(`/api/v1/capabilities/${id}/experts`, request);
  },

  async addTag(id: CapabilityId, request: AddCapabilityTagRequest): Promise<void> {
    await httpClient.post(`/api/v1/capabilities/${id}/tags`, request);
  },

  async delete(id: CapabilityId): Promise<void> {
    await httpClient.delete(`/api/v1/capabilities/${id}`);
  },

  async changeParent(id: CapabilityId, parentId: CapabilityId | null): Promise<void> {
    await httpClient.patch(`/api/v1/capabilities/${id}/parent`, {
      parentId: parentId || '',
    });
  },

  async getAllDependencies(): Promise<CapabilityDependency[]> {
    const response = await httpClient.get<CollectionResponse<CapabilityDependency>>('/api/v1/capability-dependencies');
    return response.data.data || [];
  },

  async getOutgoingDependencies(capabilityId: CapabilityId): Promise<CapabilityDependency[]> {
    const response = await httpClient.get<CollectionResponse<CapabilityDependency>>(
      `/api/v1/capabilities/${capabilityId}/dependencies/outgoing`
    );
    return response.data.data || [];
  },

  async getIncomingDependencies(capabilityId: CapabilityId): Promise<CapabilityDependency[]> {
    const response = await httpClient.get<CollectionResponse<CapabilityDependency>>(
      `/api/v1/capabilities/${capabilityId}/dependencies/incoming`
    );
    return response.data.data || [];
  },

  async createDependency(request: CreateCapabilityDependencyRequest): Promise<CapabilityDependency> {
    const response = await httpClient.post<CapabilityDependency>('/api/v1/capability-dependencies', request);
    return response.data;
  },

  async deleteDependency(id: CapabilityDependencyId): Promise<void> {
    await httpClient.delete(`/api/v1/capability-dependencies/${id}`);
  },

  async getSystemsByCapability(capabilityId: CapabilityId): Promise<CapabilityRealization[]> {
    const response = await httpClient.get<CollectionResponse<CapabilityRealization>>(
      `/api/v1/capabilities/${capabilityId}/systems`
    );
    return response.data.data || [];
  },

  async getCapabilitiesByComponent(componentId: ComponentId): Promise<CapabilityRealization[]> {
    const response = await httpClient.get<CollectionResponse<CapabilityRealization>>(
      `/api/v1/capability-realizations/by-component/${componentId}`
    );
    return response.data.data || [];
  },

  async linkSystem(capabilityId: CapabilityId, request: LinkSystemToCapabilityRequest): Promise<CapabilityRealization> {
    const response = await httpClient.post<CapabilityRealization>(
      `/api/v1/capabilities/${capabilityId}/systems`,
      request
    );
    return response.data;
  },

  async updateRealization(id: RealizationId, request: UpdateRealizationRequest): Promise<CapabilityRealization> {
    const response = await httpClient.put<CapabilityRealization>(
      `/api/v1/capability-realizations/${id}`,
      request
    );
    return response.data;
  },

  async deleteRealization(id: RealizationId): Promise<void> {
    await httpClient.delete(`/api/v1/capability-realizations/${id}`);
  },
};

export default capabilitiesApi;
