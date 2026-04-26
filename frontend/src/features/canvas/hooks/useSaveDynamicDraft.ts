import { useState } from 'react';
import type { ViewId } from '../../../api/types';
import {
  useAddCapabilityToView,
  useAddComponentToView,
  useAddOriginEntityToView,
  useRemoveCapabilityFromView,
  useRemoveComponentFromView,
  useRemoveOriginEntityFromView,
  useUpdateCapabilityPosition,
  useUpdateComponentPosition,
  useUpdateOriginEntityPosition,
} from '../../views/hooks/useViews';
import {
  saveDraft,
  type DraftSaveApi,
  type DraftSaveInput,
  type DraftSaveResult,
} from '../utils/saveDraft';

type ComponentId = Parameters<DraftSaveApi['addComponent']>[1];
type CapabilityId = Parameters<DraftSaveApi['addCapability']>[1];

export function useSaveDynamicDraft(): {
  save: (input: DraftSaveInput) => Promise<DraftSaveResult>;
  isSaving: boolean;
} {
  const addComponent = useAddComponentToView();
  const addCapability = useAddCapabilityToView();
  const addOriginEntity = useAddOriginEntityToView();
  const removeComponent = useRemoveComponentFromView();
  const removeCapability = useRemoveCapabilityFromView();
  const removeOriginEntity = useRemoveOriginEntityFromView();
  const updateComponentPosition = useUpdateComponentPosition();
  const updateCapabilityPosition = useUpdateCapabilityPosition();
  const updateOriginEntityPosition = useUpdateOriginEntityPosition();

  const [isSaving, setIsSaving] = useState(false);

  const save = async (input: DraftSaveInput): Promise<DraftSaveResult> => {
    const api: DraftSaveApi = {
      addComponent: (viewId, id, x, y) =>
        addComponent.mutateAsync({ viewId, request: { componentId: id as ComponentId, x, y } }),
      addCapability: (viewId, id, x, y) =>
        addCapability.mutateAsync({ viewId, request: { capabilityId: id as CapabilityId, x, y } }),
      addOriginEntity: (viewId, id, x, y) =>
        addOriginEntity.mutateAsync({ viewId, request: { originEntityId: id, x, y } }),
      removeComponent: (viewId, id) => removeComponent.mutateAsync({ viewId, componentId: id as ComponentId }),
      removeCapability: (viewId, id) => removeCapability.mutateAsync({ viewId, capabilityId: id as CapabilityId }),
      removeOriginEntity: (viewId, id) => removeOriginEntity.mutateAsync({ viewId, originEntityId: id }),
      updateComponentPosition: (viewId, id, x, y) =>
        updateComponentPosition.mutateAsync({
          viewId,
          componentId: id as ComponentId,
          request: { x, y },
        }),
      updateCapabilityPosition: (viewId, id, x, y) =>
        updateCapabilityPosition.mutateAsync({
          viewId,
          capabilityId: id as CapabilityId,
          position: { x, y },
        }),
      updateOriginEntityPosition: (viewId, id, x, y) =>
        updateOriginEntityPosition.mutateAsync({ viewId, originEntityId: id, position: { x, y } }),
    };

    setIsSaving(true);
    try {
      return await saveDraft(api, input);
    } finally {
      setIsSaving(false);
    }
  };

  return { save, isSaving };
}

export type { DraftSaveInput, DraftSaveResult, ViewId };
