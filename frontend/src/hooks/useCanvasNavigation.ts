import { useCallback } from 'react';
import { useAppStore } from '../store/appStore';
import type { ComponentCanvasRef } from '../features/canvas/components/ComponentCanvas';
import type { ComponentId, CapabilityId } from '../api/types';

export function useCanvasNavigation(canvasRef: React.RefObject<ComponentCanvasRef | null>) {
  const currentView = useAppStore((state) => state.currentView);
  const selectNode = useAppStore((state) => state.selectNode);
  const selectCapability = useAppStore((state) => state.selectCapability);

  const navigateToComponent = useCallback(
    (componentId: string) => {
      const isInCurrentView = currentView?.components.some(
        (vc) => vc.componentId === componentId
      );

      if (isInCurrentView) {
        selectNode(componentId as ComponentId);
        canvasRef.current?.centerOnNode(componentId);
      }
    },
    [currentView, selectNode, canvasRef]
  );

  const navigateToCapability = useCallback(
    (capabilityId: string) => {
      const isOnCanvas = (currentView?.capabilities || []).some(
        (vc) => vc.capabilityId === capabilityId
      );

      if (isOnCanvas) {
        selectCapability(capabilityId as CapabilityId);
        selectNode(null);
        canvasRef.current?.centerOnNode(`cap-${capabilityId}`);
      }
    },
    [currentView, selectCapability, selectNode, canvasRef]
  );

  return { navigateToComponent, navigateToCapability };
}
