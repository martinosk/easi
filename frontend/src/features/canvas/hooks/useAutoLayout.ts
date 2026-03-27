import { useCallback, useState } from 'react';
import type { Node } from '@xyflow/react';
import { useReactFlow } from '@xyflow/react';
import toast from 'react-hot-toast';
import { calculateAutoLayout } from '../../../utils/autoLayout';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { canEdit } from '../../../utils/hateoas';
import { useUpdateOriginEntityPosition } from '../../views/hooks/useViews';

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

export function useAutoLayout() {
  const reactFlowInstance = useReactFlow();
  const { batchUpdatePositions } = useCanvasLayoutContext();
  const { currentView, currentViewId } = useCurrentView();
  const updateOriginEntityPositionMutation = useUpdateOriginEntityPosition();
  const [isLayouting, setIsLayouting] = useState(false);

  const applyAutoLayout = useCallback(async () => {
    const isEditable = reactFlowInstance && currentViewId && canEdit(currentView);
    if (!isEditable) return;

    const nodes = reactFlowInstance.getNodes();
    if (nodes.length === 0) {
      toast.error('No entities to layout');
      return;
    }

    setIsLayouting(true);

    try {
      const layoutedNodes = calculateAutoLayout(nodes, reactFlowInstance.getEdges());
      const { layout, origin } = partitionByType(layoutedNodes);

      if (layout.length > 0) {
        await batchUpdatePositions(layout.map(toLayoutUpdate));
      }

      if (origin.length > 0) {
        await Promise.all(origin.map((node) =>
          updateOriginEntityPositionMutation.mutateAsync({
            viewId: currentViewId,
            originEntityId: node.id,
            position: { x: node.position.x, y: node.position.y },
          })
        ));
      }

      const { zoom } = reactFlowInstance.getViewport();
      window.requestAnimationFrame(() => {
        reactFlowInstance.fitView({ padding: 0.2, duration: 800, minZoom: zoom, maxZoom: zoom });
      });

      toast.success('Layout applied successfully');
    } catch (error) {
      console.error('Auto-layout failed:', error);
      toast.error('Failed to apply layout');
    } finally {
      setIsLayouting(false);
    }
  }, [reactFlowInstance, currentViewId, currentView, batchUpdatePositions, updateOriginEntityPositionMutation]);

  return { applyAutoLayout, isLayouting };
}