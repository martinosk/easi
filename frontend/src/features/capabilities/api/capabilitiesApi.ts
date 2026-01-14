import { httpClient } from '../../../api/core';
import { followLink } from '../../../utils/hateoas';
import type {
  Capability,
  CapabilityId,
  CapabilityDependency,
  CapabilityRealization,
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

  async update(capability: Capability, request: UpdateCapabilityRequest): Promise<Capability> {
    const response = await httpClient.put<Capability>(followLink(capability, 'edit'), request);
    return response.data;
  },

  async updateMetadata(id: CapabilityId, request: UpdateCapabilityMetadataRequest): Promise<Capability> {
    const response = await httpClient.put<Capability>(`/api/v1/capabilities/${id}/metadata`, request);
    return response.data;
  },

  async addExpert(id: CapabilityId, request: AddCapabilityExpertRequest): Promise<void> {
    await httpClient.post(`/api/v1/capabilities/${id}/experts`, request);
  },

  async removeExpert(
    id: CapabilityId,
    expert: { name: string; role: string; contact: string }
  ): Promise<void> {
    const params = new URLSearchParams({
      name: expert.name,
      role: expert.role,
      contact: expert.contact,
    });
    await httpClient.delete(`/api/v1/capabilities/${id}/experts?${params.toString()}`);
  },

  async getExpertRoles(): Promise<string[]> {
    const response = await httpClient.get<{ roles: string[] }>('/api/v1/capabilities/expert-roles');
    return response.data.roles || [];
  },

  async addTag(id: CapabilityId, request: AddCapabilityTagRequest): Promise<void> {
    await httpClient.post(`/api/v1/capabilities/${id}/tags`, request);
  },

  async delete(capability: Capability): Promise<void> {
    await httpClient.delete(followLink(capability, 'delete'));
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

  async deleteDependency(dependency: CapabilityDependency): Promise<void> {
    await httpClient.delete(followLink(dependency, 'delete'));
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

  async updateRealization(realization: CapabilityRealization, request: UpdateRealizationRequest): Promise<CapabilityRealization> {
    const response = await httpClient.put<CapabilityRealization>(
      followLink(realization, 'edit'),
      request
    );
    return response.data;
  },

  async deleteRealization(realization: CapabilityRealization): Promise<void> {
    await httpClient.delete(followLink(realization, 'delete'));
  },
};

export default capabilitiesApi;
