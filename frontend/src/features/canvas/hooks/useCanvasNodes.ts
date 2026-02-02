import { useMemo } from 'react';
import type { Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import {
  createComponentNode,
  createCapabilityNode,
  createAcquiredEntityNode,
  createVendorNode,
  createInternalTeamNode,
  isComponentInView,
  ORIGIN_ENTITY_PREFIXES,
} from '../utils/nodeFactory';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import type { ViewCapability } from '../../../api/types';

function isOriginEntityInView(
  entityId: string,
  prefix: string,
  viewOriginEntityIds: Set<string>
): boolean {
  const nodeId = `${prefix}${entityId}`;
  return viewOriginEntityIds.has(nodeId);
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

  return useMemo(() => {
    if (!currentView) return [];

    const viewOriginEntityIds = new Set((currentView.originEntities || []).map((oe) => oe.originEntityId));
    const viewOriginEntityPositions: Record<string, { x: number; y: number }> = {};
    for (const oe of currentView.originEntities || []) {
      viewOriginEntityPositions[oe.originEntityId] = { x: oe.x, y: oe.y };
    }

    const componentNodes = components
      .filter((component) => isComponentInView(component, currentView))
      .map((component) => createComponentNode(component, currentView, layoutPositions, selectedNodeId));

    const capabilityNodes = (currentView.capabilities || [])
      .map((vc: ViewCapability) => {
        const capability = capabilities.find((c) => c.id === vc.capabilityId);
        return capability ? createCapabilityNode({ capabilityId: vc.capabilityId, capability, layoutPositions, viewCapability: vc, selectedCapabilityId }) : null;
      })
      .filter((n): n is Node => n !== null);

    const acquiredEntityNodes = acquiredEntities
      .filter((e) => isOriginEntityInView(e.id, ORIGIN_ENTITY_PREFIXES.acquired, viewOriginEntityIds))
      .map((entity) => createAcquiredEntityNode(entity, viewOriginEntityPositions, selectedNodeId));

    const vendorNodes = vendors
      .filter((v) => isOriginEntityInView(v.id, ORIGIN_ENTITY_PREFIXES.vendor, viewOriginEntityIds))
      .map((vendor) => createVendorNode(vendor, viewOriginEntityPositions, selectedNodeId));

    const internalTeamNodes = internalTeams
      .filter((t) => isOriginEntityInView(t.id, ORIGIN_ENTITY_PREFIXES.team, viewOriginEntityIds))
      .map((team) => createInternalTeamNode(team, viewOriginEntityPositions, selectedNodeId));

    return [...componentNodes, ...capabilityNodes, ...acquiredEntityNodes, ...vendorNodes, ...internalTeamNodes];
  }, [components, currentView, selectedNodeId, capabilities, selectedCapabilityId, layoutPositions, acquiredEntities, vendors, internalTeams]);
};
