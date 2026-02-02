import { useCallback, useState } from 'react';
import { useReactFlow } from '@xyflow/react';
import toast from 'react-hot-toast';
import { calculateAutoLayout } from '../../../utils/autoLayout';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { canEdit } from '../../../utils/hateoas';
import { useUpdateOriginEntityPosition } from '../../views/hooks/useViews';

export function useAutoLayout() {
  const reactFlowInstance = useReactFlow();
  const { batchUpdatePositions } = useCanvasLayoutContext();
  const { currentView, currentViewId } = useCurrentView();
  const updateOriginEntityPositionMutation = useUpdateOriginEntityPosition();
  const [isLayouting, setIsLayouting] = useState(false);

  const applyAutoLayout = useCallback(async () => {
    if (!reactFlowInstance || !currentViewId || !canEdit(currentView)) return;

    const nodes = reactFlowInstance.getNodes();
    const edges = reactFlowInstance.getEdges();

    if (nodes.length === 0) {
      toast.error('No entities to layout');
      return;
    }

    setIsLayouting(true);

    try {
      const layoutedNodes = calculateAutoLayout(nodes, edges);

      const layoutUpdates = layoutedNodes
        .filter((node) => node.type !== 'originEntity')
        .map((node) => ({
          elementId: node.type === 'capability' ? node.id.replace('cap-', '') : node.id,
          x: node.position.x,
          y: node.position.y,
        }));

      if (layoutUpdates.length > 0) {
        await batchUpdatePositions(layoutUpdates);
      }

      const originUpdates = layoutedNodes
        .filter((node) => node.type === 'originEntity')
        .map((node) => updateOriginEntityPositionMutation.mutateAsync({
          viewId: currentViewId,
          originEntityId: node.id,
          position: { x: node.position.x, y: node.position.y },
        }));

      if (originUpdates.length > 0) {
        await Promise.all(originUpdates);
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