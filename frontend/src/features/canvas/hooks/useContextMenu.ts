import { useCallback } from 'react';
import { useNodeContextMenu, type NodeContextMenu } from './useNodeContextMenu';
import { useEdgeContextMenu, type EdgeContextMenu } from './useEdgeContextMenu';

export type { NodeContextMenu, EdgeContextMenu };

export const useContextMenu = () => {
  const {
    nodeContextMenu,
    onNodeContextMenu,
    closeNodeMenu,
  } = useNodeContextMenu();

  const {
    edgeContextMenu,
    onEdgeContextMenu,
    closeEdgeMenu,
  } = useEdgeContextMenu();

  const closeMenus = useCallback(() => {
    closeNodeMenu();
    closeEdgeMenu();
  }, [closeNodeMenu, closeEdgeMenu]);

  return {
    nodeContextMenu,
    edgeContextMenu,
    onNodeContextMenu,
    onEdgeContextMenu,
    closeMenus,
  };
};
