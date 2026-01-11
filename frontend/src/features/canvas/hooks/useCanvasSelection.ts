import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { toComponentId, toCapabilityId, toRelationId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../views/hooks/useCurrentView';
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
        const capId = toCapabilityId(node.id.replace('cap-', ''));
        selectCapability(capId);
        selectNode(null);
      } else {
        selectNode(toComponentId(node.id));
        selectCapability(null);
      }
    },
    [selectNode, selectCapability]
  );

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      selectEdge(toRelationId(edge.id));
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
        const capId = toCapabilityId(node.id.replace('cap-', ''));
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else {
        updateComponentPosition(toComponentId(node.id), node.position.x, node.position.y);
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
