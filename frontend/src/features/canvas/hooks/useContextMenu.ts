import { useState, useCallback, useMemo } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useCapabilities, useRealizationsForComponents } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useRelations } from '../../relations/hooks/useRelations';
import type { Component, Capability, Relation, CapabilityRealization, ComponentId } from '../../../api/types';

export interface NodeContextMenu {
  x: number;
  y: number;
  nodeId: string;
  nodeName: string;
  nodeType: 'component' | 'capability';
}

export interface EdgeContextMenu {
  x: number;
  y: number;
  edgeId: string;
  edgeName: string;
  edgeType: 'relation' | 'parent' | 'realization';
  realizationId?: string;
  isInherited?: boolean;
}

interface MenuPosition {
  x: number;
  y: number;
}

interface EdgeLookupDependencies {
  relations: Relation[];
  capabilityRealizations: CapabilityRealization[];
  capabilities: Capability[];
  components: Component[];
}

function resolveCapabilityNode(
  node: Node,
  capabilities: Capability[],
  position: MenuPosition
): NodeContextMenu | null {
  const capId = node.id.replace('cap-', '');
  const capability = capabilities.find((c) => c.id === capId);
  if (!capability) return null;

  return {
    ...position,
    nodeId: capId,
    nodeName: capability.name,
    nodeType: 'capability',
  };
}

function resolveComponentNode(
  node: Node,
  components: Component[],
  position: MenuPosition
): NodeContextMenu | null {
  const component = components.find((c) => c.id === node.id);
  if (!component) return null;

  return {
    ...position,
    nodeId: node.id,
    nodeName: component.name,
    nodeType: 'component',
  };
}

function resolveParentEdge(edge: Edge, position: MenuPosition): EdgeContextMenu {
  return {
    ...position,
    edgeId: edge.id,
    edgeName: 'Parent',
    edgeType: 'parent',
  };
}

function resolveRealizationEdge(
  edge: Edge,
  deps: EdgeLookupDependencies,
  position: MenuPosition
): EdgeContextMenu | null {
  const realizationId = edge.id.replace('realization-', '');
  const realization = deps.capabilityRealizations.find((r) => r.id === realizationId);
  if (!realization) return null;

  const capability = deps.capabilities.find((c) => c.id === realization.capabilityId);
  const component = deps.components.find((c) => c.id === realization.componentId);
  const edgeName = `${capability?.name || 'Capability'} -> ${component?.name || 'Component'}`;

  return {
    ...position,
    edgeId: edge.id,
    edgeName,
    edgeType: 'realization',
    realizationId,
    isInherited: realization.origin === 'Inherited',
  };
}

function resolveRelationEdge(
  edge: Edge,
  relations: Relation[],
  position: MenuPosition
): EdgeContextMenu | null {
  const relation = relations.find((r) => r.id === edge.id);
  if (!relation) return null;

  return {
    ...position,
    edgeId: edge.id,
    edgeName: relation.name || relation.relationType,
    edgeType: 'relation',
  };
}

function resolveEdgeContextMenu(
  edge: Edge,
  deps: EdgeLookupDependencies,
  position: MenuPosition
): EdgeContextMenu | null {
  if (edge.id.startsWith('parent-')) {
    return resolveParentEdge(edge, position);
  }
  if (edge.id.startsWith('realization-')) {
    return resolveRealizationEdge(edge, deps, position);
  }
  return resolveRelationEdge(edge, deps.relations, position);
}

export const useContextMenu = () => {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: relations = [] } = useRelations();
  const { currentView } = useCurrentView();

  const componentIdsInView = useMemo(() =>
    currentView?.components.map((vc) => vc.componentId as ComponentId) || [],
    [currentView?.components]
  );
  const { data: capabilityRealizations = [] } = useRealizationsForComponents(componentIdsInView);

  const [nodeContextMenu, setNodeContextMenu] = useState<NodeContextMenu | null>(null);
  const [edgeContextMenu, setEdgeContextMenu] = useState<EdgeContextMenu | null>(null);

  const edgeLookupDeps = useMemo<EdgeLookupDependencies>(
    () => ({ relations, capabilityRealizations, capabilities, components }),
    [relations, capabilityRealizations, capabilities, components]
  );

  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      event.preventDefault();
      const position: MenuPosition = { x: event.clientX, y: event.clientY };

      const menu =
        node.type === 'capability'
          ? resolveCapabilityNode(node, capabilities, position)
          : resolveComponentNode(node, components, position);

      if (menu) {
        setNodeContextMenu(menu);
      }
    },
    [components, capabilities]
  );

  const onEdgeContextMenu = useCallback(
    (event: React.MouseEvent, edge: Edge) => {
      event.preventDefault();
      const position: MenuPosition = { x: event.clientX, y: event.clientY };

      const menu = resolveEdgeContextMenu(edge, edgeLookupDeps, position);
      if (menu) {
        setEdgeContextMenu(menu);
      }
    },
    [edgeLookupDeps]
  );

  const closeMenus = useCallback(() => {
    setNodeContextMenu(null);
    setEdgeContextMenu(null);
  }, []);

  return {
    nodeContextMenu,
    edgeContextMenu,
    onNodeContextMenu,
    onEdgeContextMenu,
    closeMenus,
  };
};
