import type { RelatedTargetType } from '../../../utils/xRelated';

export type RelationSubType = 'Triggers' | 'Serves';

export type ContainmentSubType = 'composition' | 'aggregation';

export type RelationCallSpec =
  | {
      kind: 'component-relation';
      sourceComponentId: string;
      targetComponentId: string;
      relationSubType: RelationSubType;
    }
  | { kind: 'component-containment'; partId: string; parentId: string; containmentKind: ContainmentSubType }
  | { kind: 'capability-parent'; childCapabilityId: string; parentCapabilityId: string }
  | { kind: 'capability-realization'; capabilityId: string; componentId: string }
  | { kind: 'origin-acquired-via'; componentId: string; acquiredEntityId: string }
  | { kind: 'origin-purchased-from'; componentId: string; vendorId: string }
  | { kind: 'origin-built-by'; componentId: string; internalTeamId: string };

interface PlanInput {
  sourceEntityId: string;
  newEntityId: string;
  targetType?: RelatedTargetType;
}

type RelationPlanner = (input: PlanInput) => RelationCallSpec;

const PLANNERS: Record<string, RelationPlanner> = {
  'component-triggers': (input) => componentRelation(input, 'Triggers'),
  'component-serves': (input) => componentRelation(input, 'Serves'),
  'component-part-composition': (input) => componentContainment(input, 'composition'),
  'component-part-aggregation': (input) => componentContainment(input, 'aggregation'),
  'capability-parent': ({ sourceEntityId, newEntityId }) => ({
    kind: 'capability-parent',
    childCapabilityId: newEntityId,
    parentCapabilityId: sourceEntityId,
  }),
  'capability-realization': ({ sourceEntityId, newEntityId }) => ({
    kind: 'capability-realization',
    capabilityId: sourceEntityId,
    componentId: newEntityId,
  }),
  'origin-acquired-via': (input) => {
    const { componentId, entityId } = originPair(input, 'acquiredEntity');
    return { kind: 'origin-acquired-via', componentId, acquiredEntityId: entityId };
  },
  'origin-purchased-from': (input) => {
    const { componentId, entityId } = originPair(input, 'vendor');
    return { kind: 'origin-purchased-from', componentId, vendorId: entityId };
  },
  'origin-built-by': (input) => {
    const { componentId, entityId } = originPair(input, 'internalTeam');
    return { kind: 'origin-built-by', componentId, internalTeamId: entityId };
  },
};

export function planRelationCall(
  relationType: string,
  sourceEntityId: string,
  newEntityId: string,
  targetType?: RelatedTargetType,
): RelationCallSpec | null {
  const planner = PLANNERS[relationType];
  return planner ? planner({ sourceEntityId, newEntityId, targetType }) : null;
}

function componentRelation(input: PlanInput, relationSubType: RelationSubType): RelationCallSpec {
  return {
    kind: 'component-relation',
    sourceComponentId: input.sourceEntityId,
    targetComponentId: input.newEntityId,
    relationSubType,
  };
}

function componentContainment(input: PlanInput, containmentKind: ContainmentSubType): RelationCallSpec {
  return { kind: 'component-containment', partId: input.newEntityId, parentId: input.sourceEntityId, containmentKind };
}

function originPair(input: PlanInput, entityType: RelatedTargetType): { componentId: string; entityId: string } {
  const newIsEntity = input.targetType === entityType;
  return {
    componentId: newIsEntity ? input.sourceEntityId : input.newEntityId,
    entityId: newIsEntity ? input.newEntityId : input.sourceEntityId,
  };
}
