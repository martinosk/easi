import { useCallback } from 'react';
import { useDialogContext } from '../../../contexts/dialogs';

export interface CanvasDialogActions {
  openComponentDialog: () => void;
  openCapabilityDialog: () => void;
  openRelationDialog: (sourceId: string, targetId: string) => void;
  openReleaseNotesBrowser: () => void;
}

export function useCanvasDialogs(): CanvasDialogActions {
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

  const openReleaseNotesBrowser = useCallback(() => {
    openDialog('release-notes-browser');
  }, [openDialog]);

  return {
    openComponentDialog,
    openCapabilityDialog,
    openRelationDialog,
    openReleaseNotesBrowser,
  };
}
