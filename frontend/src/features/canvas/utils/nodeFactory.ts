import type { Node } from '@xyflow/react';
import type {
  AcquiredEntity,
  Capability,
  Component,
  InternalTeam,
  Position,
  Vendor,
  View,
  ViewCapability,
} from '../../../api/types';
import { makeNodeId, type OriginEntityType } from '../../../constants/entityIdentifiers';
import type { CanvasPositionMap } from '../hooks/useCanvasLayout';

const DEFAULT_POSITION: Position = { x: 400, y: 300 };

export const createComponentNode = (
  component: Component,
  currentView: View,
  layoutPositions: CanvasPositionMap,
  selectedNodeId: string | null,
): Node => {
  const viewComponent = currentView.components.find((vc) => vc.componentId === component.id);

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

interface CapabilityNodeParams {
  capabilityId: string;
  capability: Capability;
  layoutPositions: CanvasPositionMap;
  viewCapability: ViewCapability | undefined;
  selectedCapabilityId: string | null;
}

export const createCapabilityNode = (params: CapabilityNodeParams): Node => {
  const { capabilityId, capability, layoutPositions, viewCapability, selectedCapabilityId } = params;
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

interface OriginEntityNodeParams {
  entityId: string;
  entityType: OriginEntityType;
  name: string;
  layoutPositions: CanvasPositionMap;
  selectedOriginEntityId: string | null;
  subtitle?: string;
}

export const createOriginEntityNode = (params: OriginEntityNodeParams): Node => {
  const { entityId, entityType, name, layoutPositions, selectedOriginEntityId, subtitle } = params;
  const nodeId = makeNodeId(entityType, entityId);
  const layoutPosition = layoutPositions[entityId];
  const position = layoutPosition ?? DEFAULT_POSITION;

  return {
    id: nodeId,
    type: 'originEntity',
    position,
    data: {
      label: name,
      entityType,
      isSelected: selectedOriginEntityId === nodeId,
      subtitle,
    },
  };
};

export const createAcquiredEntityNode = (
  entity: AcquiredEntity,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
): Node => {
  const subtitle = entity.acquisitionDate ? new Date(entity.acquisitionDate).getFullYear().toString() : undefined;
  return createOriginEntityNode({
    entityId: entity.id,
    entityType: 'acquired',
    name: entity.name,
    layoutPositions,
    selectedOriginEntityId,
    subtitle,
  });
};

export const createVendorNode = (
  vendor: Vendor,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
): Node =>
  createOriginEntityNode({
    entityId: vendor.id,
    entityType: 'vendor',
    name: vendor.name,
    layoutPositions,
    selectedOriginEntityId,
    subtitle: vendor.implementationPartner,
  });

export const createInternalTeamNode = (
  team: InternalTeam,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
): Node =>
  createOriginEntityNode({
    entityId: team.id,
    entityType: 'team',
    name: team.name,
    layoutPositions,
    selectedOriginEntityId,
    subtitle: team.department,
  });

