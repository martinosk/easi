import { useState, useCallback, useMemo } from 'react';
import type { HATEOASLinks } from '../../../api/types';

export type TreeItemType = 'component' | 'capability' | 'acquired' | 'vendor' | 'team';

export interface TreeSelectedItem {
  id: string;
  name: string;
  type: TreeItemType;
  links: HATEOASLinks | undefined;
}

interface Anchor {
  id: string;
  sectionId: string;
  item: TreeSelectedItem;
}

export interface MultiDragPayload {
  items: {
    type: TreeItemType;
    id: string;
    name: string;
  }[];
}

function isModifierClick(event: React.MouseEvent): boolean {
  return event.ctrlKey || event.metaKey;
}

function isShiftClick(event: React.MouseEvent): boolean {
  return event.shiftKey;
}

function selectRange(
  visibleItems: TreeSelectedItem[],
  fromId: string,
  toId: string
): Map<string, TreeSelectedItem> {
  const result = new Map<string, TreeSelectedItem>();
  const fromIndex = visibleItems.findIndex((item) => item.id === fromId);
  const toIndex = visibleItems.findIndex((item) => item.id === toId);

  if (fromIndex === -1 || toIndex === -1) {
    const toItem = visibleItems.find((item) => item.id === toId);
    if (toItem) result.set(toItem.id, toItem);
    return result;
  }

  const start = Math.min(fromIndex, toIndex);
  const end = Math.max(fromIndex, toIndex);

  for (let i = start; i <= end; i++) {
    const item = visibleItems[i];
    result.set(item.id, item);
  }

  return result;
}

function setDragImage(event: React.DragEvent, count: number): void {
  const dragLabel = document.createElement('div');
  dragLabel.textContent = `${count} items`;
  dragLabel.style.cssText = 'position:absolute;top:-1000px;padding:4px 8px;background:#4a90d9;color:#fff;border-radius:4px;font-size:12px;';
  document.body.appendChild(dragLabel);
  event.dataTransfer.setDragImage(dragLabel, 0, 0);
  requestAnimationFrame(() => document.body.removeChild(dragLabel));
}

function toggleItem(prev: Map<string, TreeSelectedItem>, item: TreeSelectedItem, anchor: Anchor | null): Map<string, TreeSelectedItem> {
  const next = new Map(prev);
  if (next.size === 0 && anchor && anchor.id !== item.id) {
    next.set(anchor.item.id, anchor.item);
  }
  if (next.has(item.id)) {
    next.delete(item.id);
  } else {
    next.set(item.id, item);
  }
  return next;
}

interface ShiftSelectionParams {
  prev: Map<string, TreeSelectedItem>;
  visibleItems: TreeSelectedItem[];
  anchor: Anchor | null;
  sectionId: string;
}

function mergeShiftSelection(params: ShiftSelectionParams, item: TreeSelectedItem): Map<string, TreeSelectedItem> {
  const { prev, visibleItems, anchor, sectionId } = params;
  const visibleIds = new Set(visibleItems.map((v) => v.id));
  const otherSections = new Map<string, TreeSelectedItem>();
  prev.forEach((val, key) => {
    if (!visibleIds.has(key)) {
      otherSections.set(key, val);
    }
  });

  const fromId = anchor?.sectionId === sectionId ? anchor.id : visibleItems[0].id;
  const sectionSelection = selectRange(visibleItems, fromId, item.id);

  const merged = new Map(otherSections);
  sectionSelection.forEach((val, key) => merged.set(key, val));
  return merged;
}

export function useTreeMultiSelect() {
  const [selectedItems, setSelectedItems] = useState<Map<string, TreeSelectedItem>>(new Map());
  const [anchor, setAnchor] = useState<Anchor | null>(null);

  const handleItemClick = useCallback(
    (
      item: TreeSelectedItem,
      sectionId: string,
      visibleItems: TreeSelectedItem[],
      event: React.MouseEvent
    ): 'multi' | 'single' => {
      if (isModifierClick(event)) {
        setSelectedItems((prev) => toggleItem(prev, item, anchor));
        setAnchor({ id: item.id, sectionId, item });
        return 'multi';
      }

      if (isShiftClick(event)) {
        setSelectedItems((prev) => mergeShiftSelection({ prev, visibleItems, anchor, sectionId }, item));
        return 'multi';
      }

      setSelectedItems(new Map());
      setAnchor({ id: item.id, sectionId, item });
      return 'single';
    },
    [anchor]
  );

  const isMultiSelected = useCallback(
    (id: string): boolean => selectedItems.has(id),
    [selectedItems]
  );

  const clearMultiSelection = useCallback(() => {
    setSelectedItems(new Map());
    setAnchor(null);
  }, []);

  const selectionCount = selectedItems.size;

  const getSelectedItems = useCallback(
    (): TreeSelectedItem[] => Array.from(selectedItems.values()),
    [selectedItems]
  );

  const buildMultiDragPayload = useCallback((): string => {
    const items = Array.from(selectedItems.values()).map((item) => ({
      type: item.type,
      id: item.id,
      name: item.name,
    }));
    return JSON.stringify({ items } satisfies MultiDragPayload);
  }, [selectedItems]);

  const handleDragStart = useCallback(
    (event: React.DragEvent, itemId: string): boolean => {
      if (selectedItems.size < 2 || !selectedItems.has(itemId)) return false;

      event.dataTransfer.setData('multiDragItems', buildMultiDragPayload());
      event.dataTransfer.effectAllowed = 'copy';
      setDragImage(event, selectedItems.size);

      return true;
    },
    [selectedItems, buildMultiDragPayload]
  );

  const selectedItemsList = useMemo(() => Array.from(selectedItems.values()), [selectedItems]);

  return {
    selectedItems: selectedItemsList,
    handleItemClick,
    isMultiSelected,
    clearMultiSelection,
    selectionCount,
    getSelectedItems,
    buildMultiDragPayload,
    handleDragStart,
  };
}

export type UseTreeMultiSelect = ReturnType<typeof useTreeMultiSelect>;
