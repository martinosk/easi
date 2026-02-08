import { useState, useCallback, useMemo } from 'react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useDeleteCapability, useChangeCapabilityParent, useDeleteRealization } from '../../capabilities/hooks/useCapabilities';
import { useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteRelation } from '../../relations/hooks/useRelations';
import { useRemoveComponentFromView, useRemoveCapabilityFromView } from '../../views/hooks/useViews';
import { useComponents } from '../../components/hooks/useComponents';
import { useRelations } from '../../relations/hooks/useRelations';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useDeleteAcquiredEntity, useUnlinkComponentFromAcquiredEntity } from '../../origin-entities/hooks/useAcquiredEntities';
import { useDeleteVendor, useUnlinkComponentFromVendor } from '../../origin-entities/hooks/useVendors';
import { useDeleteInternalTeam, useUnlinkComponentFromInternalTeam } from '../../origin-entities/hooks/useInternalTeams';
import { toComponentId, toCapabilityId, toRealizationId } from '../../../api/types';
import type { CapabilityId, ComponentId, ViewId, Component, Relation, Capability, HATEOASLinks, AcquiredEntityId, VendorId, InternalTeamId, OriginRelationshipId, OriginRelationshipType } from '../../../api/types';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';

export type DeleteTargetType =
  | 'component-from-view'
  | 'component-from-model'
  | 'relation-from-model'
  | 'capability-from-canvas'
  | 'capability-from-model'
  | 'parent-relation'
  | 'realization'
  | 'origin-entity-from-model'
  | 'origin-relationship';

export interface DeleteTarget {
  type: DeleteTargetType;
  id: string;
  name: string;
  childId?: string;
  capabilityId?: CapabilityId;
  componentId?: ComponentId;
  originEntityType?: OriginEntityType;
  originRelationshipId?: OriginRelationshipId;
  originRelationshipType?: OriginRelationshipType;
  originEntityId?: string;
  _links?: HATEOASLinks;
}

type DeleteHandler = (
  target: DeleteTarget,
  viewId: ViewId | null,
  lookups: {
    components: Component[];
    relations: Relation[];
    capabilities: Capability[];
  }
) => Promise<void>;

interface OriginEntityDeleteMutations {
  deleteAcquired: { mutateAsync: (params: { id: AcquiredEntityId; name: string }) => Promise<void> };
  deleteVendor: { mutateAsync: (params: { id: VendorId; name: string }) => Promise<void> };
  deleteTeam: { mutateAsync: (params: { id: InternalTeamId; name: string }) => Promise<void> };
}

interface OriginRelationshipUnlinkMutations {
  unlinkAcquired: { mutateAsync: (params: { entityId: AcquiredEntityId; componentId: ComponentId }) => Promise<void> };
  unlinkVendor: { mutateAsync: (params: { vendorId: VendorId; componentId: ComponentId }) => Promise<void> };
  unlinkTeam: { mutateAsync: (params: { teamId: InternalTeamId; componentId: ComponentId }) => Promise<void> };
}

type OriginEntityDeleteStrategy = (
  mutations: OriginEntityDeleteMutations,
  entityId: string,
  name: string
) => Promise<void>;

type OriginRelationshipUnlinkStrategy = (
  mutations: OriginRelationshipUnlinkMutations,
  originEntityId: string,
  componentId: ComponentId
) => Promise<void>;

const ORIGIN_ENTITY_DELETE_STRATEGIES: Record<OriginEntityType, OriginEntityDeleteStrategy> = {
  acquired: (mutations, entityId, name) =>
    mutations.deleteAcquired.mutateAsync({ id: entityId as AcquiredEntityId, name }),
  vendor: (mutations, entityId, name) =>
    mutations.deleteVendor.mutateAsync({ id: entityId as VendorId, name }),
  team: (mutations, entityId, name) =>
    mutations.deleteTeam.mutateAsync({ id: entityId as InternalTeamId, name }),
};

const ORIGIN_RELATIONSHIP_UNLINK_STRATEGIES: Record<OriginRelationshipType, OriginRelationshipUnlinkStrategy> = {
  AcquiredVia: (mutations, originEntityId, componentId) =>
    mutations.unlinkAcquired.mutateAsync({ entityId: originEntityId as AcquiredEntityId, componentId }),
  PurchasedFrom: (mutations, originEntityId, componentId) =>
    mutations.unlinkVendor.mutateAsync({ vendorId: originEntityId as VendorId, componentId }),
  BuiltBy: (mutations, originEntityId, componentId) =>
    mutations.unlinkTeam.mutateAsync({ teamId: originEntityId as InternalTeamId, componentId }),
};

function hasRealizationData(target: DeleteTarget): boolean {
  return Boolean(target.capabilityId && target.componentId && target._links);
}

function hasOriginRelationshipData(target: DeleteTarget): boolean {
  return Boolean(target.componentId && target.originRelationshipType && target.originEntityId);
}

function useDeleteHandlers() {
  const removeComponentFromViewMutation = useRemoveComponentFromView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();
  const deleteComponentMutation = useDeleteComponent();
  const deleteRelationMutation = useDeleteRelation();
  const deleteCapabilityMutation = useDeleteCapability();
  const changeCapabilityParentMutation = useChangeCapabilityParent();
  const deleteRealizationMutation = useDeleteRealization();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();
  const unlinkFromAcquiredMutation = useUnlinkComponentFromAcquiredEntity();
  const unlinkFromVendorMutation = useUnlinkComponentFromVendor();
  const unlinkFromInternalTeamMutation = useUnlinkComponentFromInternalTeam();

  const originEntityDeleteMutations: OriginEntityDeleteMutations = useMemo(() => ({
    deleteAcquired: deleteAcquiredEntityMutation,
    deleteVendor: deleteVendorMutation,
    deleteTeam: deleteInternalTeamMutation,
  }), [deleteAcquiredEntityMutation, deleteVendorMutation, deleteInternalTeamMutation]);

  const originRelationshipUnlinkMutations: OriginRelationshipUnlinkMutations = useMemo(() => ({
    unlinkAcquired: unlinkFromAcquiredMutation,
    unlinkVendor: unlinkFromVendorMutation,
    unlinkTeam: unlinkFromInternalTeamMutation,
  }), [unlinkFromAcquiredMutation, unlinkFromVendorMutation, unlinkFromInternalTeamMutation]);

  return useMemo((): Record<DeleteTargetType, DeleteHandler> => ({
    'component-from-view': async (target, viewId) => {
      if (!viewId) return;
      await removeComponentFromViewMutation.mutateAsync({
        viewId,
        componentId: toComponentId(target.id),
      });
    },
    'component-from-model': async (target, _viewId, lookups) => {
      const component = lookups.components.find(c => c.id === target.id);
      if (!component) return;
      await deleteComponentMutation.mutateAsync(component);
    },
    'relation-from-model': async (target, _viewId, lookups) => {
      const relation = lookups.relations.find(r => r.id === target.id);
      if (!relation) return;
      await deleteRelationMutation.mutateAsync(relation);
    },
    'capability-from-canvas': async (target, viewId) => {
      if (!viewId) return;
      await removeCapabilityFromViewMutation.mutateAsync({
        viewId,
        capabilityId: toCapabilityId(target.id),
      });
    },
    'capability-from-model': async (target, _viewId, lookups) => {
      const capability = lookups.capabilities.find(c => c.id === target.id);
      if (!capability) return;
      await deleteCapabilityMutation.mutateAsync({ capability });
    },
    'parent-relation': async (target) => {
      if (!target.childId) return;
      await changeCapabilityParentMutation.mutateAsync({
        id: toCapabilityId(target.childId),
        oldParentId: target.id,
        newParentId: null,
      });
    },
    'realization': async (target) => {
      if (!hasRealizationData(target)) return;
      await deleteRealizationMutation.mutateAsync({
        id: toRealizationId(target.id),
        capabilityId: target.capabilityId!,
        componentId: target.componentId!,
        realizationLevel: 'Full',
        origin: 'Direct',
        linkedAt: '',
        _links: target._links!,
      });
    },
    'origin-entity-from-model': async (target) => {
      if (!target.originEntityType) return;

      const strategy = ORIGIN_ENTITY_DELETE_STRATEGIES[target.originEntityType];
      await strategy(originEntityDeleteMutations, target.id, target.name);
    },
    'origin-relationship': async (target) => {
      if (!hasOriginRelationshipData(target)) return;

      const strategy = ORIGIN_RELATIONSHIP_UNLINK_STRATEGIES[target.originRelationshipType!];
      await strategy(originRelationshipUnlinkMutations, target.originEntityId!, target.componentId!);
    },
  }), [
    removeComponentFromViewMutation,
    removeCapabilityFromViewMutation,
    deleteComponentMutation,
    deleteRelationMutation,
    deleteCapabilityMutation,
    changeCapabilityParentMutation,
    deleteRealizationMutation,
    originEntityDeleteMutations,
    originRelationshipUnlinkMutations,
  ]);
}

export const useDeleteConfirmation = () => {
  const { currentViewId } = useCurrentView();
  const handlers = useDeleteHandlers();

  const { data: components = [] } = useComponents();
  const { data: relations = [] } = useRelations();
  const { data: capabilities = [] } = useCapabilities();

  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      const handler = handlers[deleteTarget.type];
      await handler(deleteTarget, currentViewId, { components, relations, capabilities });
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  }, [deleteTarget, handlers, currentViewId, components, relations, capabilities]);

  const handleDeleteCancel = useCallback(() => {
    setDeleteTarget(null);
  }, []);

  return {
    deleteTarget,
    isDeleting,
    setDeleteTarget,
    handleDeleteConfirm,
    handleDeleteCancel,
  };
};
