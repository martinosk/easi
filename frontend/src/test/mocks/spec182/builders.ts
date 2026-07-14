import type { StubJourney } from './store';

export function buildStubJourney(overrides: Partial<StubJourney> = {}): StubJourney {
  return {
    id: 'journey-1',
    capabilityId: 'cap-1',
    capabilityName: 'Booking management',
    capabilityStale: false,
    kind: 'migration',
    status: 'planned',
    progress: null,
    targetPeriod: { year: 2027, quarter: 2 },
    note: 'Route-by-route migration. 14 of 23 routes live on Phoenix; North Sea corridor next.',
    fromApplications: [{ componentId: 'comp-seabook', componentName: 'Seabook', stale: false }],
    toApplication: { componentId: 'comp-phoenix', componentName: 'Phoenix', stale: false },
    milestones: [],
    plannedBy: 'user-1',
    plannedByName: 'Domain Architect',
    plannedAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    startedAt: null,
    completedAt: null,
    abandonedAt: null,
    ...overrides,
  };
}
