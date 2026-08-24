import { describe, expect, it } from 'vitest';
import { buildCapabilityJourney } from '../../../test/helpers';
import { buildJourneyMilestone } from '../../../test/helpers/entityBuilders';
import type { TargetPeriod } from '../../journeys/types';
import { buildTimeline, isJourneyOverdue, isMilestoneOverdue } from './timelineModel';

const CURRENT: TargetPeriod = { year: 2026, quarter: 3 };

function domainNames(entries: Record<string, string>) {
  return (capabilityId: string) => entries[capabilityId];
}

describe('isMilestoneOverdue', () => {
  it('flags a milestone with a past target period that is not done', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: { year: 2026, quarter: 1 }, status: 'in-flight' });
    expect(isMilestoneOverdue(milestone, CURRENT)).toBe(true);
  });

  it('flags a planned milestone with a past target period', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: { year: 2025, quarter: 4 }, status: 'planned' });
    expect(isMilestoneOverdue(milestone, CURRENT)).toBe(true);
  });

  it('does not flag a done milestone with a past target period', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: { year: 2026, quarter: 1 }, status: 'done' });
    expect(isMilestoneOverdue(milestone, CURRENT)).toBe(false);
  });

  it('does not flag a milestone targeting the current quarter', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: CURRENT, status: 'planned' });
    expect(isMilestoneOverdue(milestone, CURRENT)).toBe(false);
  });

  it('does not flag a milestone targeting a future quarter', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: { year: 2027, quarter: 1 }, status: 'planned' });
    expect(isMilestoneOverdue(milestone, CURRENT)).toBe(false);
  });

  it('never flags an undated milestone', () => {
    const milestone = buildJourneyMilestone({ targetPeriod: null, status: 'planned' });
    expect(isMilestoneOverdue(milestone, CURRENT)).toBe(false);
  });
});

describe('isJourneyOverdue', () => {
  it('flags an in-flight journey with a past target period', () => {
    const journey = buildCapabilityJourney({ status: 'in-flight', targetPeriod: { year: 2026, quarter: 2 } });
    expect(isJourneyOverdue(journey, CURRENT)).toBe(true);
  });

  it('flags a planned journey with a past target period', () => {
    const journey = buildCapabilityJourney({ status: 'planned', targetPeriod: { year: 2026, quarter: 1 } });
    expect(isJourneyOverdue(journey, CURRENT)).toBe(true);
  });

  it('does not flag a done journey with a past target period', () => {
    const journey = buildCapabilityJourney({ status: 'done', targetPeriod: { year: 2026, quarter: 1 } });
    expect(isJourneyOverdue(journey, CURRENT)).toBe(false);
  });

  it('does not flag a journey targeting the current quarter', () => {
    const journey = buildCapabilityJourney({ status: 'in-flight', targetPeriod: CURRENT });
    expect(isJourneyOverdue(journey, CURRENT)).toBe(false);
  });

  it('never flags an undated journey', () => {
    const journey = buildCapabilityJourney({ status: 'planned', targetPeriod: null });
    expect(isJourneyOverdue(journey, CURRENT)).toBe(false);
  });
});

describe('buildTimeline content', () => {
  it('includes only active journeys', () => {
    const model = buildTimeline({
      journeys: [
        buildCapabilityJourney({ capabilityId: 'cap-a', capabilityName: 'A', status: 'planned' }),
        buildCapabilityJourney({ capabilityId: 'cap-b', capabilityName: 'B', status: 'in-flight' }),
        buildCapabilityJourney({ capabilityId: 'cap-c', capabilityName: 'C', status: 'done' }),
        buildCapabilityJourney({ capabilityId: 'cap-d', capabilityName: 'D', status: 'abandoned' }),
      ],
      domainNameFor: domainNames({ 'cap-a': 'Ferry', 'cap-b': 'Ferry', 'cap-c': 'Ferry', 'cap-d': 'Ferry' }),
      current: CURRENT,
    });

    const names = model.groups.flatMap((group) => group.journeys.map((row) => row.journey.capabilityName));
    expect(names).toEqual(['A', 'B']);
  });

  it('omits journeys whose capability is not on the board', () => {
    const model = buildTimeline({
      journeys: [buildCapabilityJourney({ capabilityId: 'cap-unassigned', status: 'planned' })],
      domainNameFor: domainNames({}),
      current: CURRENT,
    });

    expect(model.groups).toEqual([]);
  });
});

describe('buildTimeline ordering', () => {
  it('orders domains alphabetically', () => {
    const model = buildTimeline({
      journeys: [
        buildCapabilityJourney({ capabilityId: 'cap-z', status: 'planned' }),
        buildCapabilityJourney({ capabilityId: 'cap-a', status: 'planned' }),
      ],
      domainNameFor: domainNames({ 'cap-z': 'Zebra', 'cap-a': 'Alpha' }),
      current: CURRENT,
    });

    expect(model.groups.map((group) => group.domainName)).toEqual(['Alpha', 'Zebra']);
  });

  it('orders journeys by target period ascending, undated last, ties by capability name', () => {
    const model = buildTimeline({
      journeys: [
        buildCapabilityJourney({ capabilityId: 'c1', capabilityName: 'Late', targetPeriod: { year: 2027, quarter: 1 } }),
        buildCapabilityJourney({ capabilityId: 'c2', capabilityName: 'Undated', targetPeriod: null }),
        buildCapabilityJourney({ capabilityId: 'c3', capabilityName: 'Beta', targetPeriod: { year: 2026, quarter: 4 } }),
        buildCapabilityJourney({ capabilityId: 'c4', capabilityName: 'Alpha', targetPeriod: { year: 2026, quarter: 4 } }),
      ],
      domainNameFor: domainNames({ c1: 'Ferry', c2: 'Ferry', c3: 'Ferry', c4: 'Ferry' }),
      current: CURRENT,
    });

    expect(model.groups[0].journeys.map((row) => row.journey.capabilityName)).toEqual([
      'Alpha',
      'Beta',
      'Late',
      'Undated',
    ]);
  });

  it('keeps milestone rows in stored order, dated before undated', () => {
    const journey = buildCapabilityJourney({
      capabilityId: 'cap-a',
      milestones: [
        buildJourneyMilestone({ id: 'm1', label: 'First stored', targetPeriod: { year: 2026, quarter: 4 } }),
        buildJourneyMilestone({ id: 'm2', label: 'No date', targetPeriod: null }),
        buildJourneyMilestone({ id: 'm3', label: 'Second stored', targetPeriod: { year: 2026, quarter: 4 } }),
        buildJourneyMilestone({ id: 'm4', label: 'Earlier stored later', targetPeriod: { year: 2026, quarter: 1 } }),
      ],
    });

    const model = buildTimeline({
      journeys: [journey],
      domainNameFor: domainNames({ 'cap-a': 'Ferry' }),
      current: CURRENT,
    });

    expect(model.groups[0].journeys[0].milestones.map((row) => row.milestone.id)).toEqual(['m1', 'm3', 'm4', 'm2']);
  });
});

describe('buildTimeline axis', () => {
  it('spans the earliest to the latest period across journeys and milestones', () => {
    const model = buildTimeline({
      journeys: [
        buildCapabilityJourney({
          capabilityId: 'cap-a',
          targetPeriod: { year: 2028, quarter: 4 },
          milestones: [buildJourneyMilestone({ targetPeriod: { year: 2025, quarter: 4 } })],
        }),
      ],
      domainNameFor: domainNames({ 'cap-a': 'Ferry' }),
      current: CURRENT,
    });

    expect(model.quarters[0]).toEqual({ year: 2025, quarter: 4 });
    expect(model.quarters[model.quarters.length - 1]).toEqual({ year: 2028, quarter: 4 });
    expect(model.quarters).toHaveLength(13);
  });

  it('always includes and marks the current quarter', () => {
    const model = buildTimeline({
      journeys: [buildCapabilityJourney({ capabilityId: 'cap-a', targetPeriod: { year: 2027, quarter: 2 } })],
      domainNameFor: domainNames({ 'cap-a': 'Ferry' }),
      current: CURRENT,
    });

    expect(model.quarters[model.currentColumn]).toEqual(CURRENT);
    expect(model.quarters[0]).toEqual(CURRENT);
  });

  it('extends backwards when every period is before the current quarter', () => {
    const model = buildTimeline({
      journeys: [buildCapabilityJourney({ capabilityId: 'cap-a', targetPeriod: { year: 2026, quarter: 1 } })],
      domainNameFor: domainNames({ 'cap-a': 'Ferry' }),
      current: CURRENT,
    });

    expect(model.quarters[0]).toEqual({ year: 2026, quarter: 1 });
    expect(model.quarters[model.quarters.length - 1]).toEqual(CURRENT);
  });

  it('renders a single-quarter axis when there are no journeys', () => {
    const model = buildTimeline({ journeys: [], domainNameFor: domainNames({}), current: CURRENT });

    expect(model.groups).toEqual([]);
    expect(model.quarters).toEqual([CURRENT]);
    expect(model.currentColumn).toBe(0);
  });

  it('places journey targets and milestones at their quarter columns, undated at null', () => {
    const model = buildTimeline({
      journeys: [
        buildCapabilityJourney({
          capabilityId: 'cap-a',
          targetPeriod: { year: 2027, quarter: 1 },
          milestones: [
            buildJourneyMilestone({ id: 'm1', targetPeriod: { year: 2026, quarter: 4 } }),
            buildJourneyMilestone({ id: 'm2', targetPeriod: null }),
          ],
        }),
      ],
      domainNameFor: domainNames({ 'cap-a': 'Ferry' }),
      current: CURRENT,
    });

    const row = model.groups[0].journeys[0];
    expect(row.targetColumn).toBe(2);
    expect(row.milestones[0].column).toBe(1);
    expect(row.milestones[1].column).toBeNull();
  });
});

describe('buildTimeline summary', () => {
  it('counts planned, in-flight, overdue journeys and overdue milestones', () => {
    const model = buildTimeline({
      journeys: [
        buildCapabilityJourney({
          capabilityId: 'c1',
          status: 'in-flight',
          targetPeriod: { year: 2026, quarter: 1 },
          milestones: [
            buildJourneyMilestone({ targetPeriod: { year: 2026, quarter: 1 }, status: 'planned' }),
            buildJourneyMilestone({ targetPeriod: { year: 2026, quarter: 2 }, status: 'in-flight' }),
            buildJourneyMilestone({ targetPeriod: { year: 2026, quarter: 1 }, status: 'done' }),
          ],
        }),
        buildCapabilityJourney({
          capabilityId: 'c2',
          status: 'planned',
          targetPeriod: { year: 2026, quarter: 1 },
          milestones: [buildJourneyMilestone({ targetPeriod: { year: 2025, quarter: 4 }, status: 'planned' })],
        }),
        buildCapabilityJourney({ capabilityId: 'c3', status: 'in-flight', targetPeriod: { year: 2027, quarter: 1 } }),
      ],
      domainNameFor: domainNames({ c1: 'Ferry', c2: 'Ferry', c3: 'Group' }),
      current: CURRENT,
    });

    expect(model.summary).toEqual({ planned: 1, inFlight: 2, overdueJourneys: 2, overdueMilestones: 3 });
  });
});
