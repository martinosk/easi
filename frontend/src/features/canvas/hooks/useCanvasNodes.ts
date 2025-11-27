import { useMemo, useEffect } from 'react';
import type { Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import type { Component, View, Capability, ViewCapability, CanvasCapability } from '../../../api/types';

interface Position {
  x: number;
  y: number;
}

const DEFAULT_POSITION: Position = { x: 400, y: 300 };

const createComponentNode = (
  component: Component,
  currentView: View,
  selectedNodeId: string | null
): Node => {
  const viewComponent = currentView.components.find(
    (vc) => vc.componentId === component.id
  );

  const position = viewComponent
    ? { x: viewComponent.x, y: viewComponent.y }
    : DEFAULT_POSITION;

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
};

const createCapabilityNode = (
  cc: CanvasCapability,
  capability: Capability,
  viewCapability: ViewCapability | undefined,
  selectedCapabilityId: string | null
): Node => ({
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
});

const isComponentInView = (component: Component, currentView: View): boolean =>
  currentView.components.some((vc) => vc.componentId === component.id);

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
    currentView.components.forEach((vc) => {
      loadRealizationsByComponent(vc.componentId);
    });
  }, [currentView?.id, currentView?.components.length, loadRealizationsByComponent]);

  return useMemo(() => {
    if (!currentView) return [];

    const componentNodes = components
      .filter((component) => isComponentInView(component, currentView))
      .map((component) => createComponentNode(component, currentView, selectedNodeId));

    const capabilityNodes = canvasCapabilities
      .map((cc) => {
        const capability = capabilities.find((c) => c.id === cc.capabilityId);
        if (!capability) return null;

        const viewCapability = currentView.capabilities.find(
          (vc) => vc.capabilityId === capability.id
        );

        return createCapabilityNode(cc, capability, viewCapability, selectedCapabilityId);
      })
      .filter((n): n is Node => n !== null);

    return [...componentNodes, ...capabilityNodes];
  }, [components, currentView, selectedNodeId, canvasCapabilities, capabilities, selectedCapabilityId]);
};
