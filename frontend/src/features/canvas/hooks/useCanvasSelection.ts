import type { Edge, Node } from '@xyflow/react';
import { useReactFlow } from '@xyflow/react';
import { useCallback, useMemo } from 'react';
import type { ViewId } from '../../../api/types';
import {
  type BatchUpdateItem,
  type CapabilityId,
  type ComponentId,
  toCapabilityId,
  toComponentId,
  toRelationId,
} from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useUpdateOriginEntityPosition } from '../../views/hooks/useViews';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { getEntityId, getEntityType, toNodeId } from '../../../constants/entityIdentifiers';
import type { EntityType } from '../utils/dynamicMode';

const NODE_TYPE_TO_ENTITY_TYPE: Record<string, EntityType> = {
  component: 'component',
  capability: 'capability',
  acquired: 'originEntity',
  vendor: 'originEntity',
  team: 'originEntity',
};

function nodeIdToEntityKey(nodeId: string): { id: string; type: EntityType } {
  const parsed = toNodeId(nodeId);
  return { id: getEntityId(parsed), type: NODE_TYPE_TO_ENTITY_TYPE[getEntityType(parsed)] ?? 'component' };
}

function buildDraftPositions(nodes: Node[]): Record<string, { x: number; y: number }> {
  return Object.fromEntries(
    nodes.map((n) => [nodeIdToEntityKey(n.id).id, { x: n.position.x, y: n.position.y }]),
  );
}

function persistNonDynamic(
  nodes: Node[],
  persisters: PositionPersisters,
  batchUpdatePositions: (updates: BatchUpdateItem[]) => void,
): void {
  if (nodes.length === 1) {
    updateSingleNode(nodes[0], persisters);
    return;
  }
  const updates = buildBatchUpdates(nodes, persisters.updateOriginEntity);
  if (updates.length > 0) batchUpdatePositions(updates);
}

function isMultiSelectModifier(event: React.MouseEvent): boolean {
  return event.shiftKey || event.ctrlKey || event.metaKey;
}

function persistOriginEntityPosition(
  node: Node,
  viewId: ViewId,
  mutate: (params: { viewId: ViewId; originEntityId: string; position: { x: number; y: number } }) => void,
): void {
  const originEntityId = getEntityId(toNodeId(node.id));
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
  const dynamicEnabled = useAppStore((state) => state.dynamicEnabled);
  const draftSetPositions = useAppStore((state) => state.draftSetPositions);
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
    [selectNode, selectCapability],
  );

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      selectEdge(toRelationId(edge.id));
    },
    [selectEdge],
  );

  const onPaneClick = useCallback(() => {
    clearSelection();
    selectCapability(null);
  }, [clearSelection, selectCapability]);

  const persisters = useMemo<PositionPersisters>(
    () => ({
      updateCapabilityPosition,
      updateComponentPosition,
      updateOriginEntity: (target: Node) =>
        currentViewId && persistOriginEntityPosition(target, currentViewId, updateOriginEntityPositionMutation.mutate),
    }),
    [updateCapabilityPosition, updateComponentPosition, currentViewId, updateOriginEntityPositionMutation],
  );

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (!canEdit(currentView) || !currentViewId) return;
      const selectedNodes = reactFlowInstance.getNodes().filter((n) => n.selected);
      const nodesToPersist = getNodesToPersist(node, selectedNodes);

      if (dynamicEnabled) {
        draftSetPositions(buildDraftPositions(nodesToPersist));
        return;
      }
      persistNonDynamic(nodesToPersist, persisters, batchUpdatePositions);
    },
    [persisters, currentView, currentViewId, reactFlowInstance, batchUpdatePositions, dynamicEnabled, draftSetPositions],
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
