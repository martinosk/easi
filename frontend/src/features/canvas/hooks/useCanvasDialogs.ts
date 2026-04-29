import { useCallback, useEffect, useRef } from 'react';
import type { Capability, Component, Relation } from '../../../api/types';
import { toComponentId } from '../../../api/types';
import type { CreateConnectedEntityDialogData } from '../../../contexts/dialogs/types';
import { useDialogContext } from '../../../contexts/dialogs';
import { useAppStore } from '../../../store/appStore';

export interface CanvasDialogActions {
  openComponentDialog: () => void;
  openCapabilityDialog: () => void;
  openRelationDialog: (sourceId: string, targetId: string) => void;
  openEditRelationDialog: () => void;
  openEditComponentDialog: (componentId?: string) => void;
  openEditCapabilityDialog: (capability: Capability) => void;
  openReleaseNotesBrowser: () => void;
  openCreateConnectedEntityDialog: (params: CreateConnectedEntityDialogData) => void;
  closeCreateConnectedEntityDialog: () => void;
}

export function useCanvasDialogs(
  selectedEdgeId: string | null,
  relations: Relation[],
  components: Component[],
): CanvasDialogActions {
  const { openDialog, closeDialog } = useDialogContext();
  const selectNode = useAppStore((state) => state.selectNode);

  const componentsRef = useRef(components);
  useEffect(() => {
    componentsRef.current = components;
  });

  const openComponentDialog = useCallback(() => {
    openDialog('create-component');
  }, [openDialog]);

  const openCapabilityDialog = useCallback(() => {
    openDialog('create-capability');
  }, [openDialog]);

  const openRelationDialog = useCallback(
    (sourceId: string, targetId: string) => {
      openDialog('create-relation', { sourceComponentId: sourceId, targetComponentId: targetId });
    },
    [openDialog],
  );

  const openEditRelationDialog = useCallback(() => {
    const selectedRelation = relations.find((r) => r.id === selectedEdgeId) || null;
    if (selectedRelation) {
      openDialog('edit-relation', { relation: selectedRelation });
    }
  }, [openDialog, relations, selectedEdgeId]);

  const openEditComponentDialog = useCallback(
    (componentId?: string) => {
      if (componentId) {
        selectNode(toComponentId(componentId));
        const component = componentsRef.current.find((c: Component) => c.id === componentId);
        if (component) {
          openDialog('edit-component', { component });
        }
      }
    },
    [openDialog, selectNode],
  );

  const openEditCapabilityDialog = useCallback(
    (capability: Capability) => {
      openDialog('edit-capability', { capability });
    },
    [openDialog],
  );

  const openReleaseNotesBrowser = useCallback(() => {
    openDialog('release-notes-browser');
  }, [openDialog]);

  const openCreateConnectedEntityDialog = useCallback(
    (params: CreateConnectedEntityDialogData) => {
      openDialog('create-connected-entity', params);
    },
    [openDialog],
  );

  const closeCreateConnectedEntityDialog = useCallback(() => {
    closeDialog('create-connected-entity');
  }, [closeDialog]);

  return {
    openComponentDialog,
    openCapabilityDialog,
    openRelationDialog,
    openEditRelationDialog,
    openEditComponentDialog,
    openEditCapabilityDialog,
    openReleaseNotesBrowser,
    openCreateConnectedEntityDialog,
    closeCreateConnectedEntityDialog,
  };
}
