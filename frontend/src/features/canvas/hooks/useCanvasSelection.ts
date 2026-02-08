import { useCallback, useMemo } from 'react';
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
    originEntityId,
    position: { x: node.position.x, y: node.position.y },
  });
}

function getNodesToPersist(node: Node, selectedNodes: Node[]): Node[] {
  return selectedNodes.length > 0 ? selectedNodes : [node];
}

interface PositionPersisters {
  updateCapabilityPosition: (id: CapabilityId, x: number, y: number) => Promise<void>;
  updateComponentPosition: (id: ComponentId, x: number, y: number) => Promise<void>;
  updateOriginEntity: (node: Node) => void;
}

function updateSingleNode(node: Node, persisters: PositionPersisters): void {
  if (node.type === 'capability') {
    const capId = toCapabilityId(node.id.replace('cap-', ''));
    persisters.updateCapabilityPosition(capId, node.position.x, node.position.y);
    return;
  }
  if (node.type === 'originEntity') {
    persisters.updateOriginEntity(node);
    return;
  }
  persisters.updateComponentPosition(toComponentId(node.id), node.position.x, node.position.y);
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

  const persisters = useMemo<PositionPersisters>(() => ({
    updateCapabilityPosition,
    updateComponentPosition,
    updateOriginEntity: (target: Node) =>
      currentViewId && persistOriginEntityPosition(target, currentViewId, updateOriginEntityPositionMutation.mutate),
  }), [updateCapabilityPosition, updateComponentPosition, currentViewId, updateOriginEntityPositionMutation]);

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (!canEdit(currentView) || !currentViewId) return;
      const selectedNodes = reactFlowInstance.getNodes().filter((n) => n.selected);
      const nodesToPersist = getNodesToPersist(node, selectedNodes);

      if (nodesToPersist.length === 1) {
        updateSingleNode(nodesToPersist[0], persisters);
        return;
      }

      const updates = buildBatchUpdates(nodesToPersist, persisters.updateOriginEntity);
      if (updates.length > 0) batchUpdatePositions(updates);
    },
    [persisters, currentView, currentViewId, reactFlowInstance, batchUpdatePositions]
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
