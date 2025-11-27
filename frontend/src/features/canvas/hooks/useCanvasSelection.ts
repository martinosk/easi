import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import type { ComponentId, CapabilityId, RelationId } from '../../../api/types';

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
        const capId = node.id.replace('cap-', '') as CapabilityId;
        selectCapability(capId);
        selectNode(null);
      } else {
        selectNode(node.id as ComponentId);
        selectCapability(null);
      }
    },
    [selectNode, selectCapability]
  );

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      selectEdge(edge.id as RelationId);
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
        const capId = node.id.replace('cap-', '') as CapabilityId;
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else {
        updatePosition(node.id as ComponentId, node.position);
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
