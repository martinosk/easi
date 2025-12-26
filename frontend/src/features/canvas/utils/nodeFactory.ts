import type { Node } from '@xyflow/react';
import type { Component, View, Capability, ViewCapability, Position } from '../../../api/types';
import type { CanvasPositionMap } from '../hooks/useCanvasLayout';

const DEFAULT_POSITION: Position = { x: 400, y: 300 };

export const createComponentNode = (
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

export const createCapabilityNode = (
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
      maturityValue: capability.maturityValue,
      maturitySection: capability.maturitySection?.name,
      isSelected: selectedCapabilityId === capability.id,
      customColor: viewCapability?.customColor,
    },
  };
};

export const isComponentInView = (component: Component, currentView: View): boolean =>
  currentView.components.some((vc) => vc.componentId === component.id);
