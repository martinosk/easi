import { useState, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import type { ComponentId, RelationId, CapabilityId, RealizationId } from '../../../api/types';

export interface DeleteTarget {
  type: 'component-from-view' | 'component-from-model' | 'relation-from-model' | 'capability-from-canvas' | 'capability-from-model' | 'parent-relation' | 'realization';
  id: string;
  name: string;
  childId?: string;
}

export const useDeleteConfirmation = () => {
  const removeComponentFromView = useAppStore((state) => state.removeComponentFromView);
  const deleteComponent = useAppStore((state) => state.deleteComponent);
  const deleteRelation = useAppStore((state) => state.deleteRelation);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);
  const deleteCapability = useAppStore((state) => state.deleteCapability);
  const changeCapabilityParent = useAppStore((state) => state.changeCapabilityParent);
  const deleteRealization = useAppStore((state) => state.deleteRealization);

  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const executeDelete = useCallback(async (target: DeleteTarget) => {
    switch (target.type) {
      case 'component-from-view':
        await removeComponentFromView(target.id as ComponentId);
        break;
      case 'component-from-model':
        await deleteComponent(target.id as ComponentId);
        break;
      case 'relation-from-model':
        await deleteRelation(target.id as RelationId);
        break;
      case 'capability-from-canvas':
        removeCapabilityFromCanvas(target.id as CapabilityId);
        break;
      case 'capability-from-model':
        await deleteCapability(target.id as CapabilityId);
        break;
      case 'parent-relation':
        if (target.childId) {
          await changeCapabilityParent(target.childId as CapabilityId, null);
        }
        break;
      case 'realization':
        await deleteRealization(target.id as RealizationId);
        break;
    }
  }, [
    removeComponentFromView,
    deleteComponent,
    deleteRelation,
    removeCapabilityFromCanvas,
    deleteCapability,
    changeCapabilityParent,
    deleteRealization,
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
