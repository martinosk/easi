import { useCallback, useMemo } from 'react';
import type { Node } from '@xyflow/react';
import { useNodeContextMenu, type NodeContextMenu, type NodeContextMenuDependencies } from './useNodeContextMenu';
import { useEdgeContextMenu, type EdgeContextMenu } from './useEdgeContextMenu';
import { useMultiSelectContextMenu, type MultiSelectMenuState } from './useMultiSelectContextMenu';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useComponents } from '../../components/hooks/useComponents';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';

export type { NodeContextMenu, EdgeContextMenu, MultiSelectMenuState };

function useNodeMenuDependencies(): NodeContextMenuDependencies {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: acquiredEntities = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: internalTeams = [] } = useInternalTeamsQuery();
  const { currentView } = useCurrentView();

  return useMemo<NodeContextMenuDependencies>(() => ({
    components,
    capabilities,
    originEntityLookups: { acquiredEntities, vendors, internalTeams },
    currentViewComponents: currentView?.components ?? [],
    currentViewCapabilities: currentView?.capabilities ?? [],
    currentViewOriginEntities: currentView?.originEntities ?? [],
  }), [components, capabilities, acquiredEntities, vendors, internalTeams, currentView?.components, currentView?.capabilities, currentView?.originEntities]);
}

function isMultiSelectRightClick(internalNodes: Node[], clickedNode: Node): Node[] | null {
  const selectedNodes = internalNodes.filter((n) => n.selected);
  const clickedNodeIsSelected = selectedNodes.some((n) => n.id === clickedNode.id);
  return selectedNodes.length >= 2 && clickedNodeIsSelected ? selectedNodes : null;
}

export const useContextMenu = (internalNodes: Node[]) => {
  const deps = useNodeMenuDependencies();

  const {
    nodeContextMenu,
    onNodeContextMenu: openSingleNodeMenu,
    closeNodeMenu,
  } = useNodeContextMenu();

  const {
    edgeContextMenu,
    onEdgeContextMenu,
    closeEdgeMenu,
  } = useEdgeContextMenu();

  const {
    multiSelectMenu,
    openMultiSelectMenu,
    closeMultiSelectMenu,
  } = useMultiSelectContextMenu(deps);

  const onSelectionContextMenu = useCallback(
    (event: React.MouseEvent, nodes: Node[]) => {
      event.preventDefault();
      openMultiSelectMenu({ x: event.clientX, y: event.clientY }, nodes);
    },
    [openMultiSelectMenu]
  );

  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      const selectedNodes = isMultiSelectRightClick(internalNodes, node);
      if (selectedNodes) {
        event.preventDefault();
        openMultiSelectMenu({ x: event.clientX, y: event.clientY }, selectedNodes);
      } else {
        openSingleNodeMenu(event, node);
      }
    },
    [internalNodes, openMultiSelectMenu, openSingleNodeMenu]
  );

  const closeMenus = useCallback(() => {
    closeNodeMenu();
    closeEdgeMenu();
    closeMultiSelectMenu();
  }, [closeNodeMenu, closeEdgeMenu, closeMultiSelectMenu]);

  return {
    nodeContextMenu,
    edgeContextMenu,
    multiSelectMenu,
    onNodeContextMenu,
    onSelectionContextMenu,
    onEdgeContextMenu,
    closeMenus,
  };
};
