import axios, { type AxiosError, type AxiosInstance } from 'axios';
import type {
  Component,
  Relation,
  View,
  ViewComponent,
  CreateComponentRequest,
  CreateRelationRequest,
  CreateViewRequest,
  AddComponentToViewRequest,
  UpdatePositionRequest,
  UpdateMultiplePositionsRequest,
  RenameViewRequest,
  UpdateViewEdgeTypeRequest,
  UpdateViewLayoutDirectionRequest,
  PaginatedResponse,
  CollectionResponse,
  Capability,
  CapabilityDependency,
  CapabilityRealization,
  CreateCapabilityRequest,
  UpdateCapabilityRequest,
  UpdateCapabilityMetadataRequest,
  AddCapabilityExpertRequest,
  AddCapabilityTagRequest,
  CreateCapabilityDependencyRequest,
  LinkSystemToCapabilityRequest,
  UpdateRealizationRequest,
} from './types';
import { ApiError } from './types';

class ApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string = import.meta.env.VITE_API_URL || 'http://localhost:8080') {
    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        const message = this.extractErrorMessage(error);
        const statusCode = error.response?.status || 500;
        throw new ApiError(message, statusCode, error.response?.data);
      }
    );
  }

  private extractErrorMessage(error: AxiosError): string {
    const responseMessage = this.extractResponseMessage(error.response?.data);
    return responseMessage ?? error.message ?? 'An unknown error occurred';
  }

  private extractResponseMessage(data: unknown): string | null {
    if (!data || typeof data !== 'object') return null;

    const errorData = data as { message?: string; error?: string; details?: Record<string, string> };
    if (errorData.message) return errorData.message;
    if (errorData.error) return errorData.error;

    const detailMessages = errorData.details ? Object.values(errorData.details).filter(Boolean) : [];
    return detailMessages.length > 0 ? detailMessages.join(', ') : 'An error occurred';
  }

  async getComponents(): Promise<Component[]> {
    const response = await this.client.get<PaginatedResponse<Component>>('/api/v1/components');
    return response.data.data || [];
  }

  async getComponentById(id: string): Promise<Component> {
    const response = await this.client.get<Component>(`/api/v1/components/${id}`);
    return response.data;
  }

  async createComponent(request: CreateComponentRequest): Promise<Component> {
    const response = await this.client.post<Component>('/api/v1/components', request);
    return response.data;
  }

  async updateComponent(id: string, request: CreateComponentRequest): Promise<Component> {
    const response = await this.client.put<Component>(`/api/v1/components/${id}`, request);
    return response.data;
  }

  async deleteComponent(id: string): Promise<void> {
    await this.client.delete(`/api/v1/components/${id}`);
  }

  async getRelations(): Promise<Relation[]> {
    const response = await this.client.get<PaginatedResponse<Relation>>('/api/v1/relations');
    return response.data.data || [];
  }

  async getRelationById(id: string): Promise<Relation> {
    const response = await this.client.get<Relation>(`/api/v1/relations/${id}`);
    return response.data;
  }

  async createRelation(request: CreateRelationRequest): Promise<Relation> {
    const response = await this.client.post<Relation>('/api/v1/relations', request);
    return response.data;
  }

  async updateRelation(id: string, request: Partial<CreateRelationRequest>): Promise<Relation> {
    const response = await this.client.put<Relation>(`/api/v1/relations/${id}`, request);
    return response.data;
  }

  async deleteRelation(id: string): Promise<void> {
    await this.client.delete(`/api/v1/relations/${id}`);
  }

  async getViews(): Promise<View[]> {
    const response = await this.client.get<PaginatedResponse<View>>('/api/v1/views');
    return response.data.data || [];
  }

  async getViewById(id: string): Promise<View> {
    const response = await this.client.get<View>(`/api/v1/views/${id}`);
    return response.data;
  }

  async createView(request: CreateViewRequest): Promise<View> {
    const response = await this.client.post<View>('/api/v1/views', request);
    return response.data;
  }

  async getViewComponents(viewId: string): Promise<ViewComponent[]> {
    const response = await this.client.get<CollectionResponse<ViewComponent>>(
      `/api/v1/views/${viewId}/components`
    );
    return response.data.data || [];
  }

  async addComponentToView(
    viewId: string,
    request: AddComponentToViewRequest
  ): Promise<void> {
    await this.client.post(`/api/v1/views/${viewId}/components`, request);
  }

  async updateComponentPosition(
    viewId: string,
    componentId: string,
    request: UpdatePositionRequest
  ): Promise<void> {
    await this.client.patch(
      `/api/v1/views/${viewId}/components/${componentId}/position`,
      request
    );
  }

  async updateMultiplePositions(
    viewId: string,
    request: UpdateMultiplePositionsRequest
  ): Promise<void> {
    await this.client.patch(
      `/api/v1/views/${viewId}/layout`,
      request
    );
  }

  async renameView(viewId: string, request: RenameViewRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/name`, request);
  }

  async deleteView(viewId: string): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}`);
  }

  async removeComponentFromView(viewId: string, componentId: string): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}/components/${componentId}`);
  }

  async setDefaultView(viewId: string): Promise<void> {
    await this.client.put(`/api/v1/views/${viewId}/default`);
  }

  async updateViewEdgeType(viewId: string, request: UpdateViewEdgeTypeRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/edge-type`, request);
  }

  async updateViewLayoutDirection(viewId: string, request: UpdateViewLayoutDirectionRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/layout-direction`, request);
  }

  async getCapabilities(): Promise<Capability[]> {
    const response = await this.client.get<CollectionResponse<Capability>>('/api/v1/capabilities');
    return response.data.data || [];
  }

  async getCapabilityById(id: string): Promise<Capability> {
    const response = await this.client.get<Capability>(`/api/v1/capabilities/${id}`);
    return response.data;
  }

  async getCapabilityChildren(id: string): Promise<Capability[]> {
    const response = await this.client.get<CollectionResponse<Capability>>(`/api/v1/capabilities/${id}/children`);
    return response.data.data || [];
  }

  async createCapability(request: CreateCapabilityRequest): Promise<Capability> {
    const response = await this.client.post<Capability>('/api/v1/capabilities', request);
    return response.data;
  }

  async updateCapability(id: string, request: UpdateCapabilityRequest): Promise<Capability> {
    const response = await this.client.put<Capability>(`/api/v1/capabilities/${id}`, request);
    return response.data;
  }

  async updateCapabilityMetadata(id: string, request: UpdateCapabilityMetadataRequest): Promise<Capability> {
    const response = await this.client.put<Capability>(`/api/v1/capabilities/${id}/metadata`, request);
    return response.data;
  }

  async addCapabilityExpert(id: string, request: AddCapabilityExpertRequest): Promise<void> {
    await this.client.post(`/api/v1/capabilities/${id}/experts`, request);
  }

  async addCapabilityTag(id: string, request: AddCapabilityTagRequest): Promise<void> {
    await this.client.post(`/api/v1/capabilities/${id}/tags`, request);
  }

  async getCapabilityDependencies(): Promise<CapabilityDependency[]> {
    const response = await this.client.get<CollectionResponse<CapabilityDependency>>('/api/v1/capability-dependencies');
    return response.data.data || [];
  }

  async getOutgoingDependencies(capabilityId: string): Promise<CapabilityDependency[]> {
    const response = await this.client.get<CollectionResponse<CapabilityDependency>>(
      `/api/v1/capabilities/${capabilityId}/dependencies/outgoing`
    );
    return response.data.data || [];
  }

  async getIncomingDependencies(capabilityId: string): Promise<CapabilityDependency[]> {
    const response = await this.client.get<CollectionResponse<CapabilityDependency>>(
      `/api/v1/capabilities/${capabilityId}/dependencies/incoming`
    );
    return response.data.data || [];
  }

  async createCapabilityDependency(request: CreateCapabilityDependencyRequest): Promise<CapabilityDependency> {
    const response = await this.client.post<CapabilityDependency>('/api/v1/capability-dependencies', request);
    return response.data;
  }

  async deleteCapabilityDependency(id: string): Promise<void> {
    await this.client.delete(`/api/v1/capability-dependencies/${id}`);
  }

  async getSystemsByCapability(capabilityId: string): Promise<CapabilityRealization[]> {
    const response = await this.client.get<CollectionResponse<CapabilityRealization>>(
      `/api/v1/capabilities/${capabilityId}/systems`
    );
    return response.data.data || [];
  }

  async getCapabilitiesByComponent(componentId: string): Promise<CapabilityRealization[]> {
    const response = await this.client.get<CollectionResponse<CapabilityRealization>>(
      `/api/v1/capability-realizations/by-component/${componentId}`
    );
    return response.data.data || [];
  }

  async linkSystemToCapability(capabilityId: string, request: LinkSystemToCapabilityRequest): Promise<CapabilityRealization> {
    const response = await this.client.post<CapabilityRealization>(
      `/api/v1/capabilities/${capabilityId}/systems`,
      request
    );
    return response.data;
  }

  async updateRealization(id: string, request: UpdateRealizationRequest): Promise<CapabilityRealization> {
    const response = await this.client.put<CapabilityRealization>(
      `/api/v1/capability-realizations/${id}`,
      request
    );
    return response.data;
  }

  async deleteRealization(id: string): Promise<void> {
    await this.client.delete(`/api/v1/capability-realizations/${id}`);
  }
}

export const apiClient = new ApiClient();
export default apiClient;
