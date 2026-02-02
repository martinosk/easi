import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { toComponentId, toCapabilityId, toRelationId } from '../../../api/types';
import type { ViewId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useUpdateOriginEntityPosition } from '../../views/hooks/useViews';
import { canEdit } from '../../../utils/hateoas';
import { extractOriginEntityId } from '../utils/nodeFactory';

function isMultiSelectModifier(event: React.MouseEvent): boolean {
  return event.shiftKey || event.ctrlKey || event.metaKey;
}

function persistOriginEntityPosition(
  node: Node,
  viewId: ViewId,
  mutate: (params: { viewId: ViewId; originEntityId: string; position: { x: number; y: number } }) => void
): void {
  const originEntityId = extractOriginEntityId(node.id);
  if (!originEntityId) return;
  mutate({
    viewId,
    originEntityId: node.id,
    position: { x: node.position.x, y: node.position.y },
  });
}

export const useCanvasSelection = () => {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const { updateComponentPosition, updateCapabilityPosition } = useCanvasLayoutContext();
  const { currentView, currentViewId } = useCurrentView();
  const updateOriginEntityPositionMutation = useUpdateOriginEntityPosition();

  const onNodeClick = useCallback(
    (event: React.MouseEvent, node: Node) => {
      if (isMultiSelectModifier(event)) return;
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
      if (!canEdit(currentView) || !currentViewId) return;
      if (node.type === 'capability') {
        const capId = toCapabilityId(node.id.replace('cap-', ''));
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else if (node.type === 'originEntity') {
        persistOriginEntityPosition(node, currentViewId, updateOriginEntityPositionMutation.mutate);
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
