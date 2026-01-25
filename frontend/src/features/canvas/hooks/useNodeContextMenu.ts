import { useState, useCallback, useMemo } from 'react';
import type { Node } from '@xyflow/react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { getOriginEntityTypeFromNodeId, extractOriginEntityId } from '../utils/nodeFactory';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';
import type { Component, Capability, HATEOASLinks } from '../../../api/types';

export interface NodeContextMenu {
  x: number;
  y: number;
  nodeId: string;
  nodeName: string;
  nodeType: 'component' | 'capability' | 'originEntity';
  originEntityType?: OriginEntityType;
  modelLinks?: HATEOASLinks;
  viewElementLinks?: HATEOASLinks;
}

interface MenuPosition {
  x: number;
  y: number;
}

interface OriginEntityLookups {
  acquiredEntities: { id: string; name: string; _links?: HATEOASLinks }[];
  vendors: { id: string; name: string; _links?: HATEOASLinks }[];
  internalTeams: { id: string; name: string; _links?: HATEOASLinks }[];
}

function findOriginEntity(
  nodeId: string,
  lookups: OriginEntityLookups
): { entity: { name: string; _links?: HATEOASLinks }; originEntityType: OriginEntityType } | null {
  const originEntityType = getOriginEntityTypeFromNodeId(nodeId);
  const entityId = extractOriginEntityId(nodeId);
  if (!originEntityType || !entityId) return null;

  const lookupMap: Record<OriginEntityType, typeof lookups.acquiredEntities> = {
    acquired: lookups.acquiredEntities,
    vendor: lookups.vendors,
    team: lookups.internalTeams,
  };

  const entity = lookupMap[originEntityType].find((e) => e.id === entityId);
  return entity ? { entity, originEntityType } : null;
}

function resolveOriginEntityNode(
  node: Node,
  lookups: OriginEntityLookups,
  position: MenuPosition,
  viewElementLinks?: HATEOASLinks
): NodeContextMenu | null {
  const result = findOriginEntity(node.id, lookups);
  if (!result) return null;

  return {
    ...position,
    nodeId: node.id,
    nodeName: result.entity.name,
    nodeType: 'originEntity',
    originEntityType: result.originEntityType,
    modelLinks: result.entity._links,
    viewElementLinks,
  };
}

function resolveCapabilityNode(
  node: Node,
  capabilities: Capability[],
  position: MenuPosition,
  viewElementLinks?: HATEOASLinks
): NodeContextMenu | null {
  const capId = node.id.replace('cap-', '');
  const capability = capabilities.find((c) => c.id === capId);
  if (!capability) return null;

  return {
    ...position,
    nodeId: capId,
    nodeName: capability.name,
    nodeType: 'capability',
    modelLinks: capability._links,
    viewElementLinks,
  };
}

function resolveComponentNode(
  node: Node,
  components: Component[],
  position: MenuPosition,
  viewElementLinks?: HATEOASLinks
): NodeContextMenu | null {
  const component = components.find((c) => c.id === node.id);
  if (!component) return null;

  return {
    ...position,
    nodeId: node.id,
    nodeName: component.name,
    nodeType: 'component',
    modelLinks: component._links,
    viewElementLinks,
  };
}

interface NodeContextMenuDependencies {
  components: Component[];
  capabilities: Capability[];
  originEntityLookups: OriginEntityLookups;
  currentViewComponents: { componentId: string; _links?: HATEOASLinks }[];
  currentViewCapabilities: { capabilityId: string; _links?: HATEOASLinks }[];
  currentViewOriginEntities: { originEntityId: string; _links?: HATEOASLinks }[];
}

function resolveNodeMenu(
  node: Node,
  position: MenuPosition,
  deps: NodeContextMenuDependencies
): NodeContextMenu | null {
  if (node.type === 'capability') {
    const capId = node.id.replace('cap-', '');
    const viewElement = deps.currentViewCapabilities.find((vc) => vc.capabilityId === capId);
    return resolveCapabilityNode(node, deps.capabilities, position, viewElement?._links);
  }

  if (node.type === 'originEntity') {
    const viewElement = deps.currentViewOriginEntities.find((vo) => vo.originEntityId === node.id);
    return resolveOriginEntityNode(node, deps.originEntityLookups, position, viewElement?._links);
  }

  const viewElement = deps.currentViewComponents.find((vc) => vc.componentId === node.id);
  return resolveComponentNode(node, deps.components, position, viewElement?._links);
}

export const useNodeContextMenu = () => {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: acquiredEntities = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: internalTeams = [] } = useInternalTeamsQuery();
  const { currentView } = useCurrentView();

  const [nodeContextMenu, setNodeContextMenu] = useState<NodeContextMenu | null>(null);

  const deps = useMemo<NodeContextMenuDependencies>(() => ({
    components,
    capabilities,
    originEntityLookups: { acquiredEntities, vendors, internalTeams },
    currentViewComponents: currentView?.components ?? [],
    currentViewCapabilities: currentView?.capabilities ?? [],
    currentViewOriginEntities: currentView?.originEntities ?? [],
  }), [components, capabilities, acquiredEntities, vendors, internalTeams, currentView?.components, currentView?.capabilities, currentView?.originEntities]);

  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      event.preventDefault();
      const position: MenuPosition = { x: event.clientX, y: event.clientY };
      const menu = resolveNodeMenu(node, position, deps);
      if (menu) {
        setNodeContextMenu(menu);
      }
    },
    [deps]
  );

  const closeNodeMenu = useCallback(() => {
    setNodeContextMenu(null);
  }, []);

  return {
    nodeContextMenu,
    onNodeContextMenu,
    closeNodeMenu,
  };
};
