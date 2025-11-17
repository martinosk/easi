import { useEffect } from 'react';
import { useAppStore } from '../store/appStore';

export interface KeyboardShortcutHandlers {
  onDelete?: () => void;
}

export function useKeyboardShortcuts(handlers: KeyboardShortcutHandlers) {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const currentView = useAppStore((state) => state.currentView);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Delete' && selectedNodeId && currentView && handlers.onDelete) {
        const isInCurrentView = currentView.components.some(
          (vc) => vc.componentId === selectedNodeId
        );
        if (isInCurrentView) {
          handlers.onDelete();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedNodeId, currentView, handlers]);
}
