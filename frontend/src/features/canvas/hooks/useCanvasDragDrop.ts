import { useCallback } from 'react';
import type { ReactFlowInstance } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import type { CapabilityId, ComponentId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';

export const useCanvasDragDrop = (
  reactFlowInstance: ReactFlowInstance | null,
  onComponentDrop?: (componentId: string, x: number, y: number) => void
) => {
  const addCapabilityToCanvas = useAppStore((state) => state.addCapabilityToCanvas);
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

      if (!reactFlowInstance) return;

      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      if (componentId && onComponentDrop) {
        await onComponentDrop(componentId, position.x, position.y);
        await updateComponentPosition(componentId as ComponentId, position.x, position.y);
      } else if (capabilityId) {
        await addCapabilityToCanvas(capabilityId as CapabilityId, position.x, position.y);
        await updateCapabilityPosition(capabilityId as CapabilityId, position.x, position.y);
      }
    },
    [onComponentDrop, reactFlowInstance, addCapabilityToCanvas, updateComponentPosition, updateCapabilityPosition]
  );

  return { onDragOver, onDrop };
};
