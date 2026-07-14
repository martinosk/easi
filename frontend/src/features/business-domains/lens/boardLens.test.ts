import { describe, expect, it } from 'vitest';
import { buildCapabilityJourney } from '../../../test/helpers';
import { capabilityJourneyStatus, isBoardLens, selectBoardJourney } from './boardLens';

describe('isBoardLens', () => {
  it('accepts the three lenses and rejects anything else', () => {
    expect(isBoardLens('now')).toBe(true);
    expect(isBoardLens('journey')).toBe(true);
    expect(isBoardLens('target')).toBe(true);
    expect(isBoardLens('elsewhere')).toBe(false);
    expect(isBoardLens(null)).toBe(false);
  });
});

describe('selectBoardJourney', () => {
  it('prefers an active journey over a done one', () => {
    const active = buildCapabilityJourney({ id: 'active', status: 'in-flight' });
    const done = buildCapabilityJourney({ id: 'done', status: 'done' });
    expect(selectBoardJourney([done, active])?.id).toBe('active');
  });

  it('prefers a planned journey (also active) when present', () => {
    const planned = buildCapabilityJourney({ id: 'planned', status: 'planned' });
    const done = buildCapabilityJourney({ id: 'done', status: 'done' });
    expect(selectBoardJourney([done, planned])?.id).toBe('planned');
  });

  it('falls back to a done journey when there is no active one', () => {
    const done = buildCapabilityJourney({ id: 'done', status: 'done' });
    expect(selectBoardJourney([done])?.id).toBe('done');
  });

  it('renders nothing for an abandoned-only capability', () => {
    const abandoned = buildCapabilityJourney({ id: 'abandoned', status: 'abandoned' });
    expect(selectBoardJourney([abandoned])).toBeUndefined();
  });

  it('returns undefined when there are no journeys', () => {
    expect(selectBoardJourney([])).toBeUndefined();
  });
});

describe('capabilityJourneyStatus', () => {
  it('is steady with no journey', () => {
    expect(capabilityJourneyStatus(undefined)).toBe('steady');
  });

  it('maps planned to not-started, in-flight to in-flight, done to done', () => {
    expect(capabilityJourneyStatus(buildCapabilityJourney({ kind: 'migration', status: 'planned' }))).toBe(
      'not-started',
    );
    expect(capabilityJourneyStatus(buildCapabilityJourney({ kind: 'migration', status: 'in-flight' }))).toBe(
      'in-flight',
    );
    expect(capabilityJourneyStatus(buildCapabilityJourney({ kind: 'migration', status: 'done' }))).toBe('done');
  });

  it('maps a planned move to planned-move', () => {
    expect(capabilityJourneyStatus(buildCapabilityJourney({ kind: 'move', status: 'planned' }))).toBe('planned-move');
  });
});
