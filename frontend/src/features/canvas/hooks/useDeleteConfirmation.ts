import { useState, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useDeleteCapability, useChangeCapabilityParent, useDeleteRealization } from '../../capabilities/hooks/useCapabilities';
import { useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteRelation } from '../../relations/hooks/useRelations';
import { useRemoveComponentFromView } from '../../views/hooks/useViews';
import type { ComponentId, RelationId, CapabilityId, RealizationId, ViewId } from '../../../api/types';

export interface DeleteTarget {
  type: 'component-from-view' | 'component-from-model' | 'relation-from-model' | 'capability-from-canvas' | 'capability-from-model' | 'parent-relation' | 'realization';
  id: string;
  name: string;
  childId?: string;
}

export const useDeleteConfirmation = () => {
  const currentView = useAppStore((state) => state.currentView);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);

  const removeComponentFromViewMutation = useRemoveComponentFromView();
  const deleteComponentMutation = useDeleteComponent();
  const deleteRelationMutation = useDeleteRelation();
  const deleteCapabilityMutation = useDeleteCapability();
  const changeCapabilityParentMutation = useChangeCapabilityParent();
  const deleteRealizationMutation = useDeleteRealization();

  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const executeDelete = useCallback(async (target: DeleteTarget) => {
    switch (target.type) {
      case 'component-from-view':
        if (currentView) {
          await removeComponentFromViewMutation.mutateAsync({
            viewId: currentView.id as ViewId,
            componentId: target.id as ComponentId
          });
        }
        break;
      case 'component-from-model':
        await deleteComponentMutation.mutateAsync(target.id as ComponentId);
        break;
      case 'relation-from-model':
        await deleteRelationMutation.mutateAsync(target.id as RelationId);
        break;
      case 'capability-from-canvas':
        removeCapabilityFromCanvas(target.id as CapabilityId);
        break;
      case 'capability-from-model':
        await deleteCapabilityMutation.mutateAsync(target.id as CapabilityId);
        break;
      case 'parent-relation':
        if (target.childId) {
          await changeCapabilityParentMutation.mutateAsync({ id: target.childId as CapabilityId, parentId: null });
        }
        break;
      case 'realization':
        await deleteRealizationMutation.mutateAsync(target.id as RealizationId);
        break;
    }
  }, [
    currentView,
    removeComponentFromViewMutation,
    deleteComponentMutation,
    deleteRelationMutation,
    removeCapabilityFromCanvas,
    deleteCapabilityMutation,
    changeCapabilityParentMutation,
    deleteRealizationMutation,
  ]);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      await executeDelete(deleteTarget);
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  }, [deleteTarget, executeDelete]);

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
