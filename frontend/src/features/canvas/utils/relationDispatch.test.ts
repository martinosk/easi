import { describe, expect, it } from 'vitest';
import type { RelatedTargetType } from '../../../utils/xRelated';
import { planRelationCall } from './relationDispatch';

interface PlanCase {
  label: string;
  relationType: string;
  source: string;
  next: string;
  targetType?: RelatedTargetType;
  expected: ReturnType<typeof planRelationCall>;
}

const cases: PlanCase[] = [
  {
    label: 'component-triggers maps to Triggers component-relation',
    relationType: 'component-triggers',
    source: 'comp-a',
    next: 'comp-new',
    targetType: 'component',
    expected: {
      kind: 'component-relation',
      sourceComponentId: 'comp-a',
      targetComponentId: 'comp-new',
      relationSubType: 'Triggers',
    },
  },
  {
    label: 'component-serves maps to Serves component-relation',
    relationType: 'component-serves',
    source: 'comp-a',
    next: 'comp-new',
    targetType: 'component',
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
    targetType: 'capability',
    expected: { kind: 'capability-parent', childCapabilityId: 'cap-new', parentCapabilityId: 'cap-source' },
  },
  {
    label: 'capability-realization with source capability and new component',
    relationType: 'capability-realization',
    source: 'cap-source',
    next: 'comp-new',
    targetType: 'component',
    expected: { kind: 'capability-realization', capabilityId: 'cap-source', componentId: 'comp-new' },
  },
  {
    label: 'origin-acquired-via from acquired-entity source places new component on the relation',
    relationType: 'origin-acquired-via',
    source: 'acq-source',
    next: 'comp-new',
    targetType: 'component',
    expected: { kind: 'origin-acquired-via', componentId: 'comp-new', acquiredEntityId: 'acq-source' },
  },
  {
    label: 'origin-acquired-via from component source places new acquired-entity on the relation',
    relationType: 'origin-acquired-via',
    source: 'comp-source',
    next: 'acq-new',
    targetType: 'acquiredEntity',
    expected: { kind: 'origin-acquired-via', componentId: 'comp-source', acquiredEntityId: 'acq-new' },
  },
  {
    label: 'origin-purchased-from from vendor source places new component on the relation',
    relationType: 'origin-purchased-from',
    source: 'vendor-1',
    next: 'comp-new',
    targetType: 'component',
    expected: { kind: 'origin-purchased-from', componentId: 'comp-new', vendorId: 'vendor-1' },
  },
  {
    label: 'origin-purchased-from from component source places new vendor on the relation',
    relationType: 'origin-purchased-from',
    source: 'comp-source',
    next: 'vendor-new',
    targetType: 'vendor',
    expected: { kind: 'origin-purchased-from', componentId: 'comp-source', vendorId: 'vendor-new' },
  },
  {
    label: 'origin-built-by from team source places new component on the relation',
    relationType: 'origin-built-by',
    source: 'team-1',
    next: 'comp-new',
    targetType: 'component',
    expected: { kind: 'origin-built-by', componentId: 'comp-new', internalTeamId: 'team-1' },
  },
  {
    label: 'origin-built-by from component source places new team on the relation',
    relationType: 'origin-built-by',
    source: 'comp-source',
    next: 'team-new',
    targetType: 'internalTeam',
    expected: { kind: 'origin-built-by', componentId: 'comp-source', internalTeamId: 'team-new' },
  },
];

describe('planRelationCall', () => {
  it.each(cases)('plans $label', ({ relationType, source, next, targetType, expected }) => {
    expect(planRelationCall(relationType, source, next, targetType)).toEqual(expected);
  });

  it('returns null for an unknown relationType', () => {
    expect(planRelationCall('totally-unknown', 'src', 'tgt')).toBeNull();
  });

  it('returns null for the legacy component-relation type which is no longer in the contract', () => {
    expect(planRelationCall('component-relation', 'src', 'tgt')).toBeNull();
  });
});
