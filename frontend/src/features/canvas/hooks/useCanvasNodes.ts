import type { Node } from '@xyflow/react';
import { useMemo } from 'react';
import type {
  AcquiredEntity,
  Capability,
  Component,
  InternalTeam,
  Vendor,
  View,
  ViewCapability,
} from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import type { Position } from '../../../store/slices/dynamicModeSlice';
import type { EntityRef } from '../utils/dynamicMode';
import {
  createAcquiredEntityNode,
  createCapabilityNode,
  createComponentNode,
  createInternalTeamNode,
  createVendorNode,
} from '../utils/nodeFactory';

interface NodeBuildContext {
  positions: Record<string, Position>;
  currentView: View;
  components: Component[];
  capabilities: Capability[];
  acquiredEntities: AcquiredEntity[];
  vendors: Vendor[];
  internalTeams: InternalTeam[];
  selectedNodeId: string | null;
  selectedCapabilityId: string | null;
}

function entitiesFromView(view: View): EntityRef[] {
  return [
    ...view.components.map((c) => ({ id: c.componentId, type: 'component' as const })),
    ...(view.capabilities ?? []).map((c) => ({ id: c.capabilityId, type: 'capability' as const })),
    ...(view.originEntities ?? []).map((oe) => ({ id: oe.originEntityId, type: 'originEntity' as const })),
  ];
}

function buildOriginEntityNode(id: string, ctx: NodeBuildContext): Node | null {
  const acquired = ctx.acquiredEntities.find((e) => e.id === id);
  if (acquired) return createAcquiredEntityNode(acquired, ctx.positions, ctx.selectedNodeId);
  const vendor = ctx.vendors.find((v) => v.id === id);
  if (vendor) return createVendorNode(vendor, ctx.positions, ctx.selectedNodeId);
  const team = ctx.internalTeams.find((t) => t.id === id);
  if (team) return createInternalTeamNode(team, ctx.positions, ctx.selectedNodeId);
  return null;
}

function buildCapabilityNode(id: string, ctx: NodeBuildContext): Node | null {
  const capability = ctx.capabilities.find((c) => c.id === id);
  if (!capability) return null;
  const viewCapability = ctx.currentView.capabilities?.find((vc: ViewCapability) => vc.capabilityId === id);
  return createCapabilityNode({
    capabilityId: id,
    capability,
    layoutPositions: ctx.positions,
    viewCapability,
    selectedCapabilityId: ctx.selectedCapabilityId,
  });
}

function buildComponentNode(id: string, ctx: NodeBuildContext): Node | null {
  const component = ctx.components.find((c) => c.id === id);
  if (!component) return null;
  return createComponentNode(component, ctx.currentView, ctx.positions, ctx.selectedNodeId);
}

const NODE_BUILDERS: Record<EntityRef['type'], (id: string, ctx: NodeBuildContext) => Node | null> = {
  component: buildComponentNode,
  capability: buildCapabilityNode,
  originEntity: buildOriginEntityNode,
};

function buildNodesFromRefs(refs: readonly EntityRef[], ctx: NodeBuildContext): Node[] {
  const nodes: Node[] = [];
  for (const ref of refs) {
    const node = NODE_BUILDERS[ref.type](ref.id, ctx);
    if (node) nodes.push(node);
  }
  return nodes;
}

function originEntityPositions(view: View): Record<string, Position> {
  const out: Record<string, Position> = {};
  for (const oe of view.originEntities ?? []) out[oe.originEntityId] = { x: oe.x, y: oe.y };
  return out;
}

function selectRefsAndPositions(
  dynamicEnabled: boolean,
  dynamicEntities: readonly EntityRef[],
  dynamicPositions: Record<string, Position>,
  layoutPositions: Record<string, Position>,
  view: View,
): { refs: readonly EntityRef[]; positions: Record<string, Position> } {
  if (dynamicEnabled) {
    return { refs: dynamicEntities, positions: { ...layoutPositions, ...dynamicPositions } };
  }
  return {
    refs: entitiesFromView(view),
    positions: { ...layoutPositions, ...originEntityPositions(view) },
  };
}

export const useCanvasNodes = (): Node[] => {
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: capabilities = [] } = useCapabilities();
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { positions: layoutPositions } = useCanvasLayoutContext();
  const { data: acquiredEntities = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: internalTeams = [] } = useInternalTeamsQuery();

  const dynamicEnabled = useAppStore((state) => state.dynamicEnabled);
  const dynamicEntities = useAppStore((state) => state.dynamicEntities);
  const dynamicPositions = useAppStore((state) => state.dynamicPositions);

  return useMemo(() => {
    if (!currentView) return [];
    const { refs, positions } = selectRefsAndPositions(
      dynamicEnabled,
      dynamicEntities,
      dynamicPositions,
      layoutPositions,
      currentView,
    );
    return buildNodesFromRefs(refs, {
      positions,
      currentView,
      components,
      capabilities,
      acquiredEntities,
      vendors,
      internalTeams,
      selectedNodeId,
      selectedCapabilityId,
    });
  }, [
    components,
    currentView,
    selectedNodeId,
    capabilities,
    selectedCapabilityId,
    layoutPositions,
    acquiredEntities,
    vendors,
    internalTeams,
    dynamicEnabled,
    dynamicEntities,
    dynamicPositions,
  ]);
};

export { entitiesFromView };
