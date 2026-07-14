import { HttpResponse, http } from 'msw';
import type { HttpMethod, TimeGrade } from '../../../api/types';
import type {
  TimeAssessment,
  TimeAssessmentGradeCounts,
  TimeAssessmentRollup,
} from '../../../features/architecture-direction/types';
import {
  assessmentsForCapabilities,
  assessmentsForComponents,
  canWriteAssessments,
  findAssessment,
  removeAssessment,
  type StubTimeAssessment,
  upsertAssessment,
} from './store';

const BASE_URL = '*';
const link = (href: string, method: HttpMethod) => ({ href, method });
const STALE_THRESHOLD_MS = 365 * 24 * 60 * 60 * 1000;

function isStale(assessedAt: string): boolean {
  return Date.now() - new Date(assessedAt).getTime() > STALE_THRESHOLD_MS;
}

function assessmentPath(capabilityId: string, componentId: string): string {
  return `/api/v1/capabilities/${capabilityId}/components/${componentId}/time-assessment`;
}

function toDto(a: StubTimeAssessment): TimeAssessment {
  const base = assessmentPath(a.capabilityId, a.componentId);
  return {
    id: a.id,
    capabilityId: a.capabilityId,
    capabilityName: a.capabilityName,
    componentId: a.componentId,
    componentName: a.componentName,
    grade: a.grade,
    rationale: a.rationale,
    assessedBy: a.assessedBy,
    assessedByName: a.assessedByName,
    assessedAt: a.assessedAt,
    stale: isStale(a.assessedAt),
    _links: {
      self: link(base, 'GET'),
      ...(canWriteAssessments() ? { edit: link(base, 'PUT'), delete: link(base, 'DELETE') } : {}),
    },
  };
}

function emptyCounts(): TimeAssessmentGradeCounts {
  return { Invest: 0, Tolerate: 0, Migrate: 0, Eliminate: 0 };
}

function splitIds(raw: string | null): string[] {
  return (raw ?? '').split(',').filter(Boolean);
}

export const spec180Handlers = [
  http.get(`${BASE_URL}/api/v1/time-assessments/rollups`, ({ request }) => {
    const url = new URL(request.url);
    const componentIds = splitIds(url.searchParams.get('componentIds'));
    const matches = assessmentsForComponents(componentIds);

    const data: TimeAssessmentRollup[] = componentIds.map((componentId) => {
      const counts = emptyCounts();
      for (const a of matches) {
        if (a.componentId === componentId) counts[a.grade]++;
      }
      return { componentId, counts };
    });

    return HttpResponse.json({ data, _links: { self: link(url.pathname + url.search, 'GET') } });
  }),

  http.get(`${BASE_URL}/api/v1/time-assessments`, ({ request }) => {
    const url = new URL(request.url);
    const capabilityIds = splitIds(url.searchParams.get('capabilityIds'));
    const data = assessmentsForCapabilities(capabilityIds).map(toDto);

    return HttpResponse.json({
      data,
      _links: {
        self: link(url.pathname + url.search, 'GET'),
        ...(canWriteAssessments()
          ? { 'x-assess': link('/api/v1/capabilities/{id}/components/{componentId}/time-assessment', 'PUT') }
          : {}),
      },
    });
  }),

  http.get(`${BASE_URL}/api/v1/capabilities/:id/components/:componentId/time-assessment`, ({ params }) => {
    const found = findAssessment(params.id as string, params.componentId as string);
    if (!found) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(toDto(found));
  }),

  http.put(
    `${BASE_URL}/api/v1/capabilities/:id/components/:componentId/time-assessment`,
    async ({ params, request }) => {
      const capabilityId = params.id as string;
      const componentId = params.componentId as string;
      const body = (await request.json()) as { grade: TimeGrade; rationale?: string };
      const existing = findAssessment(capabilityId, componentId);

      const record: StubTimeAssessment = {
        id: existing?.id ?? `ta-${capabilityId}-${componentId}`,
        capabilityId,
        capabilityName: existing?.capabilityName ?? 'Test Capability',
        componentId,
        componentName: existing?.componentName ?? 'Test Component',
        grade: body.grade,
        rationale: body.rationale ?? '',
        assessedBy: 'user-1',
        assessedByName: 'Test Architect',
        assessedAt: new Date().toISOString(),
      };
      upsertAssessment(record);

      return HttpResponse.json(toDto(record), { status: existing ? 200 : 201 });
    },
  ),

  http.delete(`${BASE_URL}/api/v1/capabilities/:id/components/:componentId/time-assessment`, ({ params }) => {
    removeAssessment(params.id as string, params.componentId as string);
    return new HttpResponse(null, { status: 204 });
  }),
];
