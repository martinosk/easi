import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { toComponentId, toCapabilityId, toRelationId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useUpdateOriginEntityPosition } from '../../views/hooks/useViews';
import { canEdit } from '../../../utils/hateoas';
import { extractOriginEntityId } from '../utils/nodeFactory';

export const useCanvasSelection = () => {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const { updateComponentPosition, updateCapabilityPosition } = useCanvasLayoutContext();
  const { currentView, currentViewId } = useCurrentView();
  const updateOriginEntityPositionMutation = useUpdateOriginEntityPosition();

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
      if (!canEdit(currentView) || !currentViewId) {
        return;
      }
      if (node.type === 'capability') {
        const capId = toCapabilityId(node.id.replace('cap-', ''));
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else if (node.type === 'originEntity') {
        const originEntityId = extractOriginEntityId(node.id);
        if (originEntityId) {
          updateOriginEntityPositionMutation.mutate({
            viewId: currentViewId,
            originEntityId: node.id,
            position: { x: node.position.x, y: node.position.y },
          });
        }
      } else {
        updateComponentPosition(toComponentId(node.id), node.position.x, node.position.y);
      }
    },
    [updateComponentPosition, updateCapabilityPosition, updateOriginEntityPositionMutation, currentView, currentViewId]
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
