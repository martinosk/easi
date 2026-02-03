import { useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useReactFlow } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import {
  toComponentId,
  toCapabilityId,
  toRelationId,
  type BatchUpdateItem,
  type CapabilityId,
  type ComponentId,
} from '../../../api/types';
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

function getNodesToPersist(node: Node, selectedNodes: Node[]): Node[] {
  return selectedNodes.length > 0 ? selectedNodes : [node];
}

function updateSingleNode(
  node: Node,
  currentViewId: ViewId,
  updateCapabilityPosition: (id: CapabilityId, x: number, y: number) => Promise<void>,
  updateComponentPosition: (id: ComponentId, x: number, y: number) => Promise<void>,
  updateOriginEntity: (node: Node, viewId: ViewId) => void
): void {
  if (node.type === 'capability') {
    const capId = toCapabilityId(node.id.replace('cap-', ''));
    updateCapabilityPosition(capId, node.position.x, node.position.y);
    return;
  }
  if (node.type === 'originEntity') {
    updateOriginEntity(node, currentViewId);
    return;
  }
  updateComponentPosition(toComponentId(node.id), node.position.x, node.position.y);
}

function buildBatchUpdates(nodes: Node[], updateOriginEntity: (node: Node) => void): BatchUpdateItem[] {
  const updates: BatchUpdateItem[] = [];
  for (const target of nodes) {
    if (target.type === 'originEntity') {
      updateOriginEntity(target);
      continue;
    }
    if (target.type === 'capability') {
      updates.push({ elementId: target.id.replace('cap-', ''), x: target.position.x, y: target.position.y });
      continue;
    }
    updates.push({ elementId: target.id, x: target.position.x, y: target.position.y });
  }
  return updates;
}

export const useCanvasSelection = () => {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const { updateComponentPosition, updateCapabilityPosition, batchUpdatePositions } = useCanvasLayoutContext();
  const { currentView, currentViewId } = useCurrentView();
  const updateOriginEntityPositionMutation = useUpdateOriginEntityPosition();
  const reactFlowInstance = useReactFlow();

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
      const selectedNodes = reactFlowInstance.getNodes().filter((n) => n.selected);
      const nodesToPersist = getNodesToPersist(node, selectedNodes);
      const updateOriginEntity = (target: Node) =>
        persistOriginEntityPosition(target, currentViewId, updateOriginEntityPositionMutation.mutate);

      if (nodesToPersist.length === 1) {
        updateSingleNode(
          nodesToPersist[0],
          currentViewId,
          updateCapabilityPosition,
          updateComponentPosition,
          updateOriginEntity
        );
        return;
      }

      const updates = buildBatchUpdates(nodesToPersist, updateOriginEntity);
      if (updates.length > 0) batchUpdatePositions(updates);
    },
    [
      updateComponentPosition,
      updateCapabilityPosition,
      updateOriginEntityPositionMutation,
      currentView,
      currentViewId,
      reactFlowInstance,
      batchUpdatePositions,
    ]
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
