import { useState, useCallback, useMemo } from 'react';
import type { Edge } from '@xyflow/react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useCapabilities, useRealizationsForComponents } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useRelations } from '../../relations/hooks/useRelations';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { ORIGIN_RELATIONSHIP_LABELS } from '../../../constants/entityIdentifiers';
import type {
  Component,
  Capability,
  Relation,
  CapabilityRealization,
  ComponentId,
  CapabilityId,
  HATEOASLinks,
  OriginRelationship,
  OriginRelationshipId,
  OriginRelationshipType,
} from '../../../api/types';

export interface EdgeContextMenu {
  x: number;
  y: number;
  edgeId: string;
  edgeName: string;
  edgeType: 'relation' | 'parent' | 'realization' | 'origin-relationship';
  realizationId?: string;
  capabilityId?: CapabilityId;
  componentId?: ComponentId;
  isInherited?: boolean;
  originRelationshipId?: OriginRelationshipId;
  originRelationshipType?: OriginRelationshipType;
  originEntityId?: string;
  _links?: HATEOASLinks;
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
  originRelationships: OriginRelationship[];
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
    capabilityId: realization.capabilityId,
    componentId: realization.componentId,
    isInherited: realization.origin === 'Inherited',
    _links: realization._links,
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
    _links: relation._links,
  };
}

function resolveOriginRelationshipEdge(
  edge: Edge,
  originRelationships: OriginRelationship[],
  components: Component[],
  position: MenuPosition
): EdgeContextMenu | null {
  const relationshipId = edge.id.replace('origin-', '');
  const relationship = originRelationships.find((r) => r.id === relationshipId);
  if (!relationship) return null;

  const component = components.find((c) => c.id === relationship.componentId);
  const label = ORIGIN_RELATIONSHIP_LABELS[relationship.relationshipType];
  const edgeName = `${component?.name || 'Component'} ${label} ${relationship.originEntityName}`;

  return {
    ...position,
    edgeId: edge.id,
    edgeName,
    edgeType: 'origin-relationship',
    originRelationshipId: relationship.id,
    originRelationshipType: relationship.relationshipType,
    componentId: relationship.componentId,
    originEntityId: relationship.originEntityId,
    _links: relationship._links,
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
  if (edge.id.startsWith('origin-')) {
    return resolveOriginRelationshipEdge(edge, deps.originRelationships, deps.components, position);
  }
  return resolveRelationEdge(edge, deps.relations, position);
}

export const useEdgeContextMenu = () => {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: relations = [] } = useRelations();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();
  const { currentView } = useCurrentView();

  const componentIdsInView = useMemo(
    () => currentView?.components.map((vc) => vc.componentId) || [],
    [currentView?.components]
  );
  const { data: capabilityRealizations = [] } = useRealizationsForComponents(componentIdsInView);

  const [edgeContextMenu, setEdgeContextMenu] = useState<EdgeContextMenu | null>(null);

  const edgeLookupDeps = useMemo<EdgeLookupDependencies>(
    () => ({ relations, capabilityRealizations, capabilities, components, originRelationships }),
    [relations, capabilityRealizations, capabilities, components, originRelationships]
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

  const closeEdgeMenu = useCallback(() => {
    setEdgeContextMenu(null);
  }, []);

  return {
    edgeContextMenu,
    onEdgeContextMenu,
    closeEdgeMenu,
  };
};
