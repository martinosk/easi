import { useState, useCallback } from 'react';

export interface RelationDialogState {
  isOpen: boolean;
  sourceId: string | undefined;
  targetId: string | undefined;
  open: (source: string, target: string) => void;
  close: () => void;
}

export function useRelationDialog(): RelationDialogState {
  const [isOpen, setIsOpen] = useState(false);
  const [sourceId, setSourceId] = useState<string | undefined>();
  const [targetId, setTargetId] = useState<string | undefined>();

  const open = useCallback((source: string, target: string) => {
    setSourceId(source);
    setTargetId(target);
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setSourceId(undefined);
    setTargetId(undefined);
  }, []);

  return { isOpen, sourceId, targetId, open, close };
}
