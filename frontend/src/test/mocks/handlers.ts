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
    const view = getView(params.id as ViewId);
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
        color: body.color,
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
      const rest = Object.fromEntries(Object.entries(view.capabilities[capIndex]).filter(([key]) => key !== 'color')) as typeof view.capabilities[number];
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
        color: body.color,
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
      const rest = Object.fromEntries(Object.entries(view.components[compIndex]).filter(([key]) => key !== 'color')) as typeof view.components[number];
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
];
