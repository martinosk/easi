import { useCallback } from 'react';
import type { ReactFlowInstance } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';

export const useCanvasDragDrop = (
  reactFlowInstance: ReactFlowInstance | null,
  onComponentDrop?: (componentId: string, x: number, y: number) => void
) => {
  const addCapabilityToCanvas = useAppStore((state) => state.addCapabilityToCanvas);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const componentId = event.dataTransfer.getData('componentId');
      const capabilityId = event.dataTransfer.getData('capabilityId');

      if (!reactFlowInstance) return;

      const bounds = (event.target as HTMLElement).getBoundingClientRect();
      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX - bounds.left,
        y: event.clientY - bounds.top,
      });

      if (componentId && onComponentDrop) {
        onComponentDrop(componentId, position.x, position.y);
      } else if (capabilityId) {
        addCapabilityToCanvas(capabilityId, position.x, position.y);
      }
    },
    [onComponentDrop, reactFlowInstance, addCapabilityToCanvas]
  );

  return { onDragOver, onDrop };
};
