import { HttpResponse, http } from 'msw';
import { toCapabilityId, toComponentId } from '../../../api/types';
import type {
  AddJourneyMilestoneRequest,
  CaptureJourneyRequest,
  ChangeJourneySourceApplicationsRequest,
  JourneyMilestone,
  UpdateJourneyDetailsRequest,
  UpdateJourneyMilestoneRequest,
  UpdateJourneyProgressRequest,
} from '../../../features/journeys/types';
import type { CapabilityJourney } from '../../../features/journeys/types';
import { getCapability, getComponent } from '../db';
import {
  addJourney,
  canWriteJourneys,
  findActiveJourneyForCapability,
  findJourney,
  getCurrentByCapabilityIds,
  getHistoryForCapability,
  isActiveStatus,
  nextJourneyId,
  nextMilestoneId,
  type StubApplicationRef,
  type StubJourney,
  type StubMilestone,
} from './store';

const BASE_URL = '*';
const link = (href: string, method: 'GET' | 'POST' | 'PUT' | 'DELETE') => ({ href, method });

function splitIds(raw: string | null): string[] {
  return (raw ?? '').split(',').filter(Boolean);
}

function milestoneDto(milestone: StubMilestone, journeyId: string, writable: boolean): JourneyMilestone {
  const base = `/api/v1/capability-journeys/${journeyId}/milestones/${milestone.id}`;
  return {
    id: milestone.id,
    label: milestone.label,
    targetPeriod: milestone.targetPeriod,
    status: milestone.status,
    _links: writable ? { edit: link(base, 'PUT'), delete: link(base, 'DELETE') } : {},
  };
}

function journeyDto(journey: StubJourney): CapabilityJourney {
  const writer = canWriteJourneys();
  const active = isActiveStatus(journey.status);
  const writable = active && writer;
  const base = `/api/v1/capability-journeys/${journey.id}`;

  const links: CapabilityJourney['_links'] = {
    self: link(base, 'GET'),
    'x-history': link(`/api/v1/capabilities/${journey.capabilityId}/journey/history`, 'GET'),
  };
  if (writable) {
    if (journey.status === 'planned') links['x-start'] = link(`${base}/start`, 'POST');
    if (journey.status === 'in-flight') links['x-complete'] = link(`${base}/complete`, 'POST');
    links['x-abandon'] = link(`${base}/abandon`, 'POST');
    links.edit = link(`${base}/details`, 'PUT');
    links['x-progress'] = link(`${base}/progress`, 'PUT');
    links['x-change-sources'] = link(`${base}/source-applications`, 'PUT');
    links['x-add-milestone'] = link(`${base}/milestones`, 'POST');
  }

  return {
    id: journey.id,
    capabilityId: journey.capabilityId,
    capabilityName: journey.capabilityName,
    capabilityStale: journey.capabilityStale,
    kind: journey.kind,
    status: journey.status,
    progress: journey.progress,
    targetPeriod: journey.targetPeriod,
    note: journey.note,
    fromApplications: journey.fromApplications,
    toApplication: journey.toApplication,
    ...(journey.move ? { move: journey.move } : {}),
    milestones: journey.milestones.map((m) => milestoneDto(m, journey.id, writable)),
    plannedBy: journey.plannedBy,
    plannedByName: journey.plannedByName,
    plannedAt: journey.plannedAt,
    updatedAt: journey.updatedAt,
    startedAt: journey.startedAt,
    completedAt: journey.completedAt,
    abandonedAt: journey.abandonedAt,
    _links: links,
  };
}

function resolveApplicationRef(componentId: string): StubApplicationRef {
  const component = getComponent(toComponentId(componentId));
  return { componentId, componentName: component?.name ?? componentId, stale: false };
}

function buildCapturedJourney(capabilityId: string, body: CaptureJourneyRequest): StubJourney {
  const capability = getCapability(toCapabilityId(capabilityId));
  const now = new Date().toISOString();
  return {
    id: nextJourneyId(),
    capabilityId,
    capabilityName: capability?.name ?? 'Test Capability',
    capabilityStale: false,
    kind: body.kind,
    status: 'planned',
    progress: null,
    targetPeriod: body.targetPeriod ?? null,
    note: body.note ?? '',
    fromApplications: body.fromComponentIds.map(resolveApplicationRef),
    toApplication: resolveApplicationRef(body.toComponentId),
    move:
      body.kind === 'move'
        ? {
            targetDomainId: body.targetDomainId ?? '',
            targetDomainName: 'Test Domain',
            targetDomainStale: false,
            targetParentId: body.targetParentId ?? null,
            targetParentName: body.targetParentId ? 'Test Parent Capability' : '',
            targetParentStale: false,
            resultingName: body.resultingName ?? '',
          }
        : undefined,
    milestones: [],
    plannedBy: 'user-1',
    plannedByName: 'Test Architect',
    plannedAt: now,
    updatedAt: now,
    startedAt: null,
    completedAt: null,
    abandonedAt: null,
  };
}

function transitionHandler(action: 'start' | 'complete' | 'abandon', nextStatus: StubJourney['status']) {
  return http.post(`${BASE_URL}/api/v1/capability-journeys/:journeyId/${action}`, ({ params }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    journey.status = nextStatus;
    journey.updatedAt = new Date().toISOString();
    if (action === 'start') journey.startedAt = journey.updatedAt;
    if (action === 'complete') journey.completedAt = journey.updatedAt;
    if (action === 'abandon') journey.abandonedAt = journey.updatedAt;
    return HttpResponse.json(journeyDto(journey));
  });
}

export const spec182Handlers = [
  http.get(`${BASE_URL}/api/v1/capabilities/:id/journey`, ({ params }) => {
    const capabilityId = params.id as string;
    const active = findActiveJourneyForCapability(capabilityId);
    const links: { self: ReturnType<typeof link>; 'x-capture'?: ReturnType<typeof link> } = {
      self: link(`/api/v1/capabilities/${capabilityId}/journey`, 'GET'),
    };
    if (canWriteJourneys() && !active) {
      links['x-capture'] = link(`/api/v1/capabilities/${capabilityId}/journey`, 'POST');
    }
    return HttpResponse.json({ journey: active ? journeyDto(active) : null, _links: links });
  }),

  http.post(`${BASE_URL}/api/v1/capabilities/:id/journey`, async ({ params, request }) => {
    const capabilityId = params.id as string;
    const existingActive = findActiveJourneyForCapability(capabilityId);
    if (existingActive) {
      return HttpResponse.json(
        { message: `An active journey already exists for this capability (${existingActive.id})` },
        { status: 409 },
      );
    }
    const body = (await request.json()) as CaptureJourneyRequest;
    const journey = buildCapturedJourney(capabilityId, body);
    addJourney(journey);
    return HttpResponse.json(journeyDto(journey), { status: 201 });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/:id/journey/history`, ({ params }) => {
    const capabilityId = params.id as string;
    const data = getHistoryForCapability(capabilityId).map(journeyDto);
    return HttpResponse.json({
      data,
      _links: { self: link(`/api/v1/capabilities/${capabilityId}/journey/history`, 'GET') },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capability-journeys`, ({ request }) => {
    const url = new URL(request.url);
    const capabilityIds = splitIds(url.searchParams.get('capabilityIds'));
    const data = getCurrentByCapabilityIds(capabilityIds).map(journeyDto);
    const links: { self: ReturnType<typeof link>; 'x-capture'?: ReturnType<typeof link> } = {
      self: link(url.pathname + url.search, 'GET'),
    };
    if (canWriteJourneys()) {
      links['x-capture'] = link('/api/v1/capabilities/{id}/journey', 'POST');
    }
    return HttpResponse.json({ data, _links: links });
  }),

  transitionHandler('start', 'in-flight'),
  transitionHandler('complete', 'done'),
  transitionHandler('abandon', 'abandoned'),

  http.put(`${BASE_URL}/api/v1/capability-journeys/:journeyId/details`, async ({ params, request }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    const body = (await request.json()) as UpdateJourneyDetailsRequest;
    journey.note = body.note ?? '';
    journey.targetPeriod = body.targetPeriod ?? null;
    if (journey.move && body.resultingName !== undefined) {
      journey.move = { ...journey.move, resultingName: body.resultingName };
    }
    journey.updatedAt = new Date().toISOString();
    return HttpResponse.json(journeyDto(journey));
  }),

  http.put(`${BASE_URL}/api/v1/capability-journeys/:journeyId/progress`, async ({ params, request }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    const body = (await request.json()) as UpdateJourneyProgressRequest;
    journey.progress = body.progress;
    journey.updatedAt = new Date().toISOString();
    return HttpResponse.json(journeyDto(journey));
  }),

  http.put(`${BASE_URL}/api/v1/capability-journeys/:journeyId/source-applications`, async ({ params, request }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    const body = (await request.json()) as ChangeJourneySourceApplicationsRequest;
    journey.fromApplications = body.componentIds.map(resolveApplicationRef);
    journey.updatedAt = new Date().toISOString();
    return HttpResponse.json(journeyDto(journey));
  }),

  http.post(`${BASE_URL}/api/v1/capability-journeys/:journeyId/milestones`, async ({ params, request }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    const body = (await request.json()) as AddJourneyMilestoneRequest;
    journey.milestones.push({
      id: nextMilestoneId(),
      label: body.label,
      targetPeriod: body.targetPeriod ?? null,
      status: body.status ?? 'planned',
    });
    journey.updatedAt = new Date().toISOString();
    return HttpResponse.json(journeyDto(journey), { status: 201 });
  }),

  http.put(`${BASE_URL}/api/v1/capability-journeys/:journeyId/milestones/:milestoneId`, async ({ params, request }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    const index = journey.milestones.findIndex((m) => m.id === params.milestoneId);
    if (index === -1) return new HttpResponse(null, { status: 404 });
    const body = (await request.json()) as UpdateJourneyMilestoneRequest;
    journey.milestones[index] = {
      id: journey.milestones[index].id,
      label: body.label,
      targetPeriod: body.targetPeriod ?? null,
      status: body.status,
    };
    journey.updatedAt = new Date().toISOString();
    return HttpResponse.json(journeyDto(journey));
  }),

  http.delete(`${BASE_URL}/api/v1/capability-journeys/:journeyId/milestones/:milestoneId`, ({ params }) => {
    const journey = findJourney(params.journeyId as string);
    if (!journey) return new HttpResponse(null, { status: 404 });
    journey.milestones = journey.milestones.filter((m) => m.id !== params.milestoneId);
    journey.updatedAt = new Date().toISOString();
    return new HttpResponse(null, { status: 204 });
  }),
];
