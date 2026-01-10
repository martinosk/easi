import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import type { ComponentId, CapabilityId, RelationId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { canEdit } from '../../../utils/hateoas';

export const useCanvasSelection = () => {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const { updateComponentPosition, updateCapabilityPosition } = useCanvasLayoutContext();
  const { currentView } = useCurrentView();

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
      if (!canEdit(currentView)) {
        return;
      }
      if (node.type === 'capability') {
        const capId = node.id.replace('cap-', '') as CapabilityId;
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else {
        updateComponentPosition(node.id as ComponentId, node.position.x, node.position.y);
      }
    },
    [updateComponentPosition, updateCapabilityPosition, currentView]
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
