import type { Edge, Node } from '@xyflow/react';
import { useMemo } from 'react';
import type { Relation, View, ViewCapability, ViewComponent } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useRealizations } from '../../capabilities/hooks/useCapabilities';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { useRelations } from '../../relations/hooks/useRelations';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import {
  createOriginRelationshipEdges,
  createParentEdges,
  createRealizationEdges,
  createRelationEdges,
} from '../utils/edgeCreators';
import { isOriginEntity, toNodeId } from '../../../constants/entityIdentifiers';
import type { EntityRef } from '../utils/dynamicMode';

interface ViewProjection {
  components: ViewComponent[];
  capabilities: ViewCapability[];
}

function projectionFromDraft(entities: EntityRef[]): ViewProjection {
  const components: ViewComponent[] = [];
  const capabilities: ViewCapability[] = [];
  for (const ent of entities) {
    if (ent.type === 'component') {
      components.push({ componentId: ent.id as ViewComponent['componentId'], x: 0, y: 0 });
    } else if (ent.type === 'capability') {
      capabilities.push({ capabilityId: ent.id as ViewCapability['capabilityId'], x: 0, y: 0 });
    }
  }
  return { components, capabilities };
}

function projectionFromView(view: View | null): ViewProjection {
  return { components: view?.components ?? [], capabilities: view?.capabilities ?? [] };
}

function selectProjection(
  dynamicEnabled: boolean,
  dynamicEntities: EntityRef[],
  view: View | null,
): ViewProjection {
  return dynamicEnabled ? projectionFromDraft(dynamicEntities) : projectionFromView(view);
}

function relationsBetweenCanvasComponents(relations: Relation[], componentIds: Set<string>): Relation[] {
  return relations.filter(
    (r) => componentIds.has(r.sourceComponentId) && componentIds.has(r.targetComponentId),
  );
}

export const useCanvasEdges = (nodes: Node[]): Edge[] => {
  const { data: relations = [] } = useRelations();
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { currentView } = useCurrentView();
  const { data: capabilities = [] } = useCapabilities();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();
  const { data: capabilityRealizations = [] } = useRealizations();
  const dynamicEnabled = useAppStore((state) => state.dynamicEnabled);
  const dynamicEntities = useAppStore((state) => state.dynamicEntities);

  return useMemo(() => {
    const projection = selectProjection(dynamicEnabled, dynamicEntities, currentView);
    const ctx = {
      nodes,
      selectedEdgeId,
      edgeType: currentView?.edgeType ?? 'default',
      isClassicScheme: (currentView?.colorScheme ?? 'maturity') === 'classic',
    };
    const componentIdsOnCanvas = new Set(projection.components.map((vc) => vc.componentId));
    const originEntityNodeIds = new Set(
      nodes.filter((n) => isOriginEntity(toNodeId(n.id))).map((n) => n.id),
    );

    return [
      ...createRelationEdges(relationsBetweenCanvasComponents(relations, componentIdsOnCanvas), ctx),
      ...createParentEdges(projection.capabilities, capabilities, ctx),
      ...createRealizationEdges(capabilityRealizations, projection.capabilities, projection.components, ctx),
      ...createOriginRelationshipEdges(originRelationships, originEntityNodeIds, componentIdsOnCanvas, ctx),
    ];
  }, [
    relations,
    selectedEdgeId,
    currentView,
    nodes,
    capabilities,
    capabilityRealizations,
    originRelationships,
    dynamicEnabled,
    dynamicEntities,
  ]);
};
