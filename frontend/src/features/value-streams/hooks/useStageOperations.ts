import { useState, useCallback } from 'react';
import { useAddStage, useUpdateStage, useDeleteStage, useReorderStages, useAddStageCapability } from './useValueStreamStages';
import type { ValueStreamStage, ValueStreamDetail, CreateStageRequest, UpdateStageRequest } from '../../../api/types';

interface StageFormData {
  name: string;
  description: string;
}

const EMPTY_FORM: StageFormData = { name: '', description: '' };

export function useStageOperations(detail: ValueStreamDetail | undefined) {
  const addStageMutation = useAddStage();
  const updateStageMutation = useUpdateStage();
  const deleteStageMutation = useDeleteStage();
  const reorderStagesMutation = useReorderStages();
  const addCapabilityMutation = useAddStageCapability();

  const [showAddForm, setShowAddForm] = useState(false);
  const [editingStage, setEditingStage] = useState<ValueStreamStage | null>(null);
  const [formData, setFormData] = useState<StageFormData>(EMPTY_FORM);

  const isFormOpen = showAddForm || editingStage !== null;

  const openAddForm = useCallback(() => {
    setFormData(EMPTY_FORM);
    setShowAddForm(true);
  }, []);

  const openEditForm = useCallback((stage: ValueStreamStage) => {
    setEditingStage(stage);
    setFormData({ name: stage.name, description: stage.description || '' });
  }, []);

  const closeForm = useCallback(() => {
    setShowAddForm(false);
    setEditingStage(null);
    setFormData(EMPTY_FORM);
  }, []);

  const submitForm = useCallback(async () => {
    if (!formData.name.trim()) return;

    if (editingStage) {
      const request: UpdateStageRequest = {
        name: formData.name,
        description: formData.description || undefined,
      };
      await updateStageMutation.mutateAsync({ stage: editingStage, request });
    } else if (detail) {
      const request: CreateStageRequest = {
        name: formData.name,
        description: formData.description || undefined,
      };
      await addStageMutation.mutateAsync({ valueStream: detail, request });
    }
    closeForm();
  }, [formData, editingStage, detail, updateStageMutation, addStageMutation, closeForm]);

  const deleteStage = useCallback(async (stage: ValueStreamStage) => {
    await deleteStageMutation.mutateAsync(stage);
  }, [deleteStageMutation]);

  const reorderStages = useCallback(async (orderedStageIds: string[]) => {
    if (!detail) return;
    const positions = orderedStageIds.map((stageId, index) => ({
      stageId,
      position: index + 1,
    }));
    await reorderStagesMutation.mutateAsync({ valueStream: detail, request: { positions } });
  }, [detail, reorderStagesMutation]);

  const addCapability = useCallback((stageId: string, capabilityId: string) => {
    if (!detail) return;
    const stage = detail.stages.find(s => s.id === stageId);
    if (!stage) return;
    addCapabilityMutation.mutate({ stage, capabilityId });
  }, [detail, addCapabilityMutation]);

  return {
    isFormOpen,
    editingStage,
    formData,
    setFormData,
    openAddForm,
    openEditForm,
    closeForm,
    submitForm,
    deleteStage,
    reorderStages,
    addCapability,
  };
}
