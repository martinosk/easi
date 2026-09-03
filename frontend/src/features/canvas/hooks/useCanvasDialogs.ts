import { useCallback } from 'react';
import type { Relation } from '../../../api/types';
import { useDialogContext } from '../../../contexts/dialogs';

export interface CanvasDialogActions {
  openComponentDialog: () => void;
  openCapabilityDialog: () => void;
  openRelationDialog: (sourceId: string, targetId: string) => void;
  openEditRelationDialog: () => void;
  openReleaseNotesBrowser: () => void;
}

export function useCanvasDialogs(selectedEdgeId: string | null, relations: Relation[]): CanvasDialogActions {
  const { openDialog } = useDialogContext();

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

  const openReleaseNotesBrowser = useCallback(() => {
    openDialog('release-notes-browser');
  }, [openDialog]);

  return {
    openComponentDialog,
    openCapabilityDialog,
    openRelationDialog,
    openEditRelationDialog,
    openReleaseNotesBrowser,
  };
}
