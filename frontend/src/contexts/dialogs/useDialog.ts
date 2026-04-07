import { useCallback, useMemo } from 'react';
import type { DialogDataMap, DialogId } from './types';
import { useDialogContext } from './useDialogContext';

export interface UseDialogReturn<T extends DialogId> {
  isOpen: boolean;
  data: DialogDataMap[T] | undefined;
  open: (data?: DialogDataMap[T]) => void;
  close: () => void;
}

export function useDialog<T extends DialogId>(id: T): UseDialogReturn<T> {
  const { isOpen: checkIsOpen, getData, openDialog, closeDialog } = useDialogContext();

  const isOpen = checkIsOpen(id);
  const data = getData(id);

  const open = useCallback(
    (dialogData?: DialogDataMap[T]) => {
      openDialog(id, dialogData);
    },
    [id, openDialog],
  );

  const close = useCallback(() => {
    closeDialog(id);
  }, [id, closeDialog]);

  return useMemo(
    () => ({
      isOpen,
      data,
      open,
      close,
    }),
    [isOpen, data, open, close],
  );
}
