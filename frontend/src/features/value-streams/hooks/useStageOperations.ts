import { useState, useCallback } from 'react';
import { useAddStage, useUpdateStage, useDeleteStage, useReorderStages, useAddStageCapability } from './useValueStreamStages';
import type { ValueStreamStage, ValueStreamDetail } from '../../../api/types';

interface StageFormData {
  name: string;
  description: string;
}

const EMPTY_FORM: StageFormData = { name: '', description: '' };

function useStageForm() {
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingStage, setEditingStage] = useState<ValueStreamStage | null>(null);
  const [formData, setFormData] = useState<StageFormData>(EMPTY_FORM);
  const [insertPosition, setInsertPosition] = useState<number | undefined>(undefined);

  const isFormOpen = showAddForm || editingStage !== null;

  const openAddForm = useCallback((position?: number) => {
    setFormData(EMPTY_FORM);
    setInsertPosition(position);
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
    setInsertPosition(undefined);
  }, []);

  return { isFormOpen, editingStage, formData, setFormData, insertPosition, openAddForm, openEditForm, closeForm };
}

export function useStageOperations(detail: ValueStreamDetail | undefined) {
  const addStageMutation = useAddStage();
  const updateStageMutation = useUpdateStage();
  const deleteStageMutation = useDeleteStage();
  const reorderStagesMutation = useReorderStages();
  const addCapabilityMutation = useAddStageCapability();

  const form = useStageForm();

  const submitForm = useCallback(async () => {
    if (!form.formData.name.trim() || !detail) return;
    const desc = form.formData.description || undefined;
    if (form.editingStage) {
      await updateStageMutation.mutateAsync({ stage: form.editingStage, request: { name: form.formData.name, description: desc } });
    } else {
      await addStageMutation.mutateAsync({ valueStream: detail, request: { name: form.formData.name, description: desc, position: form.insertPosition } });
    }
    form.closeForm();
  }, [form, detail, updateStageMutation, addStageMutation]);

  const deleteStage = useCallback(async (stage: ValueStreamStage) => {
    await deleteStageMutation.mutateAsync(stage);
  }, [deleteStageMutation]);

  const reorderStages = useCallback(async (orderedStageIds: string[]) => {
    if (!detail) return;
    const positions = orderedStageIds.map((stageId, index) => ({ stageId, position: index + 1 }));
    await reorderStagesMutation.mutateAsync({ valueStream: detail, request: { positions } });
  }, [detail, reorderStagesMutation]);

  const addCapability = useCallback((stageId: string, capabilityId: string) => {
    const stage = detail?.stages.find(s => s.id === stageId);
    if (!stage) return;
    addCapabilityMutation.mutate({ stage, capabilityId });
  }, [detail, addCapabilityMutation]);

  return {
    ...form,
    submitForm,
    deleteStage,
    reorderStages,
    addCapability,
  };
}
