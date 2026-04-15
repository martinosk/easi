import type { Edge, Node } from '@xyflow/react';
import { useMemo } from 'react';
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

export const useCanvasEdges = (nodes: Node[]): Edge[] => {
  const { data: relations = [] } = useRelations();
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { currentView } = useCurrentView();
  const { data: capabilities = [] } = useCapabilities();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  const { data: capabilityRealizations = [] } = useRealizations();

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
      nodes.filter((n) => isOriginEntity(toNodeId(n.id))).map((n) => n.id),
    );

    return [
      ...createRelationEdges(relations, ctx),
      ...createParentEdges(viewCapabilities, capabilities, ctx),
      ...createRealizationEdges(capabilityRealizations, viewCapabilities, viewComponents, ctx),
      ...createOriginRelationshipEdges(originRelationships, originEntityNodeIds, componentIdsOnCanvas, ctx),
    ];
  }, [relations, selectedEdgeId, currentView, nodes, capabilities, capabilityRealizations, originRelationships]);
};
