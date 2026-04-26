import type { Node, ReactFlowInstance } from '@xyflow/react';
import { useReactFlow } from '@xyflow/react';
import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import { useAppStore } from '../../../store/appStore';
import { calculateAutoLayout } from '../../../utils/autoLayout';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useUpdateOriginEntityPosition } from '../../views/hooks/useViews';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';

function toLayoutUpdate(node: Node) {
  return {
    elementId: node.type === 'capability' ? node.id.replace('cap-', '') : node.id,
    x: node.position.x,
    y: node.position.y,
  };
}

function partitionByType(nodes: Node[]) {
  const layout: Node[] = [];
  const origin: Node[] = [];
  for (const node of nodes) {
    (node.type === 'originEntity' ? origin : layout).push(node);
  }
  return { layout, origin };
}

function nodeEntityId(node: Node): string {
  return node.type === 'capability' ? node.id.replace('cap-', '') : node.id;
}

function fitAfterRender(reactFlow: ReactFlowInstance) {
  const { zoom } = reactFlow.getViewport();
  window.requestAnimationFrame(() => {
    reactFlow.fitView({ padding: 0.2, duration: 800, minZoom: zoom, maxZoom: zoom });
  });
}

function useDraftApply() {
  const draftSetPositions = useAppStore((s) => s.draftSetPositions);
  return useCallback(
    (nodes: Node[]) => {
      const updates: Record<string, { x: number; y: number }> = {};
      for (const node of nodes) {
        updates[nodeEntityId(node)] = { x: node.position.x, y: node.position.y };
      }
      draftSetPositions(updates);
    },
    [draftSetPositions],
  );
}

function useServerApply() {
  const { batchUpdatePositions } = useCanvasLayoutContext();
  const { currentViewId } = useCurrentView();
  const updateOriginEntityPositionMutation = useUpdateOriginEntityPosition();

  return useCallback(
    async (nodes: Node[]) => {
      if (!currentViewId) return;
      const { layout, origin } = partitionByType(nodes);
      if (layout.length > 0) {
        await batchUpdatePositions(layout.map(toLayoutUpdate));
      }
      if (origin.length > 0) {
        await Promise.all(
          origin.map((node) =>
            updateOriginEntityPositionMutation.mutateAsync({
              viewId: currentViewId,
              originEntityId: node.id,
              position: { x: node.position.x, y: node.position.y },
            }),
          ),
        );
      }
    },
    [currentViewId, batchUpdatePositions, updateOriginEntityPositionMutation],
  );
}

function isLayoutAvailable(
  reactFlowInstance: ReactFlowInstance | null,
  currentViewId: string | null,
  currentView: Parameters<typeof canEdit>[0],
): boolean {
  return Boolean(reactFlowInstance) && Boolean(currentViewId) && canEdit(currentView);
}

export function useAutoLayout() {
  const reactFlowInstance = useReactFlow();
  const { currentView, currentViewId } = useCurrentView();
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const draftActive = dynamicViewId !== null && dynamicViewId === currentViewId;
  const applyToDraft = useDraftApply();
  const applyToServer = useServerApply();
  const [isLayouting, setIsLayouting] = useState(false);

  const applyAutoLayout = useCallback(async () => {
    if (!isLayoutAvailable(reactFlowInstance, currentViewId, currentView)) return;
    const nodes = reactFlowInstance.getNodes();
    if (nodes.length === 0) {
      toast.error('No entities to layout');
      return;
    }

    setIsLayouting(true);
    try {
      const layoutedNodes = calculateAutoLayout(nodes, reactFlowInstance.getEdges());
      if (draftActive) applyToDraft(layoutedNodes);
      else await applyToServer(layoutedNodes);
      fitAfterRender(reactFlowInstance);
      toast.success('Layout applied successfully');
    } catch (error) {
      console.error('Auto-layout failed:', error);
      toast.error('Failed to apply layout');
    } finally {
      setIsLayouting(false);
    }
  }, [reactFlowInstance, currentViewId, currentView, draftActive, applyToDraft, applyToServer]);

  return { applyAutoLayout, isLayouting };
}
