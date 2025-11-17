import { useCallback } from 'react';
import { useAppStore } from '../store/appStore';
import type { ComponentCanvasRef } from '../components/ComponentCanvas';

export function useCanvasNavigation(canvasRef: React.RefObject<ComponentCanvasRef | null>) {
  const currentView = useAppStore((state) => state.currentView);
  const selectNode = useAppStore((state) => state.selectNode);

  const navigateToComponent = useCallback(
    (componentId: string) => {
      const isInCurrentView = currentView?.components.some(
        (vc) => vc.componentId === componentId
      );

      if (isInCurrentView) {
        selectNode(componentId);
        canvasRef.current?.centerOnNode(componentId);
      }
    },
    [currentView, selectNode, canvasRef]
  );

  return { navigateToComponent };
}
