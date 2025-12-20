import { httpClient } from '../../../api/core';
import { ApiError } from '../../../api/types';
import type {
  LayoutContextType,
  LayoutContainer,
  LayoutContainerSummary,
  ElementPosition,
  UpsertLayoutRequest,
  ElementPositionInput,
  BatchUpdateItem,
  BatchUpdateResponse,
} from '../../../api/types';

export const layoutsApi = {
  async get(contextType: LayoutContextType, contextRef: string): Promise<LayoutContainer | null> {
    try {
      const response = await httpClient.get<LayoutContainer>(
        `/api/v1/layouts/${contextType}/${contextRef}`
      );
      return response.data;
    } catch (error) {
      if (error instanceof ApiError && error.statusCode === 404) {
        return null;
      }
      throw error;
    }
  },

  async upsert(
    contextType: LayoutContextType,
    contextRef: string,
    request: UpsertLayoutRequest = {}
  ): Promise<LayoutContainer> {
    const response = await httpClient.put<LayoutContainer>(
      `/api/v1/layouts/${contextType}/${contextRef}`,
      request
    );
    return response.data;
  },

  async delete(contextType: LayoutContextType, contextRef: string): Promise<void> {
    await httpClient.delete(`/api/v1/layouts/${contextType}/${contextRef}`);
  },

  async updatePreferences(
    contextType: LayoutContextType,
    contextRef: string,
    preferences: Record<string, unknown>,
    version: number
  ): Promise<LayoutContainerSummary> {
    const response = await httpClient.patch<LayoutContainerSummary>(
      `/api/v1/layouts/${contextType}/${contextRef}/preferences`,
      { preferences },
      {
        headers: {
          'If-Match': `"${version}"`,
        },
      }
    );
    return response.data;
  },

  async upsertElement(
    contextType: LayoutContextType,
    contextRef: string,
    elementId: string,
    position: ElementPositionInput
  ): Promise<ElementPosition> {
    const response = await httpClient.put<ElementPosition>(
      `/api/v1/layouts/${contextType}/${contextRef}/elements/${elementId}`,
      position
    );
    return response.data;
  },

  async deleteElement(
    contextType: LayoutContextType,
    contextRef: string,
    elementId: string
  ): Promise<void> {
    await httpClient.delete(
      `/api/v1/layouts/${contextType}/${contextRef}/elements/${elementId}`
    );
  },

  async batchUpdateElements(
    contextType: LayoutContextType,
    contextRef: string,
    updates: BatchUpdateItem[]
  ): Promise<BatchUpdateResponse> {
    const response = await httpClient.patch<BatchUpdateResponse>(
      `/api/v1/layouts/${contextType}/${contextRef}/elements`,
      { updates }
    );
    return response.data;
  },
};

export default layoutsApi;
