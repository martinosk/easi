import { useState, useCallback, useMemo } from 'react';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useDeleteCapability, useChangeCapabilityParent, useDeleteRealization } from '../../capabilities/hooks/useCapabilities';
import { useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteRelation } from '../../relations/hooks/useRelations';
import { useRemoveComponentFromView, useRemoveCapabilityFromView } from '../../views/hooks/useViews';
import { useComponents } from '../../components/hooks/useComponents';
import { useRelations } from '../../relations/hooks/useRelations';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import type { ComponentId, CapabilityId, ViewId, Component, Relation, Capability, HATEOASLinks, RealizationId } from '../../../api/types';

export type DeleteTargetType =
  | 'component-from-view'
  | 'component-from-model'
  | 'relation-from-model'
  | 'capability-from-canvas'
  | 'capability-from-model'
  | 'parent-relation'
  | 'realization';

export interface DeleteTarget {
  type: DeleteTargetType;
  id: string;
  name: string;
  childId?: string;
  capabilityId?: CapabilityId;
  componentId?: ComponentId;
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

function useDeleteHandlers() {
  const removeComponentFromViewMutation = useRemoveComponentFromView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();
  const deleteComponentMutation = useDeleteComponent();
  const deleteRelationMutation = useDeleteRelation();
  const deleteCapabilityMutation = useDeleteCapability();
  const changeCapabilityParentMutation = useChangeCapabilityParent();
  const deleteRealizationMutation = useDeleteRealization();

  return useMemo((): Record<DeleteTargetType, DeleteHandler> => ({
    'component-from-view': async (target, viewId) => {
      if (!viewId) return;
      await removeComponentFromViewMutation.mutateAsync({
        viewId,
        componentId: target.id as ComponentId,
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
        capabilityId: target.id as CapabilityId,
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
        id: target.childId as CapabilityId,
        oldParentId: target.id,
        newParentId: null,
      });
    },
    'realization': async (target) => {
      if (!target.capabilityId || !target.componentId || !target._links) return;
      await deleteRealizationMutation.mutateAsync({
        id: target.id as RealizationId,
        capabilityId: target.capabilityId,
        componentId: target.componentId,
        realizationLevel: 'Full',
        origin: 'Direct',
        linkedAt: '',
        _links: target._links,
      });
    },
  }), [
    removeComponentFromViewMutation,
    removeCapabilityFromViewMutation,
    deleteComponentMutation,
    deleteRelationMutation,
    deleteCapabilityMutation,
    changeCapabilityParentMutation,
    deleteRealizationMutation,
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
