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
    if (error.response?.data) {
      const data = error.response.data as { message?: string; error?: string; details?: Record<string, string> };
      if (data.message) {
        return data.message;
      }
      if (data.error) {
        return data.error;
      }
      if (data.details) {
        const detailMessages = Object.values(data.details).filter(Boolean);
        if (detailMessages.length > 0) {
          return detailMessages.join(', ');
        }
      }
      return 'An error occurred';
    }
    if (error.message) {
      return error.message;
    }
    return 'An unknown error occurred';
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
}

export const apiClient = new ApiClient();
export default apiClient;
