import type { Node } from '@xyflow/react';
import type { Component, View, Capability, ViewCapability, Position, AcquiredEntity, Vendor, InternalTeam } from '../../../api/types';
import type { CanvasPositionMap } from '../hooks/useCanvasLayout';
import type { OriginEntityType } from '../../../components/canvas';

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

export const ORIGIN_ENTITY_PREFIXES = {
  acquired: 'acq-',
  vendor: 'vendor-',
  team: 'team-',
} as const;

export const createOriginEntityNode = (
  entityId: string,
  entityType: OriginEntityType,
  name: string,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null,
  subtitle?: string
): Node => {
  const prefix = ORIGIN_ENTITY_PREFIXES[entityType];
  const nodeId = `${prefix}${entityId}`;
  const layoutPosition = layoutPositions[nodeId];
  const position = layoutPosition ?? DEFAULT_POSITION;

  return {
    id: nodeId,
    type: 'originEntity',
    position,
    data: {
      label: name,
      entityType,
      isSelected: selectedOriginEntityId === entityId,
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
  return createOriginEntityNode(
    entity.id,
    'acquired',
    entity.name,
    layoutPositions,
    selectedOriginEntityId,
    subtitle
  );
};

export const createVendorNode = (
  vendor: Vendor,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null
): Node => {
  return createOriginEntityNode(
    vendor.id,
    'vendor',
    vendor.name,
    layoutPositions,
    selectedOriginEntityId,
    vendor.implementationPartner
  );
};

export const createInternalTeamNode = (
  team: InternalTeam,
  layoutPositions: CanvasPositionMap,
  selectedOriginEntityId: string | null
): Node => {
  return createOriginEntityNode(
    team.id,
    'team',
    team.name,
    layoutPositions,
    selectedOriginEntityId,
    team.department
  );
};

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
