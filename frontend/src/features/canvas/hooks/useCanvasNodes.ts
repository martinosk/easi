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
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import type { ViewCapability, OriginRelationship } from '../../../api/types';

type RelationshipTypeMapping = Record<string, 'acquired' | 'vendor' | 'team'>;

const RELATIONSHIP_TYPE_MAP: RelationshipTypeMapping = {
  AcquiredVia: 'acquired',
  PurchasedFrom: 'vendor',
  BuiltBy: 'team',
};

function categorizeOriginRelationships(
  relationships: OriginRelationship[],
  componentIdsOnCanvas: Set<string>
): Record<'acquired' | 'vendor' | 'team', Set<string>> {
  const result = { acquired: new Set<string>(), vendor: new Set<string>(), team: new Set<string>() };
  for (const rel of relationships) {
    const category = RELATIONSHIP_TYPE_MAP[rel.relationshipType];
    if (category && componentIdsOnCanvas.has(rel.componentId)) {
      result[category].add(rel.originEntityId);
    }
  }
  return result;
}

function shouldIncludeOriginEntity(
  entityId: string,
  prefix: string,
  layoutPositions: Record<string, unknown>,
  entitiesWithRelationships: Set<string>
): boolean {
  const nodeId = `${prefix}${entityId}`;
  return layoutPositions[nodeId] !== undefined || entitiesWithRelationships.has(entityId);
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
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  return useMemo(() => {
    if (!currentView) return [];

    const componentIdsOnCanvas = new Set(currentView.components.map((vc) => vc.componentId));
    const originEntitiesWithRelationships = categorizeOriginRelationships(originRelationships, componentIdsOnCanvas);

    const componentNodes = components
      .filter((component) => isComponentInView(component, currentView))
      .map((component) => createComponentNode(component, currentView, layoutPositions, selectedNodeId));

    const capabilityNodes = (currentView.capabilities || [])
      .map((vc: ViewCapability) => {
        const capability = capabilities.find((c) => c.id === vc.capabilityId);
        return capability ? createCapabilityNode(vc.capabilityId, capability, layoutPositions, vc, selectedCapabilityId) : null;
      })
      .filter((n): n is Node => n !== null);

    const acquiredEntityNodes = acquiredEntities
      .filter((e) => shouldIncludeOriginEntity(e.id, ORIGIN_ENTITY_PREFIXES.acquired, layoutPositions, originEntitiesWithRelationships.acquired))
      .map((entity) => createAcquiredEntityNode(entity, layoutPositions, selectedNodeId));

    const vendorNodes = vendors
      .filter((v) => shouldIncludeOriginEntity(v.id, ORIGIN_ENTITY_PREFIXES.vendor, layoutPositions, originEntitiesWithRelationships.vendor))
      .map((vendor) => createVendorNode(vendor, layoutPositions, selectedNodeId));

    const internalTeamNodes = internalTeams
      .filter((t) => shouldIncludeOriginEntity(t.id, ORIGIN_ENTITY_PREFIXES.team, layoutPositions, originEntitiesWithRelationships.team))
      .map((team) => createInternalTeamNode(team, layoutPositions, selectedNodeId));

    return [...componentNodes, ...capabilityNodes, ...acquiredEntityNodes, ...vendorNodes, ...internalTeamNodes];
  }, [components, currentView, selectedNodeId, capabilities, selectedCapabilityId, layoutPositions, acquiredEntities, vendors, internalTeams, originRelationships]);
};
