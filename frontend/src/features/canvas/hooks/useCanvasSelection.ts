import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';

export const useCanvasSelection = () => {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const updatePosition = useAppStore((state) => state.updatePosition);
  const updateCapabilityPosition = useAppStore((state) => state.updateCapabilityPosition);

  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (node.type === 'capability') {
        const capId = node.id.replace('cap-', '');
        selectCapability(capId);
        selectNode(null);
      } else {
        selectNode(node.id);
        selectCapability(null);
      }
    },
    [selectNode, selectCapability]
  );

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      selectEdge(edge.id);
    },
    [selectEdge]
  );

  const onPaneClick = useCallback(() => {
    clearSelection();
    selectCapability(null);
  }, [clearSelection, selectCapability]);

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (node.type === 'capability') {
        const capId = node.id.replace('cap-', '');
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else {
        updatePosition(node.id, node.position);
      }
    },
    [updatePosition, updateCapabilityPosition]
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
