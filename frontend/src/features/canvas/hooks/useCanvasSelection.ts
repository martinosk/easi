import type { Edge, Node } from '@xyflow/react';
import { useReactFlow } from '@xyflow/react';
import { useCallback } from 'react';
import { toCapabilityId, toComponentId, toRelationId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
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

function isMultiSelectModifier(event: React.MouseEvent): boolean {
  return event.shiftKey || event.ctrlKey || event.metaKey;
}

function getNodesToPersist(node: Node, selectedNodes: Node[]): Node[] {
  return selectedNodes.length > 0 ? selectedNodes : [node];
}

export const useCanvasSelection = () => {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const draftSetPositions = useAppStore((state) => state.draftSetPositions);
  const { currentView, currentViewId } = useCurrentView();
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

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (!canEdit(currentView) || !currentViewId) return;
      const selectedNodes = reactFlowInstance.getNodes().filter((n) => n.selected);
      const nodesToPersist = getNodesToPersist(node, selectedNodes);
      draftSetPositions(buildDraftPositions(nodesToPersist));
    },
    [currentView, currentViewId, reactFlowInstance, draftSetPositions],
  );

  return {
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    onNodeDragStop,
  };
};
