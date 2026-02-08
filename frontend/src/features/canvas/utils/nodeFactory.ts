import type { Node } from '@xyflow/react';
import type { Component, View, Capability, ViewCapability, Position, AcquiredEntity, Vendor, InternalTeam } from '../../../api/types';
import type { CanvasPositionMap } from '../hooks/useCanvasLayout';
import { NODE_PREFIXES, type OriginEntityType } from '../../../constants/entityIdentifiers';

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

export const ORIGIN_ENTITY_PREFIXES = {
  acquired: NODE_PREFIXES.acquired,
  vendor: NODE_PREFIXES.vendor,
  team: NODE_PREFIXES.team,
} as const;

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
  const prefix = ORIGIN_ENTITY_PREFIXES[entityType];
  const nodeId = `${prefix}${entityId}`;
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
  selectedOriginEntityId: string | null
): Node => {
  const subtitle = entity.acquisitionDate
    ? new Date(entity.acquisitionDate).getFullYear().toString()
    : undefined;
  return createOriginEntityNode({ entityId: entity.id, entityType: 'acquired', name: entity.name, layoutPositions, selectedOriginEntityId, subtitle });
};

export const createVendorNode = (
  vendor: Vendor,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null
): Node =>
  createOriginEntityNode({ entityId: vendor.id, entityType: 'vendor', name: vendor.name, layoutPositions, selectedOriginEntityId, subtitle: vendor.implementationPartner });

export const createInternalTeamNode = (
  team: InternalTeam,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null
): Node =>
  createOriginEntityNode({ entityId: team.id, entityType: 'team', name: team.name, layoutPositions, selectedOriginEntityId, subtitle: team.department });

export const isOriginEntityNode = (nodeId: string): boolean => {
  return nodeId.startsWith(ORIGIN_ENTITY_PREFIXES.acquired) ||
         nodeId.startsWith(ORIGIN_ENTITY_PREFIXES.vendor) ||
         nodeId.startsWith(ORIGIN_ENTITY_PREFIXES.team);
};

export const getOriginEntityTypeFromNodeId = (nodeId: string): OriginEntityType | null => {
  if (nodeId.startsWith(ORIGIN_ENTITY_PREFIXES.acquired)) return 'acquired';
  if (nodeId.startsWith(ORIGIN_ENTITY_PREFIXES.vendor)) return 'vendor';
  if (nodeId.startsWith(ORIGIN_ENTITY_PREFIXES.team)) return 'team';
  return null;
};

export const extractOriginEntityId = (nodeId: string): string | null => {
  const entityType = getOriginEntityTypeFromNodeId(nodeId);
  if (!entityType) return null;
  return nodeId.replace(ORIGIN_ENTITY_PREFIXES[entityType], '');
};
