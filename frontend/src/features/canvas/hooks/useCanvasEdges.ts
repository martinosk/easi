import { useMemo } from 'react';
import type { Edge, Node } from '@xyflow/react';
import { MarkerType } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { getBestHandles } from '../utils/handleCalculation';
import type { CapabilityRealization } from '../../../api/types';

export const useCanvasEdges = (nodes: Node[]): Edge[] => {
  const relations = useAppStore((state) => state.relations);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const currentView = useAppStore((state) => state.currentView);
  const canvasCapabilities = useAppStore((state) => state.canvasCapabilities);
  const capabilities = useAppStore((state) => state.capabilities);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);

  return useMemo(() => {
    const edgeType = currentView?.edgeType || 'default';
    const colorScheme = currentView?.colorScheme || 'maturity';
    const isClassicScheme = colorScheme === 'classic';

    const relationEdges: Edge[] = relations.map((relation) => {
      const isSelected = selectedEdgeId === relation.id;
      const isTriggers = relation.relationType === 'Triggers';

      const sourceNode = nodes.find(n => n.id === relation.sourceComponentId);
      const targetNode = nodes.find(n => n.id === relation.targetComponentId);
      const { sourceHandle, targetHandle } = getBestHandles(sourceNode, targetNode);

      const edgeColor = isClassicScheme ? '#000000' : (isTriggers ? '#f97316' : '#3b82f6');

      return {
        id: relation.id,
        source: relation.sourceComponentId,
        target: relation.targetComponentId,
        sourceHandle,
        targetHandle,
        label: relation.name || relation.relationType,
        type: edgeType,
        animated: isSelected,
        style: {
          stroke: edgeColor,
          strokeWidth: isSelected ? 3 : 2,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: edgeColor,
        },
        labelStyle: {
          fill: edgeColor,
          fontWeight: isSelected ? 700 : 500,
        },
        labelBgStyle: {
          fill: '#ffffff',
        },
      };
    });

    const canvasCapabilityIds = new Set(canvasCapabilities.map((cc) => cc.capabilityId));
    const parentEdges: Edge[] = canvasCapabilities
      .map((cc) => {
        const capability = capabilities.find((c) => c.id === cc.capabilityId);
        if (!capability || !capability.parentId) return null;

        if (!canvasCapabilityIds.has(capability.parentId)) return null;

        const childNodeId = `cap-${capability.id}`;
        const parentNodeId = `cap-${capability.parentId}`;
        const edgeId = `parent-${capability.parentId}-${capability.id}`;
        const isSelected = selectedEdgeId === edgeId;

        const parentNode = nodes.find((n) => n.id === parentNodeId);
        const childNode = nodes.find((n) => n.id === childNodeId);
        const { sourceHandle, targetHandle } = getBestHandles(parentNode, childNode);

        const parentEdgeColor = isClassicScheme ? '#000000' : '#374151';

        return {
          id: edgeId,
          source: parentNodeId,
          target: childNodeId,
          sourceHandle,
          targetHandle,
          label: 'Parent',
          type: edgeType,
          animated: isSelected,
          style: {
            stroke: parentEdgeColor,
            strokeWidth: 3,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: parentEdgeColor,
          },
          labelStyle: {
            fill: parentEdgeColor,
            fontWeight: isSelected ? 700 : 600,
          },
          labelBgStyle: {
            fill: '#ffffff',
          },
        };
      })
      .filter((e) => e !== null) as Edge[];

    const visibleCapabilityIds = new Set(canvasCapabilities.map((cc) => cc.capabilityId));
    const componentIdsOnCanvas = new Set(
      currentView?.components.map((vc) => vc.componentId) || []
    );

    const shouldShowRealizationEdge = (realization: CapabilityRealization): boolean => {
      if (!componentIdsOnCanvas.has(realization.componentId)) return false;
      if (!visibleCapabilityIds.has(realization.capabilityId)) return false;

      if (realization.origin === 'Direct') {
        return true;
      }

      if (realization.origin === 'Inherited' && realization.sourceRealizationId) {
        const sourceRealization = capabilityRealizations.find(
          (r) => r.id === realization.sourceRealizationId
        );
        if (sourceRealization) {
          return !visibleCapabilityIds.has(sourceRealization.capabilityId);
        }
      }
      return false;
    };

    const realizationEdges = capabilityRealizations
      .filter(shouldShowRealizationEdge)
      .map((realization) => {
        const edgeId = `realization-${realization.id}`;
        const isSelected = selectedEdgeId === edgeId;
        const isInherited = realization.origin === 'Inherited';

        const sourceNodeId = realization.componentId;
        const targetNodeId = `cap-${realization.capabilityId}`;

        const sourceNode = nodes.find((n) => n.id === sourceNodeId);
        const targetNode = nodes.find((n) => n.id === targetNodeId);
        const { sourceHandle, targetHandle } = getBestHandles(sourceNode, targetNode);

        const realizationColor = isClassicScheme ? '#000000' : '#10B981';

        return {
          id: edgeId,
          source: sourceNodeId,
          target: targetNodeId,
          sourceHandle,
          targetHandle,
          label: isInherited ? 'Realizes (inherited)' : 'Realizes',
          type: edgeType,
          animated: isSelected,
          className: isInherited ? 'realization-edge inherited' : 'realization-edge',
          style: {
            stroke: realizationColor,
            strokeWidth: isSelected ? 3 : 2,
            strokeDasharray: '5,5',
            opacity: isInherited ? 0.6 : 1.0,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: realizationColor,
          },
          labelStyle: {
            fill: realizationColor,
            fontWeight: isSelected ? 700 : 500,
            opacity: isInherited ? 0.8 : 1.0,
          },
          labelBgStyle: {
            fill: '#ffffff',
          },
        };
      });

    return [...relationEdges, ...parentEdges, ...realizationEdges];
  }, [relations, selectedEdgeId, currentView?.edgeType, currentView?.colorScheme, currentView?.components, nodes, canvasCapabilities, capabilities, capabilityRealizations]);
};
