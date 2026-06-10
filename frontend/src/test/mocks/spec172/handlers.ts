import { HttpResponse, http } from 'msw';
import type { HttpMethod } from '../../../api/types';
import type {
  AddSourceRequest,
  CaptureDirectionRequest,
  CompositionPreviewRequest,
  CompositionPreviewResponse,
  SourceCandidate,
  SourceCandidatesResponse,
} from '../../../features/architecture-direction/types';
import { evaluateEligibility, resolveComposition } from './composition';
import {
  type StubDirection,
  addDirection,
  buildCompositionResponse,
  buildDirectionDto,
  buildEcDirectionResponse,
  buildEnterpriseCapabilityDto,
  getActiveDirection,
  getStubCapabilities,
  getStubEnterpriseCapability,
  otherActiveDirections,
  upsertDirection,
} from './store';

const BASE_URL = 'http://localhost:8080';

const link = (href: string, method: HttpMethod) => ({ href, method });

function conflictResponse(ecName: string, capabilityName: string, capabilityId: string, conflictEcId: string) {
  return HttpResponse.json(
    {
      error: 'Conflict',
      message: `Capability '${capabilityName}' is already an explicit source of an active direction on '${ecName}'. A domain capability may be the explicit source of at most one active direction.`,
      details: {
        capabilityId,
        capabilityName,
        conflictingEnterpriseCapabilityId: conflictEcId,
        conflictingEnterpriseCapabilityName: ecName,
      },
      _links: { 'x-conflicting-ec': link(`/api/v1/enterprise-capabilities/${conflictEcId}`, 'GET') },
    },
    { status: 409 },
  );
}

function immutableResponse(ecId: string) {
  return HttpResponse.json(
    {
      error: 'Conflict',
      message: 'This direction is agreed and its source set is frozen. To recompose, reject the direction and capture a new one.',
      details: { directionStatus: 'agreed' },
      _links: { 'x-reject': link(`/api/v1/enterprise-capabilities/${ecId}/direction/reject`, 'POST') },
    },
    { status: 409 },
  );
}

function newDirectionId(): string {
  return `dir-${Math.random().toString(36).slice(2, 10)}`;
}

export const spec172Handlers = [
  http.get(`${BASE_URL}/api/v1/enterprise-capabilities/:id`, ({ params }) => {
    const ec = getStubEnterpriseCapability(params.id as string);
    if (!ec) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(buildEnterpriseCapabilityDto(ec));
  }),

  http.get(`${BASE_URL}/api/v1/enterprise-capabilities/:id/composition`, ({ params }) => {
    const ec = getStubEnterpriseCapability(params.id as string);
    if (!ec) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(buildCompositionResponse(params.id as string));
  }),

  http.get(`${BASE_URL}/api/v1/enterprise-capabilities/:id/standard-application`, ({ params }) => {
    const base = `/api/v1/enterprise-capabilities/${params.id}/standard-application`;
    return HttpResponse.json({
      standard: null,
      _links: { self: link(base, 'GET'), up: link(`/api/v1/enterprise-capabilities/${params.id}`, 'GET') },
    });
  }),

  http.get(`${BASE_URL}/api/v1/enterprise-capabilities/:id/direction`, ({ params }) => {
    const ec = getStubEnterpriseCapability(params.id as string);
    if (!ec) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(buildEcDirectionResponse(params.id as string));
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/source-candidates`, ({ request }) => {
    const url = new URL(request.url);
    const q = (url.searchParams.get('q') ?? '').trim();
    const ecId = url.searchParams.get('ecId') ?? '';
    const domainId = url.searchParams.get('domainId') ?? undefined;
    const limit = Number(url.searchParams.get('limit') ?? '20');
    if (!q || !ecId) {
      return HttpResponse.json({ error: 'BadRequest', message: 'q and ecId are required' }, { status: 400 });
    }

    const others = otherActiveDirections(ecId);
    const matches = getStubCapabilities()
      .filter((c) => c.name.toLowerCase().includes(q.toLowerCase()))
      .filter((c) => !domainId || c.businessDomainId === domainId);
    const limited = matches.slice(0, limit);

    const data: SourceCandidate[] = limited.map((c) => {
      const eligibility = evaluateEligibility(c.id, ecId, others);
      const links: SourceCandidate['_links'] = { self: link(`/api/v1/capabilities/${c.id}`, 'GET') };
      if (!eligibility.eligible && eligibility.conflictingEnterpriseCapability) {
        links['x-conflicting-ec'] = link(
          `/api/v1/enterprise-capabilities/${eligibility.conflictingEnterpriseCapability.id}`,
          'GET',
        );
      }
      return {
        capabilityId: c.id,
        name: c.name,
        level: c.level,
        parentId: c.parentId ?? null,
        businessDomainId: c.businessDomainId ?? null,
        businessDomainName: c.businessDomainName ?? null,
        eligible: eligibility.eligible,
        ineligibilityReason: eligibility.ineligibilityReason,
        conflictingEnterpriseCapability: eligibility.conflictingEnterpriseCapability,
        _links: links,
      };
    });

    const response: SourceCandidatesResponse = {
      data,
      pagination: { hasMore: matches.length > limited.length, limit, cursor: '' },
      _links: { self: link(url.pathname + url.search, 'GET') },
    };
    return HttpResponse.json(response);
  }),

  http.post(`${BASE_URL}/api/v1/enterprise-capabilities/:id/direction/composition-preview`, async ({ params, request }) => {
    const ecId = params.id as string;
    const ec = getStubEnterpriseCapability(ecId);
    if (!ec) return new HttpResponse(null, { status: 404 });
    const body = (await request.json()) as CompositionPreviewRequest;
    const sourceIds = body.sourceCapabilityIds ?? [];

    const others = otherActiveDirections(ecId);
    const synthetic = [{ ecId, ecName: ec.name, sourceCapabilityIds: sourceIds }, ...others];
    const included = resolveComposition(ecId, synthetic, getStubCapabilities());

    const response: CompositionPreviewResponse = {
      includedCapabilities: included.map((i) => ({
        capabilityId: i.capabilityId,
        name: i.name,
        level: i.level,
        businessDomainId: i.businessDomainId,
        businessDomainName: i.businessDomainName,
        role: i.role,
        carvedOutBy: i.carvedOutBy,
      })),
      sourceEligibility: sourceIds.map((capId) => {
        const e = evaluateEligibility(capId, ecId, others);
        return {
          capabilityId: capId,
          eligible: e.eligible,
          ineligibilityReason: e.ineligibilityReason,
          conflictingEnterpriseCapability: e.conflictingEnterpriseCapability,
        };
      }),
      meta: {
        sourceCount: sourceIds.length,
        includedCount: included.filter((i) => i.role !== 'carved-out').length,
        carvedOutCount: included.filter((i) => i.role === 'carved-out').length,
      },
      _links: { self: link(`/api/v1/enterprise-capabilities/${ecId}/direction/composition-preview`, 'POST') },
    };
    return HttpResponse.json(response);
  }),

  http.post(`${BASE_URL}/api/v1/enterprise-capabilities/:id/direction`, async ({ params, request }) => {
    const ecId = params.id as string;
    const ec = getStubEnterpriseCapability(ecId);
    if (!ec) return new HttpResponse(null, { status: 404 });
    if (!ec.active) {
      return HttpResponse.json(
        { error: 'Conflict', message: 'Directions can only be captured on active enterprise capabilities.' },
        { status: 409 },
      );
    }
    const body = (await request.json()) as CaptureDirectionRequest;
    const others = otherActiveDirections(ecId);
    for (const capId of body.sourceCapabilityIds) {
      const e = evaluateEligibility(capId, ecId, others);
      if (!e.eligible && e.conflictingEnterpriseCapability) {
        const cap = getStubCapabilities().find((c) => c.id === capId);
        return conflictResponse(e.conflictingEnterpriseCapability.name, cap?.name ?? capId, capId, e.conflictingEnterpriseCapability.id);
      }
    }
    const direction: StubDirection = {
      id: newDirectionId(),
      enterpriseCapabilityId: ecId,
      type: body.type,
      status: 'draft',
      horizon: body.horizon,
      narrative: body.narrative,
      sourceCapabilityIds: [...body.sourceCapabilityIds],
      createdAt: new Date().toISOString(),
    };
    addDirection(direction);
    return HttpResponse.json(buildDirectionDto(direction), { status: 201 });
  }),

  http.post(`${BASE_URL}/api/v1/enterprise-capabilities/:id/direction/sources`, async ({ params, request }) => {
    const ecId = params.id as string;
    const direction = getActiveDirection(ecId);
    if (!direction) return new HttpResponse(null, { status: 404 });
    if (direction.status === 'agreed') return immutableResponse(ecId);

    const body = (await request.json()) as AddSourceRequest;
    const others = otherActiveDirections(ecId);
    const e = evaluateEligibility(body.capabilityId, ecId, others);
    if (!e.eligible && e.conflictingEnterpriseCapability) {
      const cap = getStubCapabilities().find((c) => c.id === body.capabilityId);
      return conflictResponse(e.conflictingEnterpriseCapability.name, cap?.name ?? body.capabilityId, body.capabilityId, e.conflictingEnterpriseCapability.id);
    }
    if (!direction.sourceCapabilityIds.includes(body.capabilityId)) {
      direction.sourceCapabilityIds.push(body.capabilityId);
      direction.updatedAt = new Date().toISOString();
      upsertDirection(direction);
    }
    return HttpResponse.json(buildDirectionDto(direction));
  }),

  http.delete(`${BASE_URL}/api/v1/enterprise-capabilities/:id/direction/sources/:capabilityId`, ({ params }) => {
    const ecId = params.id as string;
    const capabilityId = params.capabilityId as string;
    const direction = getActiveDirection(ecId);
    if (!direction) return new HttpResponse(null, { status: 404 });
    if (direction.status === 'agreed') return immutableResponse(ecId);
    if (!direction.sourceCapabilityIds.includes(capabilityId)) return new HttpResponse(null, { status: 404 });

    direction.sourceCapabilityIds = direction.sourceCapabilityIds.filter((id) => id !== capabilityId);
    direction.updatedAt = new Date().toISOString();
    upsertDirection(direction);
    return new HttpResponse(null, { status: 204 });
  }),

  http.post(`${BASE_URL}/api/v1/enterprise-capabilities/:id/direction/:action`, ({ params }) => {
    const ecId = params.id as string;
    const action = params.action as string;
    const direction = getActiveDirection(ecId);
    if (!direction) return new HttpResponse(null, { status: 404 });
    const transitions: Record<string, StubDirection['status']> = {
      propose: 'proposed',
      agree: 'agreed',
      reject: 'rejected',
    };
    const next = transitions[action];
    if (!next) return new HttpResponse(null, { status: 404 });
    direction.status = next;
    direction.updatedAt = new Date().toISOString();
    upsertDirection(direction);
    return HttpResponse.json(buildDirectionDto(direction));
  }),
];
