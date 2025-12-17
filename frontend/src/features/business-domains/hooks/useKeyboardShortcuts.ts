import { useEffect } from 'react';

interface UseKeyboardShortcutsProps {
  hasSelection: boolean;
  onSelectAll: () => void;
  onClearSelection: () => void;
}

export function useKeyboardShortcuts({
  hasSelection,
  onSelectAll,
  onClearSelection,
}: UseKeyboardShortcutsProps): void {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const isSelectAllShortcut = (e.ctrlKey || e.metaKey) && e.key === 'a';
      if (isSelectAllShortcut) {
        e.preventDefault();
        onSelectAll();
        return;
      }

      if (e.key === 'Escape' && hasSelection) {
        e.preventDefault();
        onClearSelection();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [hasSelection, onSelectAll, onClearSelection]);
}
