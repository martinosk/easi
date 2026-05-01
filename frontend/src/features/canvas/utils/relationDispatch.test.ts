import { describe, expect, it } from 'vitest';
import { planRelationCall } from './relationDispatch';

describe('planRelationCall', () => {
  it('plans component-to-component as source -> new target', () => {
    expect(planRelationCall('component-relation', 'comp-a', 'comp-new')).toEqual({
      kind: 'component-relation',
      sourceComponentId: 'comp-a',
      targetComponentId: 'comp-new',
    });
  });

  it('plans capability-parent so the new capability is the child', () => {
    expect(planRelationCall('capability-parent', 'cap-source', 'cap-new')).toEqual({
      kind: 'capability-parent',
      childCapabilityId: 'cap-new',
      parentCapabilityId: 'cap-source',
    });
  });

  it('plans capability-realization with the source capability and new component', () => {
    expect(planRelationCall('capability-realization', 'cap-source', 'comp-new')).toEqual({
      kind: 'capability-realization',
      capabilityId: 'cap-source',
      componentId: 'comp-new',
    });
  });

  it('plans origin-acquired-via with the new component and source entity', () => {
    expect(planRelationCall('origin-acquired-via', 'acq-source', 'comp-new')).toEqual({
      kind: 'origin-acquired-via',
      componentId: 'comp-new',
      acquiredEntityId: 'acq-source',
    });
  });

  it('plans origin-purchased-from with the new component and source vendor', () => {
    expect(planRelationCall('origin-purchased-from', 'vendor-1', 'comp-new')).toEqual({
      kind: 'origin-purchased-from',
      componentId: 'comp-new',
      vendorId: 'vendor-1',
    });
  });

  it('plans origin-built-by with the new component and source team', () => {
    expect(planRelationCall('origin-built-by', 'team-1', 'comp-new')).toEqual({
      kind: 'origin-built-by',
      componentId: 'comp-new',
      internalTeamId: 'team-1',
    });
  });

  it('returns null for an unknown relationType', () => {
    expect(planRelationCall('totally-unknown', 'src', 'tgt')).toBeNull();
  });
});
