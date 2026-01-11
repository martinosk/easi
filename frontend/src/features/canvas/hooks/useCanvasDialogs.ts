import { useCallback, useRef } from 'react';
import { useDialogContext } from '../../../contexts/dialogs';
import { useAppStore } from '../../../store/appStore';
import type { Capability, Component, Relation, ComponentId } from '../../../api/types';

export interface CanvasDialogActions {
  openComponentDialog: () => void;
  openCapabilityDialog: () => void;
  openRelationDialog: (sourceId: string, targetId: string) => void;
  openEditRelationDialog: () => void;
  openEditComponentDialog: (componentId?: string) => void;
  openEditCapabilityDialog: (capability: Capability) => void;
  openReleaseNotesBrowser: () => void;
}

export function useCanvasDialogs(
  selectedEdgeId: string | null,
  relations: Relation[],
  components: Component[]
): CanvasDialogActions {
  const { openDialog } = useDialogContext();
  const selectNode = useAppStore((state) => state.selectNode);

  const componentsRef = useRef(components);
  componentsRef.current = components;

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
    [openDialog]
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
        selectNode(componentId as ComponentId);
        const component = componentsRef.current.find((c: Component) => c.id === componentId);
        if (component) {
          openDialog('edit-component', { component });
        }
      }
    },
    [openDialog, selectNode]
  );

  const openEditCapabilityDialog = useCallback(
    (capability: Capability) => {
      openDialog('edit-capability', { capability });
    },
    [openDialog]
  );

  const openReleaseNotesBrowser = useCallback(() => {
    openDialog('release-notes-browser');
  }, [openDialog]);

  return {
    openComponentDialog,
    openCapabilityDialog,
    openRelationDialog,
    openEditRelationDialog,
    openEditComponentDialog,
    openEditCapabilityDialog,
    openReleaseNotesBrowser,
  };
}
