import { httpClient } from '../../../api/core/httpClient';
import { ApiError, type CapabilityId, type ComponentId } from '../../../api/types';
import { followLink } from '../../../utils/hateoas';
import type { AssignRealizationRoleRequest, RealizationRoleAssignment, RealizationRolesResponse } from '../types';

const rolePath = (capabilityId: CapabilityId | string, componentId: ComponentId | string) =>
  `/api/v1/capabilities/${capabilityId}/components/${componentId}/realization-role`;

export const realizationRoleApi = {
  async getByCapabilityIds(capabilityIds: (CapabilityId | string)[]): Promise<RealizationRolesResponse> {
    const response = await httpClient.get<RealizationRolesResponse>(
      `/api/v1/realization-roles?capabilityIds=${capabilityIds.join(',')}`,
    );
    return response.data;
  },

  async getAll(): Promise<RealizationRolesResponse> {
    const response = await httpClient.get<RealizationRolesResponse>('/api/v1/realization-roles');
    return response.data;
  },

  async getOne(
    capabilityId: CapabilityId | string,
    componentId: ComponentId | string,
  ): Promise<RealizationRoleAssignment | null> {
    try {
      const response = await httpClient.get<RealizationRoleAssignment>(rolePath(capabilityId, componentId));
      return response.data;
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 404) return null;
      throw err;
    }
  },

  async assign(
    capabilityId: CapabilityId | string,
    componentId: ComponentId | string,
    request: AssignRealizationRoleRequest,
  ): Promise<RealizationRoleAssignment> {
    const response = await httpClient.put<RealizationRoleAssignment>(rolePath(capabilityId, componentId), request);
    return response.data;
  },

  async clear(role: RealizationRoleAssignment): Promise<void> {
    await httpClient.delete(followLink(role, 'delete'));
  },
};
