import type { RelatedTargetType } from '../../../utils/xRelated';

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
  targetType?: RelatedTargetType,
): RelationCallSpec | null {
  switch (relationType) {
    case 'component-triggers':
      return componentRelation(sourceEntityId, newEntityId, 'Triggers');
    case 'component-serves':
      return componentRelation(sourceEntityId, newEntityId, 'Serves');
    case 'capability-parent':
      return { kind: 'capability-parent', childCapabilityId: newEntityId, parentCapabilityId: sourceEntityId };
    case 'capability-realization':
      return { kind: 'capability-realization', capabilityId: sourceEntityId, componentId: newEntityId };
    case 'origin-acquired-via':
      return originAcquiredVia(sourceEntityId, newEntityId, targetType);
    case 'origin-purchased-from':
      return originPurchasedFrom(sourceEntityId, newEntityId, targetType);
    case 'origin-built-by':
      return originBuiltBy(sourceEntityId, newEntityId, targetType);
    default:
      return null;
  }
}

function componentRelation(
  sourceComponentId: string,
  targetComponentId: string,
  relationSubType: RelationSubType,
): RelationCallSpec {
  return { kind: 'component-relation', sourceComponentId, targetComponentId, relationSubType };
}

function originAcquiredVia(
  sourceEntityId: string,
  newEntityId: string,
  targetType: RelatedTargetType | undefined,
): RelationCallSpec {
  const newIsAcquiredEntity = targetType === 'acquiredEntity';
  return {
    kind: 'origin-acquired-via',
    componentId: newIsAcquiredEntity ? sourceEntityId : newEntityId,
    acquiredEntityId: newIsAcquiredEntity ? newEntityId : sourceEntityId,
  };
}

function originPurchasedFrom(
  sourceEntityId: string,
  newEntityId: string,
  targetType: RelatedTargetType | undefined,
): RelationCallSpec {
  const newIsVendor = targetType === 'vendor';
  return {
    kind: 'origin-purchased-from',
    componentId: newIsVendor ? sourceEntityId : newEntityId,
    vendorId: newIsVendor ? newEntityId : sourceEntityId,
  };
}

function originBuiltBy(
  sourceEntityId: string,
  newEntityId: string,
  targetType: RelatedTargetType | undefined,
): RelationCallSpec {
  const newIsInternalTeam = targetType === 'internalTeam';
  return {
    kind: 'origin-built-by',
    componentId: newIsInternalTeam ? sourceEntityId : newEntityId,
    internalTeamId: newIsInternalTeam ? newEntityId : sourceEntityId,
  };
}
