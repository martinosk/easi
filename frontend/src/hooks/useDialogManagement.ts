import { useState, useCallback } from 'react';
import { useDialogState } from './useDialogState';
import { useRelationDialog } from './useRelationDialog';
import { useAppStore } from '../store/appStore';
import type { Capability, Component, Relation } from '../api/types';

export interface DialogManagementState {
  componentDialog: {
    isOpen: boolean;
    onClose: () => void;
  };
  editComponentDialog: {
    isOpen: boolean;
    onClose: () => void;
    component: Component | null;
  };
  relationDialog: {
    isOpen: boolean;
    onClose: () => void;
    sourceComponentId: string | undefined;
    targetComponentId: string | undefined;
  };
  editRelationDialog: {
    isOpen: boolean;
    onClose: () => void;
    relation: Relation | null;
  };
  capabilityDialog: {
    isOpen: boolean;
    onClose: () => void;
  };
  editCapabilityDialog: {
    isOpen: boolean;
    onClose: () => void;
    capability: Capability | null;
  };
  releaseNotesBrowserDialog: {
    isOpen: boolean;
    onClose: () => void;
    onOpen: () => void;
  };
}

export interface DialogManagementActions {
  openComponentDialog: () => void;
  openCapabilityDialog: () => void;
  openRelationDialog: (sourceId: string, targetId: string) => void;
  openEditRelationDialog: () => void;
  openEditComponentDialog: (componentId?: string) => void;
  openEditCapabilityDialog: (capability: Capability) => void;
}

export interface UseDialogManagementReturn {
  state: DialogManagementState;
  actions: DialogManagementActions;
}

export function useDialogManagement(
  selectedEdgeId: string | null,
  relations: Relation[]
): UseDialogManagementReturn {
  const componentDialog = useDialogState();
  const editComponentDialog = useDialogState();
  const relationDialog = useRelationDialog();
  const editRelationDialog = useDialogState();
  const capabilityDialog = useDialogState();
  const editCapabilityDialogState = useDialogState();
  const releaseNotesBrowserDialog = useDialogState();

  const [editCapabilityTarget, setEditCapabilityTarget] = useState<Capability | null>(null);
  const [editComponentTarget, setEditComponentTarget] = useState<Component | null>(null);

  const selectNode = useAppStore((state) => state.selectNode);
  const getComponents = useAppStore.getState;

  const openEditCapabilityDialog = useCallback((capability: Capability) => {
    setEditCapabilityTarget(capability);
    editCapabilityDialogState.open();
  }, [editCapabilityDialogState]);

  const closeEditCapabilityDialog = useCallback(() => {
    editCapabilityDialogState.close();
    setEditCapabilityTarget(null);
  }, [editCapabilityDialogState]);

  const openEditComponentDialog = useCallback((componentId?: string) => {
    if (componentId) {
      selectNode(componentId);
      const component = getComponents().components.find((c) => c.id === componentId);
      setEditComponentTarget(component || null);
    }
    editComponentDialog.open();
  }, [selectNode, editComponentDialog, getComponents]);

  const closeEditComponentDialog = useCallback(() => {
    editComponentDialog.close();
    setEditComponentTarget(null);
  }, [editComponentDialog]);

  const selectedRelation = relations.find((r) => r.id === selectedEdgeId) || null;

  return {
    state: {
      componentDialog: {
        isOpen: componentDialog.isOpen,
        onClose: componentDialog.close,
      },
      editComponentDialog: {
        isOpen: editComponentDialog.isOpen,
        onClose: closeEditComponentDialog,
        component: editComponentTarget,
      },
      relationDialog: {
        isOpen: relationDialog.isOpen,
        onClose: relationDialog.close,
        sourceComponentId: relationDialog.sourceId,
        targetComponentId: relationDialog.targetId,
      },
      editRelationDialog: {
        isOpen: editRelationDialog.isOpen,
        onClose: editRelationDialog.close,
        relation: selectedRelation,
      },
      capabilityDialog: {
        isOpen: capabilityDialog.isOpen,
        onClose: capabilityDialog.close,
      },
      editCapabilityDialog: {
        isOpen: editCapabilityDialogState.isOpen,
        onClose: closeEditCapabilityDialog,
        capability: editCapabilityTarget,
      },
      releaseNotesBrowserDialog: {
        isOpen: releaseNotesBrowserDialog.isOpen,
        onClose: releaseNotesBrowserDialog.close,
        onOpen: releaseNotesBrowserDialog.open,
      },
    },
    actions: {
      openComponentDialog: componentDialog.open,
      openCapabilityDialog: capabilityDialog.open,
      openRelationDialog: relationDialog.open,
      openEditRelationDialog: editRelationDialog.open,
      openEditComponentDialog,
      openEditCapabilityDialog,
    },
  };
}
