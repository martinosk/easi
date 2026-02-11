import { http, HttpResponse } from 'msw';
import {
  getComponents,
  getComponent,
  addComponent,
  getCapabilities,
  getCapability,
  addCapability,
  getCapabilityRealizations,
  getRealizationsByCapability,
  getRealizationsByComponent,
  getViews,
  getView,
  updateView,
  getRelations,
  addRelation,
} from './db';
import { toComponentId, toCapabilityId, toViewId } from '../../api/types';

const BASE_URL = 'http://localhost:8080';

export const handlers = [
  http.get(`${BASE_URL}/api/v1/components`, () => {
    return HttpResponse.json({
      data: getComponents(),
      _links: { self: '/api/v1/components' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/components/:id`, ({ params }) => {
    const component = getComponent(toComponentId(params.id as string));
    if (!component) {
      return new HttpResponse(null, { status: 404 });
    }
    return HttpResponse.json(component);
  }),

  http.post(`${BASE_URL}/api/v1/components`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>;
    const component = addComponent(body);
    return HttpResponse.json(component, { status: 201 });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities`, () => {
    return HttpResponse.json({
      data: getCapabilities(),
      _links: { self: '/api/v1/capabilities' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/:id`, ({ params }) => {
    const capability = getCapability(toCapabilityId(params.id as string));
    if (!capability) {
      return new HttpResponse(null, { status: 404 });
    }
    return HttpResponse.json(capability);
  }),

  http.post(`${BASE_URL}/api/v1/capabilities`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>;
    const capability = addCapability(body);
    return HttpResponse.json(capability, { status: 201 });
  }),

  http.put(`${BASE_URL}/api/v1/capabilities/:id`, async ({ params, request }) => {
    const capability = getCapability(toCapabilityId(params.id as string));
    if (!capability) {
      return new HttpResponse(null, { status: 404 });
    }
    const body = await request.json() as Record<string, unknown>;
    const updated = { ...capability, ...body };
    return HttpResponse.json(updated);
  }),

  http.put(`${BASE_URL}/api/v1/capabilities/:id/metadata`, async ({ params, request }) => {
    const capability = getCapability(toCapabilityId(params.id as string));
    if (!capability) {
      return new HttpResponse(null, { status: 404 });
    }
    const body = await request.json() as Record<string, unknown>;
    const updated = { ...capability, ...body };
    return HttpResponse.json(updated);
  }),

  http.delete(`${BASE_URL}/api/v1/capabilities/:id`, ({ params }) => {
    const capability = getCapability(toCapabilityId(params.id as string));
    if (!capability) {
      return new HttpResponse(null, { status: 404 });
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/:id/systems`, ({ params }) => {
    const realizations = getRealizationsByCapability(toCapabilityId(params.id as string));
    return HttpResponse.json({
      data: realizations,
      _links: { self: `/api/v1/capabilities/${params.id}/systems` },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capability-realizations/by-component/:componentId`, ({ params }) => {
    const realizations = getRealizationsByComponent(toComponentId(params.componentId as string));
    return HttpResponse.json({
      data: realizations,
      _links: { self: `/api/v1/capability-realizations/by-component/${params.componentId}` },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capability-realizations`, () => {
    return HttpResponse.json({
      data: getCapabilityRealizations(),
      _links: { self: '/api/v1/capability-realizations' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/views`, () => {
    return HttpResponse.json({
      data: getViews(),
      _links: { self: '/api/v1/views' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/views/:id`, ({ params }) => {
    const view = getView(toViewId(params.id as string));
    if (!view) {
      return new HttpResponse(null, { status: 404 });
    }
    return HttpResponse.json(view);
  }),

  http.patch(`${BASE_URL}/api/v1/views/:viewId/capabilities/:capabilityId/color`, async ({ params, request }) => {
    const view = getView(toViewId(params.viewId as string));
    if (!view) {
      return new HttpResponse(null, { status: 404 });
    }
    const body = await request.json() as { color: string };
    const capIndex = view.capabilities?.findIndex(
      (c) => c.capabilityId === params.capabilityId
    ) ?? -1;
    if (capIndex >= 0 && view.capabilities) {
      view.capabilities[capIndex] = {
        ...view.capabilities[capIndex],
        customColor: body.color,
      };
      updateView(toViewId(params.viewId as string), { capabilities: view.capabilities });
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.delete(`${BASE_URL}/api/v1/views/:viewId/capabilities/:capabilityId/color`, ({ params }) => {
    const view = getView(toViewId(params.viewId as string));
    if (!view) {
      return new HttpResponse(null, { status: 404 });
    }
    const capIndex = view.capabilities?.findIndex(
      (c) => c.capabilityId === params.capabilityId
    ) ?? -1;
    if (capIndex >= 0 && view.capabilities) {
      const rest = Object.fromEntries(Object.entries(view.capabilities[capIndex]).filter(([key]) => key !== 'customColor')) as typeof view.capabilities[number];
      view.capabilities[capIndex] = rest as typeof view.capabilities[number];
      updateView(toViewId(params.viewId as string), { capabilities: view.capabilities });
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.patch(`${BASE_URL}/api/v1/views/:viewId/components/:componentId/color`, async ({ params, request }) => {
    const view = getView(toViewId(params.viewId as string));
    if (!view) {
      return new HttpResponse(null, { status: 404 });
    }
    const body = await request.json() as { color: string };
    const compIndex = view.components?.findIndex(
      (c) => c.componentId === params.componentId
    ) ?? -1;
    if (compIndex >= 0 && view.components) {
      view.components[compIndex] = {
        ...view.components[compIndex],
        customColor: body.color,
      };
      updateView(toViewId(params.viewId as string), { components: view.components });
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.delete(`${BASE_URL}/api/v1/views/:viewId/components/:componentId/color`, ({ params }) => {
    const view = getView(toViewId(params.viewId as string));
    if (!view) {
      return new HttpResponse(null, { status: 404 });
    }
    const compIndex = view.components?.findIndex(
      (c) => c.componentId === params.componentId
    ) ?? -1;
    if (compIndex >= 0 && view.components) {
      const rest = Object.fromEntries(Object.entries(view.components[compIndex]).filter(([key]) => key !== 'customColor')) as typeof view.components[number];
      view.components[compIndex] = rest as typeof view.components[number];
      updateView(toViewId(params.viewId as string), { components: view.components });
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${BASE_URL}/api/v1/relations`, () => {
    return HttpResponse.json({
      data: getRelations(),
      _links: { self: '/api/v1/relations' },
    });
  }),

  http.post(`${BASE_URL}/api/v1/relations`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>;
    const relation = addRelation(body);
    return HttpResponse.json(relation, { status: 201 });
  }),

  http.get(`${BASE_URL}/api/v1/value-streams`, () => {
    return HttpResponse.json({
      data: [],
      _links: {
        self: { href: '/api/v1/value-streams', method: 'GET' },
        'x-create': { href: '/api/v1/value-streams', method: 'POST' },
      },
    });
  }),

  http.post(`${BASE_URL}/api/v1/value-streams`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>;
    return HttpResponse.json({
      id: 'test-vs-id',
      name: body.name,
      description: body.description || '',
      stageCount: 0,
      createdAt: new Date().toISOString(),
      _links: {
        self: { href: '/api/v1/value-streams/test-vs-id', method: 'GET' },
        edit: { href: '/api/v1/value-streams/test-vs-id', method: 'PUT' },
        delete: { href: '/api/v1/value-streams/test-vs-id', method: 'DELETE' },
        collection: { href: '/api/v1/value-streams', method: 'GET' },
      },
    }, { status: 201 });
  }),

  http.put(`${BASE_URL}/api/v1/value-streams/:id`, async ({ params, request }) => {
    const body = await request.json() as Record<string, unknown>;
    return HttpResponse.json({
      id: params.id,
      name: body.name,
      description: body.description || '',
      stageCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      _links: {
        self: { href: `/api/v1/value-streams/${params.id}`, method: 'GET' },
        edit: { href: `/api/v1/value-streams/${params.id}`, method: 'PUT' },
        delete: { href: `/api/v1/value-streams/${params.id}`, method: 'DELETE' },
        collection: { href: '/api/v1/value-streams', method: 'GET' },
      },
    });
  }),

  http.delete(`${BASE_URL}/api/v1/value-streams/:id`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${BASE_URL}/api/v1/business-domains`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/business-domains' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capability-dependencies`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/capability-dependencies' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/:id/importance`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/capabilities' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/metadata/statuses`, () => {
    return HttpResponse.json({
      data: [
        { value: 'Active', displayName: 'Active', sortOrder: 1 },
        { value: 'Inactive', displayName: 'Inactive', sortOrder: 2 },
      ],
    });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/metadata/ownership-models`, () => {
    return HttpResponse.json({ data: [] });
  }),

  http.get(`${BASE_URL}/api/v1/meta-model/maturity-scale`, () => {
    return HttpResponse.json({
      sections: [
        { name: 'Genesis', order: 1, minValue: 0, maxValue: 25 },
        { name: 'Custom Build', order: 2, minValue: 25, maxValue: 50 },
        { name: 'Product', order: 3, minValue: 50, maxValue: 75 },
        { name: 'Commodity', order: 4, minValue: 75, maxValue: 100 },
      ],
      version: 1,
      isDefault: true,
      _links: { self: { href: '/api/v1/meta-model/maturity-scale', method: 'GET' } },
    });
  }),

  http.get(`${BASE_URL}/api/v1/meta-model/strategy-pillars`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/meta-model/strategy-pillars' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/artifact-creators`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: { href: '/api/v1/artifact-creators', method: 'GET' } },
    });
  }),

  http.get(`${BASE_URL}/api/v1/audit/:id`, () => {
    return HttpResponse.json({
      entries: [],
      _links: { self: '/api/v1/audit' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/users`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${BASE_URL}/api/v1/components/:id/fit-scores`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/components' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/components/:id/origins`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/components' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/acquired-entities`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/acquired-entities' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/internal-teams`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/internal-teams' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/vendors`, () => {
    return HttpResponse.json({
      data: [],
      _links: { self: '/api/v1/vendors' },
    });
  }),

  http.get(`${BASE_URL}/api/v1/origin-relationships`, () => {
    return HttpResponse.json({
      acquiredVia: [],
      purchasedFrom: [],
      builtBy: [],
      _links: { self: { href: '/api/v1/origin-relationships', method: 'GET' } },
    });
  }),
];
