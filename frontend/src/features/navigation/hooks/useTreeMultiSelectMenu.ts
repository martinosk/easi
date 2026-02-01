import { useState, useCallback } from 'react';
import type { TreeSelectedItem } from './useTreeMultiSelect';

export interface TreeBulkAction {
  type: 'deleteFromModel';
  label: string;
  isDanger: boolean;
}

export interface TreeMultiSelectMenuState {
  x: number;
  y: number;
  items: TreeSelectedItem[];
  actions: TreeBulkAction[];
}

export function computeTreeBulkActions(items: TreeSelectedItem[]): TreeBulkAction[] {
  if (items.length < 2) return [];

  const allCanDelete = items.every((item) => item.links?.delete !== undefined);

  if (!allCanDelete) return [];

  return [
    {
      type: 'deleteFromModel',
      label: `Delete from Model (${items.length} items)`,
      isDanger: true,
    },
  ];
}

export function useTreeMultiSelectMenu() {
  const [menu, setMenu] = useState<TreeMultiSelectMenuState | null>(null);

  const handleMultiSelectContextMenu = useCallback(
    (
      event: React.MouseEvent,
      clickedItemId: string,
      selectedItems: TreeSelectedItem[]
    ): boolean => {
      if (selectedItems.length < 2) return false;

      const isInSelection = selectedItems.some((item) => item.id === clickedItemId);
      if (!isInSelection) return false;

      event.preventDefault();
      event.stopPropagation();

      const actions = computeTreeBulkActions(selectedItems);
      if (actions.length === 0) return false;

      setMenu({
        x: event.clientX,
        y: event.clientY,
        items: selectedItems,
        actions,
      });

      return true;
    },
    []
  );

  const closeMenu = useCallback(() => {
    setMenu(null);
  }, []);

  return {
    menu,
    handleMultiSelectContextMenu,
    closeMenu,
  };
}
