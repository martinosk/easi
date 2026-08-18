import { describe, expect, it } from 'vitest';
import { buildCapabilityJourney } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import type { CapabilityJourney } from '../../journeys/types';
import { buildHierarchyJourneys } from './hierarchyJourneys';
import { buildJourneyIndex } from './journeyIndex';

const FERRY_BOOKING = cap('ferry-booking', 'Ferry booking', 'L1');
const HAZARDOUS = cap('hazardous', 'Hazardous', 'L2', 'ferry-booking');
const SPACE_CHARTER = cap('space-charter', 'Space charter', 'L2', 'ferry-booking');
const HAZCHECK = cap('hazcheck', 'Hazcheck routing', 'L3', 'hazardous');

const CAPABILITIES = [FERRY_BOOKING, HAZARDOUS, SPACE_CHARTER, HAZCHECK];

function derive(journeys: CapabilityJourney[], capabilityId: string) {
  const index = buildJourneyIndex({ journeys, capabilityDomainNames: new Map() });
  return buildHierarchyJourneys({ capabilityId, capabilities: CAPABILITIES, getJourney: index.getJourney });
}

function journeyOn(capabilityId: string, overrides: Partial<CapabilityJourney> = {}): CapabilityJourney {
  const capability = CAPABILITIES.find((c) => String(c.id) === capabilityId);
  return buildCapabilityJourney({
    id: `journey-${capabilityId}`,
    capabilityId,
    capabilityName: capability?.name ?? capabilityId,
    ...overrides,
  });
}

describe('buildHierarchyJourneys — descendants (rules 1-3)', () => {
  it('lists journeys of descendants at any depth', () => {
    const { descendants } = derive(
      [journeyOn('hazardous'), journeyOn('hazcheck'), journeyOn('ferry-booking')],
      'ferry-booking',
    );

    expect(descendants.map((journey) => journey.capabilityId)).toEqual(['hazardous', 'hazcheck']);
  });

  it('lists descendant journeys even when the capability has no journey of its own', () => {
    const { descendants } = derive([journeyOn('hazardous')], 'ferry-booking');

    expect(descendants.map((journey) => journey.capabilityId)).toEqual(['hazardous']);
  });

  it('includes a done descendant journey', () => {
    const { descendants } = derive([journeyOn('hazardous', { status: 'done', progress: 100 })], 'ferry-booking');

    expect(descendants.map((journey) => journey.status)).toEqual(['done']);
  });

  it('excludes a descendant whose current journey was abandoned', () => {
    const { descendants } = derive([journeyOn('hazardous', { status: 'abandoned' })], 'ferry-booking');

    expect(descendants).toEqual([]);
  });

  it('orders by target period ascending with undated journeys last', () => {
    const { descendants } = derive(
      [
        journeyOn('hazardous', { targetPeriod: null }),
        journeyOn('hazcheck', { targetPeriod: { year: 2028, quarter: 1 } }),
        journeyOn('space-charter', { targetPeriod: { year: 2026, quarter: 3 } }),
      ],
      'ferry-booking',
    );

    expect(descendants.map((journey) => journey.capabilityId)).toEqual(['space-charter', 'hazcheck', 'hazardous']);
  });

  it('breaks target period ties by capability name', () => {
    const period = { year: 2027, quarter: 2 };
    const { descendants } = derive(
      [journeyOn('space-charter', { targetPeriod: period }), journeyOn('hazardous', { targetPeriod: period })],
      'ferry-booking',
    );

    expect(descendants.map((journey) => journey.capabilityName)).toEqual(['Hazardous', 'Space charter']);
  });

  it('breaks ties by capability name when every journey is undated', () => {
    const { descendants } = derive(
      [journeyOn('space-charter', { targetPeriod: null }), journeyOn('hazardous', { targetPeriod: null })],
      'ferry-booking',
    );

    expect(descendants.map((journey) => journey.capabilityName)).toEqual(['Hazardous', 'Space charter']);
  });

  it('returns nothing for a capability without descendants', () => {
    const { descendants } = derive([journeyOn('ferry-booking')], 'hazcheck');

    expect(descendants).toEqual([]);
  });
});

describe('buildHierarchyJourneys — ancestors (rule 4)', () => {
  it('lists ancestors with an active journey, nearest first', () => {
    const { ancestors } = derive(
      [journeyOn('ferry-booking', { status: 'in-flight' }), journeyOn('hazardous', { status: 'planned' })],
      'hazcheck',
    );

    expect(ancestors.map((journey) => journey.capabilityId)).toEqual(['hazardous', 'ferry-booking']);
  });

  it('omits an ancestor whose journey is done', () => {
    const { ancestors } = derive([journeyOn('ferry-booking', { status: 'done' })], 'hazardous');

    expect(ancestors).toEqual([]);
  });

  it('omits an ancestor whose journey was abandoned', () => {
    const { ancestors } = derive([journeyOn('ferry-booking', { status: 'abandoned' })], 'hazardous');

    expect(ancestors).toEqual([]);
  });

  it('returns nothing for a capability without a parent', () => {
    const { ancestors } = derive([journeyOn('hazardous')], 'ferry-booking');

    expect(ancestors).toEqual([]);
  });
});

describe('buildHierarchyJourneys — unknown capability', () => {
  it('returns an empty composition', () => {
    const composition = derive([journeyOn('hazardous')], 'not-on-the-board');

    expect(composition).toEqual({ descendants: [], ancestors: [] });
  });
});
