import React, { useState, useCallback, useMemo } from 'react';
import { DialogContext } from './context';
import type { DialogId, DialogDataMap, DialogContextValue } from './types';

export function DialogProvider({ children }: { children: React.ReactNode }) {
  const [openDialogs, setOpenDialogs] = useState<Map<DialogId, DialogDataMap[DialogId]>>(
    () => new Map()
  );

  const openDialog = useCallback(<T extends DialogId>(id: T, data?: DialogDataMap[T]) => {
    setOpenDialogs((prev) => {
      const next = new Map(prev);
      next.set(id, data as DialogDataMap[DialogId]);
      return next;
    });
  }, []);

  const closeDialog = useCallback((id: DialogId) => {
    setOpenDialogs((prev) => {
      const next = new Map(prev);
      next.delete(id);
      return next;
    });
  }, []);

  const isOpen = useCallback(
    (id: DialogId) => openDialogs.has(id),
    [openDialogs]
  );

  const getData = useCallback(
    <T extends DialogId>(id: T): DialogDataMap[T] | undefined => {
      return openDialogs.get(id) as DialogDataMap[T] | undefined;
    },
    [openDialogs]
  );

  const value = useMemo<DialogContextValue>(
    () => ({
      openDialogs,
      openDialog,
      closeDialog,
      isOpen,
      getData,
    }),
    [openDialogs, openDialog, closeDialog, isOpen, getData]
  );

  return <DialogContext.Provider value={value}>{children}</DialogContext.Provider>;
}
