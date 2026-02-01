import { useState, useCallback } from 'react';
import type { Node } from '@xyflow/react';
import { hasLink } from '../../../utils/hateoas';
import {
  resolveNodeMenu,
  type MenuPosition,
  type NodeContextMenuDependencies,
  type NodeContextMenu,
} from './useNodeContextMenu';

export interface MultiSelectAction {
  type: 'removeFromView' | 'deleteFromModel';
  label: string;
  isDanger: boolean;
}

export interface MultiSelectMenuState {
  x: number;
  y: number;
  selectedNodes: NodeContextMenu[];
  actions: MultiSelectAction[];
}

export function computeAvailableActions(
  resolvedNodes: NodeContextMenu[]
): MultiSelectAction[] {
  if (resolvedNodes.length < 2) return [];

  const actions: MultiSelectAction[] = [];
  const count = resolvedNodes.length;

  const allCanRemoveFromView = resolvedNodes.every((n) =>
    hasLink({ _links: n.viewElementLinks }, 'x-remove')
  );

  const allCanDeleteFromModel = resolvedNodes.every((n) =>
    hasLink({ _links: n.modelLinks }, 'delete')
  );

  if (allCanRemoveFromView) {
    actions.push({
      type: 'removeFromView',
      label: `Remove from View (${count} items)`,
      isDanger: false,
    });
  }

  if (allCanDeleteFromModel) {
    actions.push({
      type: 'deleteFromModel',
      label: `Delete from Model (${count} items)`,
      isDanger: true,
    });
  }

  return actions;
}

export const useMultiSelectContextMenu = (deps: NodeContextMenuDependencies) => {
  const [multiSelectMenu, setMultiSelectMenu] = useState<MultiSelectMenuState | null>(null);

  const openMultiSelectMenu = useCallback(
    (position: MenuPosition, selectedNodes: Node[]) => {
      const resolved = selectedNodes
        .map((node) => resolveNodeMenu(node, position, deps))
        .filter((menu): menu is NodeContextMenu => menu !== null);

      const actions = computeAvailableActions(resolved);
      if (actions.length === 0) return;

      setMultiSelectMenu({
        ...position,
        selectedNodes: resolved,
        actions,
      });
    },
    [deps]
  );

  const closeMultiSelectMenu = useCallback(() => {
    setMultiSelectMenu(null);
  }, []);

  return {
    multiSelectMenu,
    openMultiSelectMenu,
    closeMultiSelectMenu,
  };
};
