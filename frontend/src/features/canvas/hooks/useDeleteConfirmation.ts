import { useState, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';

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

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      if (deleteTarget.type === 'component-from-view') {
        await removeComponentFromView(deleteTarget.id);
      } else if (deleteTarget.type === 'component-from-model') {
        await deleteComponent(deleteTarget.id);
      } else if (deleteTarget.type === 'relation-from-model') {
        await deleteRelation(deleteTarget.id);
      } else if (deleteTarget.type === 'capability-from-canvas') {
        removeCapabilityFromCanvas(deleteTarget.id);
      } else if (deleteTarget.type === 'capability-from-model') {
        await deleteCapability(deleteTarget.id);
      } else if (deleteTarget.type === 'parent-relation' && deleteTarget.childId) {
        await changeCapabilityParent(deleteTarget.childId, null);
      } else if (deleteTarget.type === 'realization') {
        await deleteRealization(deleteTarget.id);
      }
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  }, [
    deleteTarget,
    removeComponentFromView,
    deleteComponent,
    deleteRelation,
    removeCapabilityFromCanvas,
    deleteCapability,
    changeCapabilityParent,
    deleteRealization,
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
