import { useMemo } from 'react';
import type { Edge, Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useRealizationsForComponents } from '../../capabilities/hooks/useCapabilities';
import { useRelations } from '../../relations/hooks/useRelations';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { createRelationEdges, createParentEdges, createRealizationEdges, createOriginRelationshipEdges } from '../utils/edgeCreators';
import { ORIGIN_ENTITY_PREFIXES } from '../utils/nodeFactory';
import type { ViewComponent } from '../../../api/types';

export const useCanvasEdges = (nodes: Node[]): Edge[] => {
  const { data: relations = [] } = useRelations();
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { currentView } = useCurrentView();
  const { data: capabilities = [] } = useCapabilities();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  const componentIdsInView = useMemo(
    () => (currentView?.components ?? []).map((vc: ViewComponent) => vc.componentId),
    [currentView?.components]
  );
  const { data: capabilityRealizations = [] } = useRealizationsForComponents(componentIdsInView);

  return useMemo(() => {
    const viewCapabilities = currentView?.capabilities ?? [];
    const viewComponents = currentView?.components ?? [];

    const ctx = {
      nodes,
      selectedEdgeId,
      edgeType: currentView?.edgeType ?? 'default',
      isClassicScheme: (currentView?.colorScheme ?? 'maturity') === 'classic',
    };

    const componentIdsOnCanvas = new Set(viewComponents.map((vc) => vc.componentId));
    const originEntityNodeIds = new Set(
      nodes
        .filter((n) =>
          n.id.startsWith(ORIGIN_ENTITY_PREFIXES.acquired) ||
          n.id.startsWith(ORIGIN_ENTITY_PREFIXES.vendor) ||
          n.id.startsWith(ORIGIN_ENTITY_PREFIXES.team)
        )
        .map((n) => n.id)
    );

    return [
      ...createRelationEdges(relations, ctx),
      ...createParentEdges(viewCapabilities, capabilities, ctx),
      ...createRealizationEdges(capabilityRealizations, viewCapabilities, viewComponents, ctx),
      ...createOriginRelationshipEdges(originRelationships, originEntityNodeIds, componentIdsOnCanvas, ctx),
    ];
  }, [relations, selectedEdgeId, currentView, nodes, capabilities, capabilityRealizations, originRelationships]);
};
