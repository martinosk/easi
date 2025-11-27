import { useCallback } from 'react';
import { useAppStore } from '../store/appStore';
import type { ComponentCanvasRef } from '../features/canvas/components/ComponentCanvas';
import type { ComponentId, CapabilityId } from '../api/types';

export function useCanvasNavigation(canvasRef: React.RefObject<ComponentCanvasRef | null>) {
  const currentView = useAppStore((state) => state.currentView);
  const selectNode = useAppStore((state) => state.selectNode);
  const canvasCapabilities = useAppStore((state) => state.canvasCapabilities);
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
      const isOnCanvas = canvasCapabilities.some(
        (cc) => cc.capabilityId === capabilityId
      );

      if (isOnCanvas) {
        selectCapability(capabilityId as CapabilityId);
        selectNode(null);
        canvasRef.current?.centerOnNode(`cap-${capabilityId}`);
      }
    },
    [canvasCapabilities, selectCapability, selectNode, canvasRef]
  );

  return { navigateToComponent, navigateToCapability };
}
