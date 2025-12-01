import { useMemo, useEffect } from 'react';
import type { Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { createComponentNode, createCapabilityNode, isComponentInView } from '../utils/nodeFactory';

export const useCanvasNodes = (): Node[] => {
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const capabilities = useAppStore((state) => state.capabilities);
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const loadRealizationsByComponent = useAppStore((state) => state.loadRealizationsByComponent);
  const { positions: layoutPositions } = useCanvasLayoutContext();

  const viewComponentIds = useMemo(
    () => currentView?.components.map((vc) => vc.componentId).join(',') ?? '',
    [currentView?.components]
  );

  useEffect(() => {
    if (!currentView || !viewComponentIds) return;
    currentView.components.forEach((vc) => {
      loadRealizationsByComponent(vc.componentId);
    });
  }, [currentView?.id, viewComponentIds, loadRealizationsByComponent]);

  return useMemo(() => {
    if (!currentView) return [];

    const componentNodes = components
      .filter((component) => isComponentInView(component, currentView))
      .map((component) => createComponentNode(component, currentView, layoutPositions, selectedNodeId));

    const capabilityNodes = (currentView.capabilities || [])
      .map((vc) => {
        const capability = capabilities.find((c) => c.id === vc.capabilityId);
        if (!capability) return null;

        return createCapabilityNode(vc.capabilityId, capability, layoutPositions, vc, selectedCapabilityId);
      })
      .filter((n): n is Node => n !== null);

    return [...componentNodes, ...capabilityNodes];
  }, [components, currentView, selectedNodeId, capabilities, selectedCapabilityId, layoutPositions]);
};
