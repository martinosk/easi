import { useEffect, useCallback, useState } from 'react';
import type { ReactFlowInstance, Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';

export const useCanvasViewport = (
  reactFlowInstance: ReactFlowInstance | null,
  nodes: Node[]
) => {
  const currentView = useAppStore((state) => state.currentView);
  const saveViewportState = useAppStore((state) => state.saveViewportState);
  const getViewportState = useAppStore((state) => state.getViewportState);
  const [isFirstLoad, setIsFirstLoad] = useState(true);

  useEffect(() => {
    if (!currentView || !reactFlowInstance) return;

    const savedViewport = getViewportState(currentView.id);
    if (savedViewport) {
      reactFlowInstance.setViewport(savedViewport, { duration: 300 });
      setIsFirstLoad(false);
      return;
    }

    if (isFirstLoad && nodes.length > 0) {
      setTimeout(() => {
        reactFlowInstance.fitView({ padding: 0.2, duration: 300 });
        setIsFirstLoad(false);
      }, 100);
    }
  }, [currentView?.id, reactFlowInstance, getViewportState, nodes.length, isFirstLoad]);

  const onMoveEnd = useCallback(() => {
    if (!currentView || !reactFlowInstance) return;

    const viewport = reactFlowInstance.getViewport();
    saveViewportState(currentView.id, viewport);
  }, [currentView, reactFlowInstance, saveViewportState]);

  return { onMoveEnd };
};
