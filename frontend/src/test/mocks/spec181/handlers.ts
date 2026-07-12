import { HttpResponse, http } from 'msw';
import type { HttpMethod } from '../../../api/types';
import type { RealizationRole, RealizationRoleAssignment } from '../../../features/architecture-direction/types';
import {
  canWriteRoles,
  findRole,
  removeRole,
  rolesForCapabilities,
  type StubRealizationRole,
  upsertRole,
} from './store';

const BASE_URL = '*';
const link = (href: string, method: HttpMethod) => ({ href, method });

function rolePath(capabilityId: string, componentId: string): string {
  return `/api/v1/capabilities/${capabilityId}/components/${componentId}/realization-role`;
}

function toDto(r: StubRealizationRole): RealizationRoleAssignment {
  const base = rolePath(r.capabilityId, r.componentId);
  return {
    capabilityId: r.capabilityId,
    capabilityName: r.capabilityName,
    componentId: r.componentId,
    componentName: r.componentName,
    role: r.role,
    assignedBy: r.assignedBy,
    assignedAt: r.assignedAt,
    _links: {
      self: link(base, 'GET'),
      ...(canWriteRoles() ? { edit: link(base, 'PUT'), delete: link(base, 'DELETE') } : {}),
    },
  };
}

function splitIds(raw: string | null): string[] {
  return (raw ?? '').split(',').filter(Boolean);
}

export const spec181Handlers = [
  http.get(`${BASE_URL}/api/v1/realization-roles`, ({ request }) => {
    const url = new URL(request.url);
    const capabilityIds = splitIds(url.searchParams.get('capabilityIds'));
    const data = rolesForCapabilities(capabilityIds).map(toDto);

    return HttpResponse.json({
      data,
      _links: {
        self: link(url.pathname + url.search, 'GET'),
        ...(canWriteRoles()
          ? { 'x-assign': link('/api/v1/capabilities/{id}/components/{componentId}/realization-role', 'PUT') }
          : {}),
      },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/:id/components/:componentId/realization-role`, ({ params }) => {
    const found = findRole(params.id as string, params.componentId as string);
    if (!found) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(toDto(found));
  }),

  http.put(
    `${BASE_URL}/api/v1/capabilities/:id/components/:componentId/realization-role`,
    async ({ params, request }) => {
      const capabilityId = params.id as string;
      const componentId = params.componentId as string;
      const body = (await request.json()) as { role: RealizationRole };
      const existing = findRole(capabilityId, componentId);

      const record: StubRealizationRole = {
        capabilityId,
        capabilityName: existing?.capabilityName ?? 'Test Capability',
        componentId,
        componentName: existing?.componentName ?? 'Test Component',
        role: body.role,
        assignedBy: 'user-1',
        assignedAt: new Date().toISOString(),
      };
      upsertRole(record);

      return HttpResponse.json(toDto(record), { status: existing ? 200 : 201 });
    },
  ),

  http.delete(`${BASE_URL}/api/v1/capabilities/:id/components/:componentId/realization-role`, ({ params }) => {
    removeRole(params.id as string, params.componentId as string);
    return new HttpResponse(null, { status: 204 });
  }),
];
