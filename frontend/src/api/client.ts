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
} from './types';
import { ApiError } from './types';

class ApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string = 'http://localhost:8080') {
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
      const data = error.response.data as { message?: string; error?: string };
      return data.message || data.error || 'An error occurred';
    }
    if (error.message) {
      return error.message;
    }
    return 'An unknown error occurred';
  }

  // Components API
  async getComponents(): Promise<Component[]> {
    const response = await this.client.get<{ data: Component[] | null }>('/api/v1/components');
    return response.data.data || [];
  }

  async getComponentById(id: string): Promise<Component> {
    const response = await this.client.get<{ data: Component }>(`/api/v1/components/${id}`);
    return response.data.data;
  }

  async createComponent(request: CreateComponentRequest): Promise<Component> {
    const response = await this.client.post<{ data: Component }>('/api/v1/components', request);
    return response.data.data;
  }

  // Relations API
  async getRelations(): Promise<Relation[]> {
    const response = await this.client.get<{ data: Relation[] | null }>('/api/v1/relations');
    return response.data.data || [];
  }

  async getRelationById(id: string): Promise<Relation> {
    const response = await this.client.get<{ data: Relation }>(`/api/v1/relations/${id}`);
    return response.data.data;
  }

  async createRelation(request: CreateRelationRequest): Promise<Relation> {
    const response = await this.client.post<{ data: Relation }>('/api/v1/relations', request);
    return response.data.data;
  }

  // Views API
  async getViews(): Promise<View[]> {
    const response = await this.client.get<{ data: View[] | null }>('/api/v1/views');
    return response.data.data || [];
  }

  async getViewById(id: string): Promise<View> {
    const response = await this.client.get<{ data: View }>(`/api/v1/views/${id}`);
    return response.data.data;
  }

  async createView(request: CreateViewRequest): Promise<View> {
    const response = await this.client.post<{ data: View }>('/api/v1/views', request);
    return response.data.data;
  }

  async getViewComponents(viewId: string): Promise<ViewComponent[]> {
    const response = await this.client.get<{ data: ViewComponent[] | null }>(
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
}

export const apiClient = new ApiClient();
export default apiClient;
