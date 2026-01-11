import { useMemo } from 'react';
import type { Edge, Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useRealizationsForComponents } from '../../capabilities/hooks/useCapabilities';
import { useRelations } from '../../relations/hooks/useRelations';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { createRelationEdges, createParentEdges, createRealizationEdges } from '../utils/edgeCreators';
import type { ComponentId, ViewComponent } from '../../../api/types';

export const useCanvasEdges = (nodes: Node[]): Edge[] => {
  const { data: relations = [] } = useRelations();
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { currentView } = useCurrentView();
  const { data: capabilities = [] } = useCapabilities();

  const componentIdsInView = useMemo(
    () => (currentView?.components ?? []).map((vc: ViewComponent) => vc.componentId as ComponentId),
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

    return [
      ...createRelationEdges(relations, ctx),
      ...createParentEdges(viewCapabilities, capabilities, ctx),
      ...createRealizationEdges(capabilityRealizations, viewCapabilities, viewComponents, ctx),
    ];
  }, [relations, selectedEdgeId, currentView, nodes, capabilities, capabilityRealizations]);
};
