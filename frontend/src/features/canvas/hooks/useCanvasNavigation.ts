import { useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { ComponentCanvasRef } from '../components/ComponentCanvas';
import { toComponentId, toCapabilityId } from '../../../api/types';
import type { ViewComponent, ViewCapability } from '../../../api/types';

export function useCanvasNavigation(canvasRef: React.RefObject<ComponentCanvasRef | null>) {
  const { currentView } = useCurrentView();
  const selectNode = useAppStore((state) => state.selectNode);
  const selectCapability = useAppStore((state) => state.selectCapability);

  const navigateToComponent = useCallback(
    (componentId: string) => {
      selectNode(toComponentId(componentId));
      selectCapability(null);

      const isInCurrentView = currentView?.components.some(
        (vc: ViewComponent) => vc.componentId === componentId
      );

      if (isInCurrentView) {
        canvasRef.current?.centerOnNode(componentId);
      }
    },
    [currentView, selectNode, selectCapability, canvasRef]
  );

  const navigateToCapability = useCallback(
    (capabilityId: string) => {
      selectCapability(toCapabilityId(capabilityId));
      selectNode(null);

      const isOnCanvas = (currentView?.capabilities || []).some(
        (vc: ViewCapability) => vc.capabilityId === capabilityId
      );

      if (isOnCanvas) {
        canvasRef.current?.centerOnNode(`cap-${capabilityId}`);
      }
    },
    [currentView, selectCapability, selectNode, canvasRef]
  );

  return { navigateToComponent, navigateToCapability };
}
