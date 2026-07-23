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

export type CanvasPositionMap = Record<string, Position>;

const DEFAULT_POSITION: Position = { x: 400, y: 300 };

export const createComponentNode = (
  component: Component,
  currentView: View,
  positions: CanvasPositionMap,
  selectedNodeId: string | null,
): Node => {
  const viewComponent = currentView.components.find((vc) => vc.componentId === component.id);

  const position = positions[component.id] ?? DEFAULT_POSITION;

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
  positions: CanvasPositionMap;
  viewCapability: ViewCapability | undefined;
  selectedCapabilityId: string | null;
}

export const createCapabilityNode = (params: CapabilityNodeParams): Node => {
  const { capabilityId, capability, positions, viewCapability, selectedCapabilityId } = params;
  const position = positions[capabilityId] ?? DEFAULT_POSITION;

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
  positions: CanvasPositionMap;
  selectedOriginEntityId: string | null;
  subtitle?: string;
}

export const createOriginEntityNode = (params: OriginEntityNodeParams): Node => {
  const { entityId, entityType, name, positions, selectedOriginEntityId, subtitle } = params;
  const nodeId = makeNodeId(entityType, entityId);
  const position = positions[entityId] ?? DEFAULT_POSITION;

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
  positions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
): Node => {
  const subtitle = entity.acquisitionDate ? new Date(entity.acquisitionDate).getFullYear().toString() : undefined;
  return createOriginEntityNode({
    entityId: entity.id,
    entityType: 'acquired',
    name: entity.name,
    positions,
    selectedOriginEntityId,
    subtitle,
  });
};

export const createVendorNode = (
  vendor: Vendor,
  positions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
): Node =>
  createOriginEntityNode({
    entityId: vendor.id,
    entityType: 'vendor',
    name: vendor.name,
    positions,
    selectedOriginEntityId,
    subtitle: vendor.implementationPartner,
  });

export const createInternalTeamNode = (
  team: InternalTeam,
  positions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
): Node =>
  createOriginEntityNode({
    entityId: team.id,
    entityType: 'team',
    name: team.name,
    positions,
    selectedOriginEntityId,
    subtitle: team.department,
  });

