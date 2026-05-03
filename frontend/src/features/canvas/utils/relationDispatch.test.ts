import { describe, expect, it } from 'vitest';
import { planRelationCall, type RelationSubType } from './relationDispatch';

interface PlanCase {
  label: string;
  relationType: string;
  source: string;
  next: string;
  subType?: RelationSubType;
  expected: ReturnType<typeof planRelationCall>;
}

const cases: PlanCase[] = [
  {
    label: 'component-relation default subtype',
    relationType: 'component-relation',
    source: 'comp-a',
    next: 'comp-new',
    expected: {
      kind: 'component-relation',
      sourceComponentId: 'comp-a',
      targetComponentId: 'comp-new',
      relationSubType: 'Triggers',
    },
  },
  {
    label: 'component-relation explicit Serves',
    relationType: 'component-relation',
    source: 'comp-a',
    next: 'comp-new',
    subType: 'Serves',
    expected: {
      kind: 'component-relation',
      sourceComponentId: 'comp-a',
      targetComponentId: 'comp-new',
      relationSubType: 'Serves',
    },
  },
  {
    label: 'capability-parent puts the new capability as child',
    relationType: 'capability-parent',
    source: 'cap-source',
    next: 'cap-new',
    expected: { kind: 'capability-parent', childCapabilityId: 'cap-new', parentCapabilityId: 'cap-source' },
  },
  {
    label: 'capability-realization with source capability and new component',
    relationType: 'capability-realization',
    source: 'cap-source',
    next: 'comp-new',
    expected: { kind: 'capability-realization', capabilityId: 'cap-source', componentId: 'comp-new' },
  },
  {
    label: 'origin-acquired-via with new component and source entity',
    relationType: 'origin-acquired-via',
    source: 'acq-source',
    next: 'comp-new',
    expected: { kind: 'origin-acquired-via', componentId: 'comp-new', acquiredEntityId: 'acq-source' },
  },
  {
    label: 'origin-purchased-from with new component and source vendor',
    relationType: 'origin-purchased-from',
    source: 'vendor-1',
    next: 'comp-new',
    expected: { kind: 'origin-purchased-from', componentId: 'comp-new', vendorId: 'vendor-1' },
  },
  {
    label: 'origin-built-by with new component and source team',
    relationType: 'origin-built-by',
    source: 'team-1',
    next: 'comp-new',
    expected: { kind: 'origin-built-by', componentId: 'comp-new', internalTeamId: 'team-1' },
  },
];

describe('planRelationCall', () => {
  it.each(cases)('plans $label', ({ relationType, source, next, subType, expected }) => {
    expect(planRelationCall(relationType, source, next, subType)).toEqual(expected);
  });

  it('returns null for an unknown relationType', () => {
    expect(planRelationCall('totally-unknown', 'src', 'tgt')).toBeNull();
  });
});
