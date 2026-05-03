import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import {
  type Position,
  toAcquiredEntityId,
  toCapabilityId,
  toComponentId,
  toInternalTeamId,
  toVendorId,
  type ViewId,
} from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import type { RelatedLink, RelatedTargetType } from '../../../utils/xRelated';
import { useChangeCapabilityParent, useLinkSystemToCapability } from '../../capabilities/hooks/useCapabilities';
import {
  useLinkComponentToAcquiredEntity,
  useLinkComponentToInternalTeam,
  useLinkComponentToVendor,
} from '../../origin-entities/hooks';
import { useCreateRelation } from '../../relations/hooks/useRelations';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAddCapabilityToView, useAddComponentToView, useAddOriginEntityToView } from '../../views/hooks/useViews';
import type { HandleSide } from '../utils/handleClick';
import { computeOffsetPosition } from '../utils/offsetPosition';
import { planRelationCall, type RelationCallSpec, type RelationSubType } from '../utils/relationDispatch';

const targetTypeToEntityType: Record<RelatedTargetType, 'component' | 'capability' | 'originEntity'> = {
  component: 'component',
  capability: 'capability',
  acquiredEntity: 'originEntity',
  vendor: 'originEntity',
  internalTeam: 'originEntity',
};

export interface PendingCreate {
  entry: RelatedLink;
  sourceEntityId: string;
  side: HandleSide;
  sourcePosition: Position;
  prefill?: { capabilityLevel?: 'L1' | 'L2' | 'L3' | 'L4' };
  relationSubType?: RelationSubType;
}

export interface UseCreateRelatedEntityResult {
  pending: PendingCreate | null;
  start: (params: PendingCreate) => void;
  cancel: () => void;
  handleEntityCreated: (entityId: string) => Promise<void>;
}

function useRelationDispatcher(): (spec: RelationCallSpec) => Promise<void> {
  const createRelation = useCreateRelation();
  const changeCapabilityParent = useChangeCapabilityParent();
  const linkSystemToCapability = useLinkSystemToCapability();
  const linkAcquiredEntity = useLinkComponentToAcquiredEntity();
  const linkVendor = useLinkComponentToVendor();
  const linkInternalTeam = useLinkComponentToInternalTeam();

  return useCallback(
    async (spec: RelationCallSpec): Promise<void> => {
      switch (spec.kind) {
        case 'component-relation':
          await createRelation.mutateAsync({
            sourceComponentId: toComponentId(spec.sourceComponentId),
            targetComponentId: toComponentId(spec.targetComponentId),
            relationType: spec.relationSubType,
          });
          return;
        case 'capability-parent':
          await changeCapabilityParent.mutateAsync({
            id: toCapabilityId(spec.childCapabilityId),
            newParentId: toCapabilityId(spec.parentCapabilityId),
          });
          return;
        case 'capability-realization':
          await linkSystemToCapability.mutateAsync({
            capabilityId: toCapabilityId(spec.capabilityId),
            request: { componentId: toComponentId(spec.componentId), realizationLevel: 'Full' },
          });
          return;
        case 'origin-acquired-via':
          await linkAcquiredEntity.mutateAsync({
            componentId: toComponentId(spec.componentId),
            entityId: toAcquiredEntityId(spec.acquiredEntityId),
          });
          return;
        case 'origin-purchased-from':
          await linkVendor.mutateAsync({
            componentId: toComponentId(spec.componentId),
            vendorId: toVendorId(spec.vendorId),
          });
          return;
        case 'origin-built-by':
          await linkInternalTeam.mutateAsync({
            componentId: toComponentId(spec.componentId),
            teamId: toInternalTeamId(spec.internalTeamId),
          });
          return;
      }
    },
    [createRelation, changeCapabilityParent, linkSystemToCapability, linkAcquiredEntity, linkVendor, linkInternalTeam],
  );
}

function useViewAdder(): (targetType: RelatedTargetType, entityId: string, position: Position) => Promise<void> {
  const { currentViewId } = useCurrentView();
  const addComponent = useAddComponentToView();
  const addCapability = useAddCapabilityToView();
  const addOrigin = useAddOriginEntityToView();

  return useCallback(
    async (targetType, entityId, position): Promise<void> => {
      if (!currentViewId) return;
      const viewId = currentViewId as ViewId;
      const xy = { x: position.x, y: position.y };
      switch (targetType) {
        case 'component':
          await addComponent.mutateAsync({ viewId, request: { componentId: toComponentId(entityId), ...xy } });
          return;
        case 'capability':
          await addCapability.mutateAsync({ viewId, request: { capabilityId: toCapabilityId(entityId), ...xy } });
          return;
        case 'acquiredEntity':
        case 'vendor':
        case 'internalTeam':
          await addOrigin.mutateAsync({ viewId, request: { originEntityId: entityId, ...xy } });
          return;
      }
    },
    [addComponent, addCapability, addOrigin, currentViewId],
  );
}

export function useCreateRelatedEntity(): UseCreateRelatedEntityResult {
  const [pending, setPending] = useState<PendingCreate | null>(null);
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const draftAddEntities = useAppStore((s) => s.draftAddEntities);

  const dispatchRelation = useRelationDispatcher();
  const addToView = useViewAdder();

  const start = useCallback((params: PendingCreate) => setPending(params), []);
  const cancel = useCallback(() => setPending(null), []);

  const handleEntityCreated = useCallback(
    async (entityId: string): Promise<void> => {
      const current = pending;
      if (!current) return;

      const targetPosition = computeOffsetPosition(current.sourcePosition, current.side);
      const spec = planRelationCall(
        current.entry.relationType,
        current.sourceEntityId,
        entityId,
        current.relationSubType,
      );

      await runRegularModePersist(spec, current, dispatchRelation);

      if (dynamicViewId) {
        draftAddEntities(
          [{ id: entityId, type: targetTypeToEntityType[current.entry.targetType] }],
          { [entityId]: targetPosition },
        );
        setPending(null);
        return;
      }

      await safeAddToView(addToView, current.entry.targetType, entityId, targetPosition);
      setPending(null);
    },
    [pending, dynamicViewId, draftAddEntities, dispatchRelation, addToView],
  );

  return { pending, start, cancel, handleEntityCreated };
}

async function runRegularModePersist(
  spec: RelationCallSpec | null,
  current: PendingCreate,
  dispatchRelation: (spec: RelationCallSpec) => Promise<void>,
): Promise<void> {
  if (!spec) {
    toast.error(`Unknown relation type "${current.entry.relationType}". Please retry manually.`);
    return;
  }
  try {
    await dispatchRelation(spec);
  } catch (error) {
    const message = error instanceof Error ? error.message : 'unknown error';
    toast.error(
      `Could not create the "${current.entry.title}" relation (${message}). The new entity remains; drag-connect from the source handle to retry.`,
    );
  }
}

async function safeAddToView(
  addToView: (targetType: RelatedTargetType, entityId: string, position: Position) => Promise<void>,
  targetType: RelatedTargetType,
  entityId: string,
  targetPosition: Position,
): Promise<void> {
  try {
    await addToView(targetType, entityId, targetPosition);
  } catch (error) {
    const message = error instanceof Error ? error.message : 'unknown error';
    toast.error(`Could not place new entity on view (${message}).`);
  }
}
