import { httpClient } from '../../../api/core';
import type {
  View,
  ViewId,
  ViewComponent,
  CreateViewRequest,
  AddComponentToViewRequest,
  AddCapabilityToViewRequest,
  UpdatePositionRequest,
  UpdateMultiplePositionsRequest,
  RenameViewRequest,
  UpdateViewEdgeTypeRequest,
  UpdateViewColorSchemeRequest,
  PaginatedResponse,
  CollectionResponse,
  ComponentId,
  CapabilityId,
  Position,
} from '../../../api/types';

export const viewsApi = {
  async getAll(): Promise<View[]> {
    const response = await httpClient.get<PaginatedResponse<View>>('/api/v1/views');
    return response.data.data || [];
  },

  async getById(id: ViewId): Promise<View> {
    const response = await httpClient.get<View>(`/api/v1/views/${id}`);
    return response.data;
  },

  async create(request: CreateViewRequest): Promise<View> {
    const response = await httpClient.post<View>('/api/v1/views', request);
    return response.data;
  },

  async delete(id: ViewId): Promise<void> {
    await httpClient.delete(`/api/v1/views/${id}`);
  },

  async rename(viewId: ViewId, request: RenameViewRequest): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/name`, request);
  },

  async setDefault(viewId: ViewId): Promise<void> {
    await httpClient.put(`/api/v1/views/${viewId}/default`);
  },

  async updateEdgeType(viewId: ViewId, request: UpdateViewEdgeTypeRequest): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/edge-type`, request);
  },

  async updateColorScheme(viewId: ViewId, request: UpdateViewColorSchemeRequest): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/color-scheme`, request);
  },

  async getComponents(viewId: ViewId): Promise<ViewComponent[]> {
    const response = await httpClient.get<CollectionResponse<ViewComponent>>(
      `/api/v1/views/${viewId}/components`
    );
    return response.data.data || [];
  },

  async addComponent(viewId: ViewId, request: AddComponentToViewRequest): Promise<void> {
    await httpClient.post(`/api/v1/views/${viewId}/components`, request);
  },

  async removeComponent(viewId: ViewId, componentId: ComponentId): Promise<void> {
    await httpClient.delete(`/api/v1/views/${viewId}/components/${componentId}`);
  },

  async updateComponentPosition(
    viewId: ViewId,
    componentId: ComponentId,
    request: UpdatePositionRequest
  ): Promise<void> {
    await httpClient.patch(
      `/api/v1/views/${viewId}/components/${componentId}/position`,
      request
    );
  },

  async updateComponentColor(viewId: ViewId, componentId: ComponentId, color: string): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/components/${componentId}/color`, { color });
  },

  async clearComponentColor(viewId: ViewId, componentId: ComponentId): Promise<void> {
    await httpClient.delete(`/api/v1/views/${viewId}/components/${componentId}/color`);
  },

  async updateMultiplePositions(viewId: ViewId, request: UpdateMultiplePositionsRequest): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/layout`, request);
  },

  async addCapability(viewId: ViewId, request: AddCapabilityToViewRequest): Promise<void> {
    await httpClient.post(`/api/v1/views/${viewId}/capabilities`, request);
  },

  async removeCapability(viewId: ViewId, capabilityId: CapabilityId): Promise<void> {
    await httpClient.delete(`/api/v1/views/${viewId}/capabilities/${capabilityId}`);
  },

  async updateCapabilityPosition(viewId: ViewId, capabilityId: CapabilityId, position: Position): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/capabilities/${capabilityId}/position`, position);
  },

  async updateCapabilityColor(viewId: ViewId, capabilityId: CapabilityId, color: string): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/capabilities/${capabilityId}/color`, { color });
  },

  async clearCapabilityColor(viewId: ViewId, capabilityId: CapabilityId): Promise<void> {
    await httpClient.delete(`/api/v1/views/${viewId}/capabilities/${capabilityId}/color`);
  },

  async changeVisibility(viewId: ViewId, isPrivate: boolean): Promise<void> {
    await httpClient.patch(`/api/v1/views/${viewId}/visibility`, { isPrivate });
  },
};

export default viewsApi;
