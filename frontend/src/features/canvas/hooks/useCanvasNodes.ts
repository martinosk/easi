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
import type { Position } from '../../../store/slices/dynamicModeSlice';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { EntityRef } from '../utils/dynamicMode';
import {
  createAcquiredEntityNode,
  createCapabilityNode,
  createComponentNode,
  createInternalTeamNode,
  createVendorNode,
} from '../utils/nodeFactory';
import { viewPositions } from '../utils/viewPositions';

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
    positions: ctx.positions,
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

function selectRefsAndPositions(
  draftActive: boolean,
  dynamicEntities: readonly EntityRef[],
  dynamicPositions: Record<string, Position>,
  view: View,
): { refs: readonly EntityRef[]; positions: Record<string, Position> } {
  const persisted = viewPositions(view);
  if (draftActive) {
    return { refs: dynamicEntities, positions: { ...persisted, ...dynamicPositions } };
  }
  return { refs: entitiesFromView(view), positions: persisted };
}

export const useCanvasNodes = (): Node[] => {
  const { data: components = [] } = useComponents();
  const { currentView, currentViewId } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: capabilities = [] } = useCapabilities();
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { data: acquiredEntities = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: internalTeams = [] } = useInternalTeamsQuery();

  const dynamicViewId = useAppStore((state) => state.dynamicViewId);
  const dynamicEntities = useAppStore((state) => state.dynamicEntities);
  const dynamicPositions = useAppStore((state) => state.dynamicPositions);
  const draftActive = dynamicViewId !== null && dynamicViewId === currentViewId;

  return useMemo(() => {
    if (!currentView) return [];
    const { refs, positions } = selectRefsAndPositions(draftActive, dynamicEntities, dynamicPositions, currentView);
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
    acquiredEntities,
    vendors,
    internalTeams,
    draftActive,
    dynamicEntities,
    dynamicPositions,
  ]);
};

export { entitiesFromView };
