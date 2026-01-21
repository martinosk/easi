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

    const componentNodes = components
      .filter((component) => isComponentInView(component, currentView))
      .map((component) => createComponentNode(component, currentView, layoutPositions, selectedNodeId));

    const capabilityNodes = (currentView.capabilities || [])
      .map((vc: ViewCapability) => {
        const capability = capabilities.find((c) => c.id === vc.capabilityId);
        if (!capability) return null;

        return createCapabilityNode(vc.capabilityId, capability, layoutPositions, vc, selectedCapabilityId);
      })
      .filter((n): n is Node => n !== null);

    const acquiredEntityNodes = acquiredEntities
      .filter((entity) => {
        const nodeId = `${ORIGIN_ENTITY_PREFIXES.acquired}${entity.id}`;
        return layoutPositions[nodeId] !== undefined;
      })
      .map((entity) => createAcquiredEntityNode(entity, layoutPositions, selectedNodeId));

    const vendorNodes = vendors
      .filter((vendor) => {
        const nodeId = `${ORIGIN_ENTITY_PREFIXES.vendor}${vendor.id}`;
        return layoutPositions[nodeId] !== undefined;
      })
      .map((vendor) => createVendorNode(vendor, layoutPositions, selectedNodeId));

    const internalTeamNodes = internalTeams
      .filter((team) => {
        const nodeId = `${ORIGIN_ENTITY_PREFIXES.team}${team.id}`;
        return layoutPositions[nodeId] !== undefined;
      })
      .map((team) => createInternalTeamNode(team, layoutPositions, selectedNodeId));

    return [...componentNodes, ...capabilityNodes, ...acquiredEntityNodes, ...vendorNodes, ...internalTeamNodes];
  }, [components, currentView, selectedNodeId, capabilities, selectedCapabilityId, layoutPositions, acquiredEntities, vendors, internalTeams]);
};
