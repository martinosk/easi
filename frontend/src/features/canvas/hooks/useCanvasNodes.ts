import { useMemo, useEffect } from 'react';
import type { Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';

export const useCanvasNodes = (): Node[] => {
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const canvasCapabilities = useAppStore((state) => state.canvasCapabilities);
  const capabilities = useAppStore((state) => state.capabilities);
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const loadRealizationsByComponent = useAppStore((state) => state.loadRealizationsByComponent);

  useEffect(() => {
    if (!currentView) return;
    const componentIdsOnCanvas = currentView.components.map((vc) => vc.componentId);
    componentIdsOnCanvas.forEach((componentId) => {
      loadRealizationsByComponent(componentId);
    });
  }, [currentView?.id, currentView?.components.length, loadRealizationsByComponent]);

  return useMemo(() => {
    if (!currentView) return [];

    const componentNodes: Node[] = components
      .filter((component) => {
        return currentView.components.some(
          (vc) => vc.componentId === component.id
        );
      })
      .map((component) => {
        const viewComponent = currentView.components.find(
          (vc) => vc.componentId === component.id
        );

        const position = viewComponent
          ? { x: viewComponent.x, y: viewComponent.y }
          : { x: 400, y: 300 };

        return {
          id: component.id,
          type: 'component',
          position,
          data: {
            label: component.name,
            description: component.description,
            isSelected: selectedNodeId === component.id,
            customColor: viewComponent?.customColor,
          },
        };
      });

    const capabilityNodes: Node[] = canvasCapabilities
      .map((cc) => {
        const capability = capabilities.find((c) => c.id === cc.capabilityId);
        if (!capability) return null;

        const viewCapability = currentView.capabilities.find((vc) => vc.capabilityId === capability.id);

        return {
          id: `cap-${capability.id}`,
          type: 'capability' as const,
          position: { x: cc.x, y: cc.y },
          data: {
            label: capability.name,
            level: capability.level,
            maturityLevel: capability.maturityLevel,
            isSelected: selectedCapabilityId === capability.id,
            customColor: viewCapability?.customColor,
          },
        };
      })
      .filter((n) => n !== null) as Node[];

    return [...componentNodes, ...capabilityNodes];
  }, [components, currentView, selectedNodeId, canvasCapabilities, capabilities, selectedCapabilityId]);
};
