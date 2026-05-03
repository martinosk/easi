export type RelationSubType = 'Triggers' | 'Serves';

export type RelationCallSpec =
  | {
      kind: 'component-relation';
      sourceComponentId: string;
      targetComponentId: string;
      relationSubType: RelationSubType;
    }
  | { kind: 'capability-parent'; childCapabilityId: string; parentCapabilityId: string }
  | { kind: 'capability-realization'; capabilityId: string; componentId: string }
  | { kind: 'origin-acquired-via'; componentId: string; acquiredEntityId: string }
  | { kind: 'origin-purchased-from'; componentId: string; vendorId: string }
  | { kind: 'origin-built-by'; componentId: string; internalTeamId: string };

export function planRelationCall(
  relationType: string,
  sourceEntityId: string,
  newEntityId: string,
  relationSubType: RelationSubType = 'Triggers',
): RelationCallSpec | null {
  switch (relationType) {
    case 'component-relation':
      return {
        kind: 'component-relation',
        sourceComponentId: sourceEntityId,
        targetComponentId: newEntityId,
        relationSubType,
      };
    case 'capability-parent':
      return { kind: 'capability-parent', childCapabilityId: newEntityId, parentCapabilityId: sourceEntityId };
    case 'capability-realization':
      return { kind: 'capability-realization', capabilityId: sourceEntityId, componentId: newEntityId };
    case 'origin-acquired-via':
      return { kind: 'origin-acquired-via', componentId: newEntityId, acquiredEntityId: sourceEntityId };
    case 'origin-purchased-from':
      return { kind: 'origin-purchased-from', componentId: newEntityId, vendorId: sourceEntityId };
    case 'origin-built-by':
      return { kind: 'origin-built-by', componentId: newEntityId, internalTeamId: sourceEntityId };
    default:
      return null;
  }
}
