import { useCallback, useMemo, useState } from 'react';
import type {
  AcquiredEntityId,
  Capability,
  CapabilityId,
  Component,
  ComponentId,
  HATEOASLinks,
  InternalTeamId,
  OriginRelationshipId,
  OriginRelationshipType,
  Relation,
  VendorId,
  ViewId,
} from '../../../api/types';
import { toCapabilityId, toComponentId, toRealizationId } from '../../../api/types';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';
import {
  useCapabilities,
  useCascadeDeleteCapability,
  useChangeCapabilityParent,
  useDeleteCapability,
  useDeleteRealization,
  useRealizations,
} from '../../capabilities/hooks/useCapabilities';
import { useComponents, useDeleteComponent } from '../../components/hooks/useComponents';
import {
  useDeleteAcquiredEntity,
  useUnlinkComponentFromAcquiredEntity,
} from '../../origin-entities/hooks/useAcquiredEntities';
import {
  useDeleteInternalTeam,
  useUnlinkComponentFromInternalTeam,
} from '../../origin-entities/hooks/useInternalTeams';
import { useDeleteVendor, useUnlinkComponentFromVendor } from '../../origin-entities/hooks/useVendors';
import { useDeleteRelation, useRelations } from '../../relations/hooks/useRelations';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useRemoveCapabilityFromView, useRemoveComponentFromView } from '../../views/hooks/useViews';
import { useAppStore } from '../../../store/appStore';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { computeOrphans } from '../utils/dynamicMode';

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
  },
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
  name: string,
) => Promise<void>;

type OriginRelationshipUnlinkStrategy = (
  mutations: OriginRelationshipUnlinkMutations,
  originEntityId: string,
  componentId: ComponentId,
) => Promise<void>;

const ORIGIN_ENTITY_DELETE_STRATEGIES: Record<OriginEntityType, OriginEntityDeleteStrategy> = {
  acquired: (mutations, entityId, name) =>
    mutations.deleteAcquired.mutateAsync({ id: entityId as AcquiredEntityId, name }),
  vendor: (mutations, entityId, name) => mutations.deleteVendor.mutateAsync({ id: entityId as VendorId, name }),
  team: (mutations, entityId, name) => mutations.deleteTeam.mutateAsync({ id: entityId as InternalTeamId, name }),
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
  const cascadeDeleteCapabilityMutation = useCascadeDeleteCapability();
  const changeCapabilityParentMutation = useChangeCapabilityParent();
  const deleteRealizationMutation = useDeleteRealization();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();
  const unlinkFromAcquiredMutation = useUnlinkComponentFromAcquiredEntity();
  const unlinkFromVendorMutation = useUnlinkComponentFromVendor();
  const unlinkFromInternalTeamMutation = useUnlinkComponentFromInternalTeam();

  const originEntityDeleteMutations: OriginEntityDeleteMutations = useMemo(
    () => ({
      deleteAcquired: deleteAcquiredEntityMutation,
      deleteVendor: deleteVendorMutation,
      deleteTeam: deleteInternalTeamMutation,
    }),
    [deleteAcquiredEntityMutation, deleteVendorMutation, deleteInternalTeamMutation],
  );

  const originRelationshipUnlinkMutations: OriginRelationshipUnlinkMutations = useMemo(
    () => ({
      unlinkAcquired: unlinkFromAcquiredMutation,
      unlinkVendor: unlinkFromVendorMutation,
      unlinkTeam: unlinkFromInternalTeamMutation,
    }),
    [unlinkFromAcquiredMutation, unlinkFromVendorMutation, unlinkFromInternalTeamMutation],
  );

  return useMemo(
    (): Record<DeleteTargetType, DeleteHandler> => ({
      'component-from-view': async (target, viewId) => {
        if (!viewId) return;
        await removeComponentFromViewMutation.mutateAsync({
          viewId,
          componentId: toComponentId(target.id),
        });
      },
      'component-from-model': async (target, _viewId, lookups) => {
        const component = lookups.components.find((c) => c.id === target.id);
        if (!component) return;
        await deleteComponentMutation.mutateAsync(component);
      },
      'relation-from-model': async (target, _viewId, lookups) => {
        const relation = lookups.relations.find((r) => r.id === target.id);
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
        const capability = lookups.capabilities.find((c) => c.id === target.id);
        if (!capability) return;
        await cascadeDeleteCapabilityMutation.mutateAsync({
          capability,
          cascade: false,
          deleteRealisingApplications: false,
          parentId: capability.parentId ?? undefined,
        });
      },
      'parent-relation': async (target) => {
        if (!target.childId) return;
        await changeCapabilityParentMutation.mutateAsync({
          id: toCapabilityId(target.childId),
          oldParentId: target.id,
          newParentId: null,
        });
      },
      realization: async (target) => {
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
    }),
    [
      removeComponentFromViewMutation,
      removeCapabilityFromViewMutation,
      deleteComponentMutation,
      deleteRelationMutation,
      deleteCapabilityMutation,
      cascadeDeleteCapabilityMutation,
      changeCapabilityParentMutation,
      deleteRealizationMutation,
      originEntityDeleteMutations,
      originRelationshipUnlinkMutations,
    ],
  );
}

const DYNAMIC_REMOVE_FROM_VIEW_TYPES = new Set<DeleteTargetType>([
  'component-from-view',
  'capability-from-canvas',
]);

const CASCADE_CONFIRM_THRESHOLD = 5;

interface DynamicRemovalContext {
  data: Parameters<typeof computeOrphans>[0];
  dynamicEntities: Parameters<typeof computeOrphans>[1];
  dynamicFilters: Parameters<typeof computeOrphans>[3];
  draftRemoveEntities: (ids: string[]) => void;
}

function performDynamicRemoval(removedId: string, ctx: DynamicRemovalContext): boolean {
  const orphans = computeOrphans(ctx.data, ctx.dynamicEntities, removedId, ctx.dynamicFilters);
  if (orphans.length >= CASCADE_CONFIRM_THRESHOLD) {
    const total = 1 + orphans.length;
    const plural = orphans.length === 1 ? '' : 's';
    const proceed = window.confirm(
      `Removing this entity will also remove ${orphans.length} orphaned descendant${plural} (${total} total). Continue?`,
    );
    if (!proceed) return false;
  }
  ctx.draftRemoveEntities([removedId, ...orphans]);
  return true;
}

export const useDeleteConfirmation = () => {
  const { currentViewId } = useCurrentView();
  const handlers = useDeleteHandlers();

  const { data: components = [] } = useComponents();
  const { data: relations = [] } = useRelations();
  const { data: capabilities = [] } = useCapabilities();
  const { data: realizations = [] } = useRealizations();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  const dynamicEnabled = useAppStore((s) => s.dynamicEnabled);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);
  const dynamicFilters = useAppStore((s) => s.dynamicFilters);
  const draftRemoveEntities = useAppStore((s) => s.draftRemoveEntities);

  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;

    if (dynamicEnabled && DYNAMIC_REMOVE_FROM_VIEW_TYPES.has(deleteTarget.type)) {
      performDynamicRemoval(deleteTarget.id, {
        data: { relations, capabilities, realizations, originRelationships },
        dynamicEntities,
        dynamicFilters,
        draftRemoveEntities,
      });
      setDeleteTarget(null);
      return;
    }

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
  }, [
    deleteTarget,
    handlers,
    currentViewId,
    components,
    relations,
    capabilities,
    realizations,
    originRelationships,
    dynamicEnabled,
    dynamicEntities,
    dynamicFilters,
    draftRemoveEntities,
  ]);

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
