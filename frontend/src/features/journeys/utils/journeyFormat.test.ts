import { describe, expect, it } from 'vitest';
import { buildJourneyMilestone } from '../../../test/helpers/entityBuilders';
import type { CapabilityJourney, JourneyMilestone } from '../types';
import {
  formatTargetPeriod,
  journeyKindLabel,
  journeyStatusLabel,
  milestoneWhenLabel,
  scheduleConflictLabel,
} from './journeyFormat';

function journey(overrides: Partial<CapabilityJourney> = {}): CapabilityJourney {
  return {
    id: 'journey-1',
    capabilityId: 'cap-1',
    capabilityName: 'Booking management',
    capabilityStale: false,
    kind: 'migration',
    status: 'planned',
    progress: null,
    targetPeriod: null,
    note: '',
    fromApplications: [],
    toApplication: { componentId: 'comp-1', componentName: 'Phoenix', stale: false },
    milestones: [],
    plannedBy: 'user-1',
    plannedByName: 'Architect',
    plannedAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    startedAt: null,
    completedAt: null,
    abandonedAt: null,
    _links: {},
    ...overrides,
  };
}

function milestone(overrides: Partial<JourneyMilestone> = {}): JourneyMilestone {
  return { id: 'ms-1', label: 'API live', targetPeriod: null, status: 'planned', _links: {}, ...overrides };
}

describe('journeyStatusLabel — mockup-literal wording', () => {
  it('renders planned as "planned"', () => {
    expect(journeyStatusLabel(journey({ status: 'planned' }))).toBe('planned');
  });

  it('renders in-flight with progress as "in flight (60%)"', () => {
    expect(journeyStatusLabel(journey({ status: 'in-flight', progress: 60 }))).toBe('in flight (60%)');
  });

  it('renders in-flight without progress as "in flight"', () => {
    expect(journeyStatusLabel(journey({ status: 'in-flight', progress: null }))).toBe('in flight');
  });

  it('renders done as "done (100%)"', () => {
    expect(journeyStatusLabel(journey({ status: 'done', progress: 60 }))).toBe('done (100%)');
  });

  it('renders abandoned as "abandoned"', () => {
    expect(journeyStatusLabel(journey({ status: 'abandoned' }))).toBe('abandoned');
  });
});

describe('journeyKindLabel', () => {
  it('renders a move as "capability move"', () => {
    expect(journeyKindLabel('move')).toBe('capability move');
  });

  it('renders other kinds verbatim', () => {
    expect(journeyKindLabel('migration')).toBe('migration');
    expect(journeyKindLabel('consolidation')).toBe('consolidation');
    expect(journeyKindLabel('carve-out')).toBe('carve-out');
  });
});

describe('formatTargetPeriod', () => {
  it('renders a period as "Q2 2027"', () => {
    expect(formatTargetPeriod({ year: 2027, quarter: 2 })).toBe('Q2 2027');
  });

  it('renders a missing period as an em dash', () => {
    expect(formatTargetPeriod(null)).toBe('—');
  });
});

describe('milestoneWhenLabel — mockup-literal wording', () => {
  it('renders a done milestone with a period as "Done · Q4 2025"', () => {
    expect(milestoneWhenLabel(milestone({ status: 'done', targetPeriod: { year: 2025, quarter: 4 } }))).toBe(
      'Done · Q4 2025',
    );
  });

  it('renders a done milestone without a period as "Done"', () => {
    expect(milestoneWhenLabel(milestone({ status: 'done' }))).toBe('Done');
  });

  it('renders a pending milestone with a period as "Q4 2026"', () => {
    expect(milestoneWhenLabel(milestone({ status: 'in-flight', targetPeriod: { year: 2026, quarter: 4 } }))).toBe(
      'Q4 2026',
    );
  });

  it('renders a pending milestone without a period as empty', () => {
    expect(milestoneWhenLabel(milestone({ status: 'planned' }))).toBe('');
  });
});

describe('scheduleConflictLabel', () => {
  it('names both periods in reading order', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: { year: 2026, quarter: 4 } });

    expect(scheduleConflictLabel(milestone, { year: 2027, quarter: 1 })).toBe(
      'Targeted for Q4 2026 but listed after a milestone targeted for Q1 2027',
    );
  });
});
