import { useCallback } from 'react';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useCurrentViewElementIds } from '../../views/hooks/useCurrentViewElementIds';
import type { ComponentCanvasRef } from '../components/ComponentCanvas';

function centerOnNodeIfNeeded(
  canvasRef: React.RefObject<ComponentCanvasRef | null>,
  nodeId: string,
  shouldCenter: boolean,
): void {
  if (shouldCenter) {
    canvasRef.current?.centerOnNode(nodeId);
  }
}

export function useCanvasNavigation(canvasRef: React.RefObject<ComponentCanvasRef | null>) {
  const selectNode = useAppStore((state) => state.selectNode);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const { components: componentsInView, capabilities: capabilitiesInView } = useCurrentViewElementIds();

  const navigateToComponent = useCallback(
    (componentId: string) => {
      selectNode(toComponentId(componentId));
      selectCapability(null);
      centerOnNodeIfNeeded(canvasRef, componentId, componentsInView.has(componentId));
    },
    [componentsInView, selectNode, selectCapability, canvasRef],
  );

  const navigateToCapability = useCallback(
    (capabilityId: string) => {
      selectCapability(toCapabilityId(capabilityId));
      selectNode(null);
      centerOnNodeIfNeeded(canvasRef, `cap-${capabilityId}`, capabilitiesInView.has(capabilityId));
    },
    [capabilitiesInView, selectCapability, selectNode, canvasRef],
  );

  const navigateToOriginEntity = useCallback(
    (nodeId: string) => {
      selectNode(nodeId as import('../../../api/types').ComponentId);
      selectCapability(null);
      canvasRef.current?.centerOnNode(nodeId);
    },
    [selectNode, selectCapability, canvasRef],
  );

  return { navigateToComponent, navigateToCapability, navigateToOriginEntity };
}
