import type { Edge, Node } from '@xyflow/react';
import { MarkerType } from '@xyflow/react';
import { getBestHandles } from './handleCalculation';
import type { Capability, CapabilityRealization, Relation, ViewCapability, ViewComponent } from '../../../api/types';

export interface EdgeCreationContext {
  nodes: Node[];
  selectedEdgeId: string | null;
  edgeType: string;
  isClassicScheme: boolean;
}

export function createRelationEdges(
  relations: Relation[],
  ctx: EdgeCreationContext
): Edge[] {
  return relations.map((relation) => {
    const isSelected = ctx.selectedEdgeId === relation.id;
    const isTriggers = relation.relationType === 'Triggers';

    const sourceNode = ctx.nodes.find(n => n.id === relation.sourceComponentId);
    const targetNode = ctx.nodes.find(n => n.id === relation.targetComponentId);
    const { sourceHandle, targetHandle } = getBestHandles(sourceNode, targetNode);

    const edgeColor = ctx.isClassicScheme ? '#000000' : (isTriggers ? '#f97316' : '#3b82f6');

    return {
      id: relation.id,
      source: relation.sourceComponentId,
      target: relation.targetComponentId,
      sourceHandle,
      targetHandle,
      label: relation.name || relation.relationType,
      type: ctx.edgeType,
      animated: isSelected,
      style: { stroke: edgeColor, strokeWidth: isSelected ? 3 : 2 },
      markerEnd: { type: MarkerType.ArrowClosed, color: edgeColor },
      labelStyle: { fill: edgeColor, fontWeight: isSelected ? 700 : 500 },
      labelBgStyle: { fill: '#ffffff' },
    };
  });
}

export function createParentEdges(
  viewCapabilities: ViewCapability[],
  capabilities: Capability[],
  ctx: EdgeCreationContext
): Edge[] {
  const canvasCapabilityIds = new Set(viewCapabilities.map((vc) => vc.capabilityId));

  return viewCapabilities
    .map((vc): Edge | null => {
      const capability = capabilities.find((c) => c.id === vc.capabilityId);
      if (!capability?.parentId || !canvasCapabilityIds.has(capability.parentId)) {
        return null;
      }

      const childNodeId = `cap-${capability.id}`;
      const parentNodeId = `cap-${capability.parentId}`;
      const edgeId = `parent-${capability.parentId}-${capability.id}`;
      const isSelected = ctx.selectedEdgeId === edgeId;

      const parentNode = ctx.nodes.find((n) => n.id === parentNodeId);
      const childNode = ctx.nodes.find((n) => n.id === childNodeId);
      const { sourceHandle, targetHandle } = getBestHandles(parentNode, childNode);

      const parentEdgeColor = ctx.isClassicScheme ? '#000000' : '#374151';

      return {
        id: edgeId,
        source: parentNodeId,
        target: childNodeId,
        sourceHandle,
        targetHandle,
        label: 'Parent',
        type: ctx.edgeType,
        animated: isSelected,
        style: { stroke: parentEdgeColor, strokeWidth: 3 },
        markerEnd: { type: MarkerType.ArrowClosed, color: parentEdgeColor },
        labelStyle: { fill: parentEdgeColor, fontWeight: isSelected ? 700 : 600 },
        labelBgStyle: { fill: '#ffffff' },
      };
    })
    .filter((e): e is Edge => e !== null);
}

interface RealizationVisibility {
  visibleCapabilityIds: Set<string>;
  componentIdsOnCanvas: Set<string>;
  allRealizations: CapabilityRealization[];
}

function isRealizationVisible(
  realization: CapabilityRealization,
  visibility: RealizationVisibility
): boolean {
  const { visibleCapabilityIds, componentIdsOnCanvas, allRealizations } = visibility;

  if (!componentIdsOnCanvas.has(realization.componentId)) return false;
  if (!visibleCapabilityIds.has(realization.capabilityId)) return false;
  if (realization.origin === 'Direct') return true;

  if (realization.origin === 'Inherited' && realization.sourceRealizationId) {
    const source = allRealizations.find((r) => r.id === realization.sourceRealizationId);
    return source ? !visibleCapabilityIds.has(source.capabilityId) : false;
  }
  return false;
}

function buildRealizationEdge(
  realization: CapabilityRealization,
  ctx: EdgeCreationContext
): Edge {
  const edgeId = `realization-${realization.id}`;
  const isSelected = ctx.selectedEdgeId === edgeId;
  const isInherited = realization.origin === 'Inherited';
  const realizationColor = ctx.isClassicScheme ? '#000000' : '#10B981';

  const sourceNode = ctx.nodes.find((n) => n.id === realization.componentId);
  const targetNode = ctx.nodes.find((n) => n.id === `cap-${realization.capabilityId}`);
  const { sourceHandle, targetHandle } = getBestHandles(sourceNode, targetNode);

  return {
    id: edgeId,
    source: realization.componentId,
    target: `cap-${realization.capabilityId}`,
    sourceHandle,
    targetHandle,
    label: isInherited ? 'Realizes (inherited)' : 'Realizes',
    type: ctx.edgeType,
    animated: isSelected,
    className: isInherited ? 'realization-edge inherited' : 'realization-edge',
    style: {
      stroke: realizationColor,
      strokeWidth: isSelected ? 3 : 2,
      strokeDasharray: '5,5',
      opacity: isInherited ? 0.6 : 1.0,
    },
    markerEnd: { type: MarkerType.ArrowClosed, color: realizationColor },
    labelStyle: {
      fill: realizationColor,
      fontWeight: isSelected ? 700 : 500,
      opacity: isInherited ? 0.8 : 1.0,
    },
    labelBgStyle: { fill: '#ffffff' },
  };
}

export function createRealizationEdges(
  capabilityRealizations: CapabilityRealization[],
  viewCapabilities: ViewCapability[],
  viewComponents: ViewComponent[],
  ctx: EdgeCreationContext
): Edge[] {
  const visibility: RealizationVisibility = {
    visibleCapabilityIds: new Set(viewCapabilities.map((vc) => vc.capabilityId)),
    componentIdsOnCanvas: new Set(viewComponents.map((vc) => vc.componentId)),
    allRealizations: capabilityRealizations,
  };

  return capabilityRealizations
    .filter((r) => isRealizationVisible(r, visibility))
    .map((r) => buildRealizationEdge(r, ctx));
}
