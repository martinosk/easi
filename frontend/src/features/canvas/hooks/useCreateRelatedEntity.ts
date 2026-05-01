import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import {
  type AcquiredEntityId,
  type CapabilityId,
  type InternalTeamId,
  type Position,
  toAcquiredEntityId,
  toCapabilityId,
  toComponentId,
  toInternalTeamId,
  toVendorId,
  type VendorId,
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
import { planRelationCall, type RelationCallSpec } from '../utils/relationDispatch';

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
}

export interface UseCreateRelatedEntityResult {
  pending: PendingCreate | null;
  start: (params: PendingCreate) => void;
  cancel: () => void;
  handleEntityCreated: (entityId: string) => Promise<void>;
}

export function useCreateRelatedEntity(): UseCreateRelatedEntityResult {
  const [pending, setPending] = useState<PendingCreate | null>(null);
  const { currentViewId } = useCurrentView();
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const draftAddEntities = useAppStore((s) => s.draftAddEntities);

  const createRelation = useCreateRelation();
  const changeCapabilityParent = useChangeCapabilityParent();
  const linkSystemToCapability = useLinkSystemToCapability();
  const linkComponentToAcquiredEntity = useLinkComponentToAcquiredEntity();
  const linkComponentToVendor = useLinkComponentToVendor();
  const linkComponentToInternalTeam = useLinkComponentToInternalTeam();
  const addComponentToView = useAddComponentToView();
  const addCapabilityToView = useAddCapabilityToView();
  const addOriginEntityToView = useAddOriginEntityToView();

  const dispatchRelation = useCallback(
    async (spec: RelationCallSpec): Promise<void> => {
      switch (spec.kind) {
        case 'component-relation':
          await createRelation.mutateAsync({
            sourceComponentId: toComponentId(spec.sourceComponentId),
            targetComponentId: toComponentId(spec.targetComponentId),
            relationType: 'Triggers',
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
            request: {
              componentId: toComponentId(spec.componentId),
              realizationLevel: 'Full',
            },
          });
          return;
        case 'origin-acquired-via':
          await linkComponentToAcquiredEntity.mutateAsync({
            componentId: toComponentId(spec.componentId),
            entityId: toAcquiredEntityId(spec.acquiredEntityId) as AcquiredEntityId,
          });
          return;
        case 'origin-purchased-from':
          await linkComponentToVendor.mutateAsync({
            componentId: toComponentId(spec.componentId),
            vendorId: toVendorId(spec.vendorId) as VendorId,
          });
          return;
        case 'origin-built-by':
          await linkComponentToInternalTeam.mutateAsync({
            componentId: toComponentId(spec.componentId),
            teamId: toInternalTeamId(spec.internalTeamId) as InternalTeamId,
          });
          return;
      }
    },
    [
      createRelation,
      changeCapabilityParent,
      linkSystemToCapability,
      linkComponentToAcquiredEntity,
      linkComponentToVendor,
      linkComponentToInternalTeam,
    ],
  );

  const addToView = useCallback(
    async (targetType: RelatedTargetType, entityId: string, position: Position): Promise<void> => {
      if (!currentViewId) return;
      const viewId = currentViewId as ViewId;
      switch (targetType) {
        case 'component':
          await addComponentToView.mutateAsync({
            viewId,
            request: { componentId: toComponentId(entityId), x: position.x, y: position.y },
          });
          return;
        case 'capability':
          await addCapabilityToView.mutateAsync({
            viewId,
            request: { capabilityId: toCapabilityId(entityId) as CapabilityId, x: position.x, y: position.y },
          });
          return;
        case 'acquiredEntity':
        case 'vendor':
        case 'internalTeam':
          await addOriginEntityToView.mutateAsync({
            viewId,
            request: { originEntityId: entityId, x: position.x, y: position.y },
          });
          return;
      }
    },
    [addCapabilityToView, addComponentToView, addOriginEntityToView, currentViewId],
  );

  const start = useCallback((params: PendingCreate) => setPending(params), []);
  const cancel = useCallback(() => setPending(null), []);

  const handleEntityCreated = useCallback(
    async (entityId: string): Promise<void> => {
      const current = pending;
      if (!current) return;

      const targetPosition = computeOffsetPosition(current.sourcePosition, current.side);

      if (dynamicViewId) {
        draftAddEntities(
          [{ id: entityId, type: targetTypeToEntityType[current.entry.targetType] }],
          { [entityId]: targetPosition },
        );
        setPending(null);
        return;
      }

      const spec = planRelationCall(current.entry.relationType, current.sourceEntityId, entityId);
      if (!spec) {
        toast.error(`Unknown relation type "${current.entry.relationType}". Please retry manually.`);
        setPending(null);
        return;
      }

      try {
        await dispatchRelation(spec);
      } catch (error) {
        const message = error instanceof Error ? error.message : 'unknown error';
        toast.error(
          `Could not create the "${current.entry.title}" relation (${message}). The new entity remains; drag-connect from the source handle to retry.`,
        );
        setPending(null);
        return;
      }

      try {
        await addToView(current.entry.targetType, entityId, targetPosition);
      } catch (error) {
        const message = error instanceof Error ? error.message : 'unknown error';
        toast.error(`Could not place new entity on view (${message}).`);
      } finally {
        setPending(null);
      }
    },
    [pending, dynamicViewId, draftAddEntities, dispatchRelation, addToView],
  );

  return { pending, start, cancel, handleEntityCreated };
}
