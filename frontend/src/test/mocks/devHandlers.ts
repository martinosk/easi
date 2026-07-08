import { HttpResponse, http } from 'msw';

const BASE_URL = '*';

export const devHandlers = [
  http.get(`${BASE_URL}/api/v1/auth/sessions/current`, () => {
    return HttpResponse.json({
      id: 'dev-session',
      user: {
        id: 'dev-user',
        email: 'architect@dev.local',
        name: 'Dev Architect',
        role: 'architect',
        permissions: [
          'enterprise-arch:read',
          'enterprise-arch:write',
          'enterprise-arch:delete',
          'architecture-direction:read',
          'architecture-direction:write',
          'valuestreams:read',
          'valuestreams:write',
          'valuestreams:delete',
        ],
      },
      tenant: { id: 'dev-tenant', name: 'Dev Tenant' },
      expiresAt: '2030-01-01T00:00:00Z',
      _links: {
        self: '/api/v1/auth/sessions/current',
        logout: '/api/v1/auth/sessions/current',
        user: '/api/v1/users/dev-user',
        tenant: '/api/v1/tenants/dev-tenant',
      },
    });
  }),
];
