import { useMemo, useEffect } from 'react';
import type { Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import type { Component, View, Capability, ViewCapability, Position } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import type { CanvasPositionMap } from './useCanvasLayout';

const DEFAULT_POSITION: Position = { x: 400, y: 300 };

const createComponentNode = (
  component: Component,
  currentView: View,
  layoutPositions: CanvasPositionMap,
  selectedNodeId: string | null
): Node => {
  const viewComponent = currentView.components.find(
    (vc) => vc.componentId === component.id
  );

  const layoutPosition = layoutPositions[component.id];
  const position = layoutPosition ?? DEFAULT_POSITION;

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
  capabilityId: string,
  capability: Capability,
  layoutPositions: CanvasPositionMap,
  viewCapability: ViewCapability | undefined,
  selectedCapabilityId: string | null
): Node => {
  const layoutPosition = layoutPositions[capabilityId];
  const position = layoutPosition ?? DEFAULT_POSITION;

  return {
    id: `cap-${capability.id}`,
    type: 'capability' as const,
    position,
    data: {
      label: capability.name,
      level: capability.level,
      maturityLevel: capability.maturityLevel,
      isSelected: selectedCapabilityId === capability.id,
      customColor: viewCapability?.customColor,
    },
  };
};

const isComponentInView = (component: Component, currentView: View): boolean =>
  currentView.components.some((vc) => vc.componentId === component.id);

export const useCanvasNodes = (): Node[] => {
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const capabilities = useAppStore((state) => state.capabilities);
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const loadRealizationsByComponent = useAppStore((state) => state.loadRealizationsByComponent);
  const { positions: layoutPositions } = useCanvasLayoutContext();

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
