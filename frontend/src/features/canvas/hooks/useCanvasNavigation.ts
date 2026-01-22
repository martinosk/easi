import { useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { ComponentCanvasRef } from '../components/ComponentCanvas';
import { toComponentId, toCapabilityId } from '../../../api/types';
import type { View, ViewComponent, ViewCapability } from '../../../api/types';

function isComponentInView(view: View | null, componentId: string): boolean {
  return view?.components.some((vc: ViewComponent) => vc.componentId === componentId) ?? false;
}

function isCapabilityInView(view: View | null, capabilityId: string): boolean {
  return view?.capabilities?.some((vc: ViewCapability) => vc.capabilityId === capabilityId) ?? false;
}

function centerOnNodeIfNeeded(
  canvasRef: React.RefObject<ComponentCanvasRef | null>,
  nodeId: string,
  shouldCenter: boolean
): void {
  if (shouldCenter) {
    canvasRef.current?.centerOnNode(nodeId);
  }
}

export function useCanvasNavigation(canvasRef: React.RefObject<ComponentCanvasRef | null>) {
  const { currentView } = useCurrentView();
  const selectNode = useAppStore((state) => state.selectNode);
  const selectCapability = useAppStore((state) => state.selectCapability);

  const navigateToComponent = useCallback(
    (componentId: string) => {
      selectNode(toComponentId(componentId));
      selectCapability(null);
      centerOnNodeIfNeeded(canvasRef, componentId, isComponentInView(currentView, componentId));
    },
    [currentView, selectNode, selectCapability, canvasRef]
  );

  const navigateToCapability = useCallback(
    (capabilityId: string) => {
      selectCapability(toCapabilityId(capabilityId));
      selectNode(null);
      centerOnNodeIfNeeded(canvasRef, `cap-${capabilityId}`, isCapabilityInView(currentView, capabilityId));
    },
    [currentView, selectCapability, selectNode, canvasRef]
  );

  const navigateToOriginEntity = useCallback(
    (nodeId: string) => {
      selectNode(nodeId as import('../../../api/types').ComponentId);
      selectCapability(null);
      canvasRef.current?.centerOnNode(nodeId);
    },
    [selectNode, selectCapability, canvasRef]
  );

  return { navigateToComponent, navigateToCapability, navigateToOriginEntity };
}
