import { useState, useCallback, useMemo } from 'react';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useDeleteCapability, useChangeCapabilityParent, useDeleteRealization } from '../../capabilities/hooks/useCapabilities';
import { useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteRelation } from '../../relations/hooks/useRelations';
import { useRemoveComponentFromView, useRemoveCapabilityFromView } from '../../views/hooks/useViews';
import type { ComponentId, RelationId, CapabilityId, RealizationId, ViewId } from '../../../api/types';

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
}

type DeleteHandler = (target: DeleteTarget, viewId: ViewId | null) => Promise<void>;

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
    'component-from-model': async (target) => {
      await deleteComponentMutation.mutateAsync(target.id as ComponentId);
    },
    'relation-from-model': async (target) => {
      await deleteRelationMutation.mutateAsync(target.id as RelationId);
    },
    'capability-from-canvas': async (target, viewId) => {
      if (!viewId) return;
      await removeCapabilityFromViewMutation.mutateAsync({
        viewId,
        capabilityId: target.id as CapabilityId,
      });
    },
    'capability-from-model': async (target) => {
      await deleteCapabilityMutation.mutateAsync({ id: target.id as CapabilityId });
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
      if (!target.capabilityId || !target.componentId) return;
      await deleteRealizationMutation.mutateAsync({
        id: target.id as RealizationId,
        capabilityId: target.capabilityId,
        componentId: target.componentId,
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

  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      const handler = handlers[deleteTarget.type];
      await handler(deleteTarget, currentViewId);
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  }, [deleteTarget, handlers, currentViewId]);

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
