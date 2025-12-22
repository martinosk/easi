import { useCallback } from 'react';
import type { ReactFlowInstance } from '@xyflow/react';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useAddCapabilityToView } from '../../views/hooks/useViews';
import type { CapabilityId, ComponentId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';

export const useCanvasDragDrop = (
  reactFlowInstance: ReactFlowInstance | null,
  onComponentDrop?: (componentId: string, x: number, y: number) => void
) => {
  const { currentViewId } = useCurrentView();
  const addCapabilityToViewMutation = useAddCapabilityToView();
  const { updateComponentPosition, updateCapabilityPosition } = useCanvasLayoutContext();

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  const onDrop = useCallback(
    async (event: React.DragEvent) => {
      event.preventDefault();

      const componentId = event.dataTransfer.getData('componentId');
      const capabilityId = event.dataTransfer.getData('capabilityId');

      if (!reactFlowInstance || !currentViewId) return;

      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      if (componentId && onComponentDrop) {
        await onComponentDrop(componentId, position.x, position.y);
        await updateComponentPosition(componentId as ComponentId, position.x, position.y);
      } else if (capabilityId) {
        await addCapabilityToViewMutation.mutateAsync({
          viewId: currentViewId,
          request: {
            capabilityId: capabilityId as CapabilityId,
            x: position.x,
            y: position.y
          }
        });
        await updateCapabilityPosition(capabilityId as CapabilityId, position.x, position.y);
      }
    },
    [onComponentDrop, reactFlowInstance, currentViewId, addCapabilityToViewMutation, updateComponentPosition, updateCapabilityPosition]
  );

  return { onDragOver, onDrop };
};
