import { useEffect, useCallback } from 'react';
import { useAppStore } from '../store/appStore';
import { useCurrentView } from '../features/views/hooks/useCurrentView';
import type { ViewComponent } from '../api/types';

export interface KeyboardShortcutHandlers {
  onDelete?: () => void;
}

const isDeleteKey = (event: KeyboardEvent): boolean => event.key === 'Delete';

export function useKeyboardShortcuts(handlers: KeyboardShortcutHandlers) {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { currentView } = useCurrentView();

  const isSelectedNodeInView = useCallback((): boolean => {
    if (!selectedNodeId || !currentView) return false;
    return currentView.components.some((vc: ViewComponent) => vc.componentId === selectedNodeId);
  }, [selectedNodeId, currentView]);

  const handleDeleteKey = useCallback(() => {
    if (!handlers.onDelete) return;
    if (isSelectedNodeInView()) {
      handlers.onDelete();
    }
  }, [handlers, isSelectedNodeInView]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (isDeleteKey(event)) {
        handleDeleteKey();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleDeleteKey]);
}
