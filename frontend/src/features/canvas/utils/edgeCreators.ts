import type { Edge, Node } from '@xyflow/react';
import { MarkerType } from '@xyflow/react';
import type {
  Capability,
  CapabilityRealization,
  OriginRelationship,
  OriginRelationshipType,
  Relation,
  ViewCapability,
  ViewComponent,
} from '../../../api/types';
import {
  makeNodeId,
  ORIGIN_RELATIONSHIP_LABELS,
  RELATIONSHIP_TO_ENTITY_TYPE,
} from '../../../constants/entityIdentifiers';
import { resolveToken } from '../../../theme/resolveToken';
import { getBestHandles } from './handleCalculation';

const CLASSIC_EDGE_COLOR = '#000000';

const edgeLabelBg = () => resolveToken('--surface', '#FFFFFF');

export interface EdgeCreationContext {
  nodes: Node[];
  selectedEdgeId: string | null;
  edgeType: string;
  isClassicScheme: boolean;
}

function resolveHandles(ctx: EdgeCreationContext, sourceNodeId: string, targetNodeId: string) {
  const sourceNode = ctx.nodes.find((n) => n.id === sourceNodeId);
  const targetNode = ctx.nodes.find((n) => n.id === targetNodeId);
  return getBestHandles(sourceNode, targetNode);
}

interface EdgeVisualOptions {
  color: string;
  isSelected: boolean;
  selectedStrokeWidth?: number;
  unselectedStrokeWidth?: number;
  unselectedFontWeight?: number;
  extraStyle?: Record<string, string | number>;
  extraLabelStyle?: Record<string, string | number>;
}

function buildEdgeVisuals({
  color,
  isSelected,
  selectedStrokeWidth = 3,
  unselectedStrokeWidth = 2,
  unselectedFontWeight = 500,
  extraStyle,
  extraLabelStyle,
}: EdgeVisualOptions) {
  return {
    style: {
      stroke: color,
      strokeWidth: isSelected ? selectedStrokeWidth : unselectedStrokeWidth,
      ...extraStyle,
    },
    markerEnd: { type: MarkerType.ArrowClosed, color },
    labelStyle: {
      fill: color,
      fontWeight: isSelected ? 700 : unselectedFontWeight,
      ...extraLabelStyle,
    },
    labelBgStyle: { fill: edgeLabelBg() },
  };
}

export function createRelationEdges(relations: Relation[], ctx: EdgeCreationContext): Edge[] {
  return relations.map((relation) => {
    const isSelected = ctx.selectedEdgeId === relation.id;
    const isTriggers = relation.relationType === 'Triggers';
    const { sourceHandle, targetHandle } = resolveHandles(ctx, relation.sourceComponentId, relation.targetComponentId);

    const edgeColor = ctx.isClassicScheme
      ? CLASSIC_EDGE_COLOR
      : isTriggers
        ? resolveToken('--color-triggers', '#C25E0A')
        : resolveToken('--color-serves', '#4768A8');

    return {
      id: relation.id,
      source: relation.sourceComponentId,
      target: relation.targetComponentId,
      sourceHandle,
      targetHandle,
      label: relation.name || relation.relationType,
      type: ctx.edgeType,
      animated: isSelected,
      ...buildEdgeVisuals({ color: edgeColor, isSelected }),
    };
  });
}

export function createParentEdges(
  viewCapabilities: ViewCapability[],
  capabilities: Capability[],
  ctx: EdgeCreationContext,
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
      const { sourceHandle, targetHandle } = resolveHandles(ctx, parentNodeId, childNodeId);
      const parentEdgeColor = ctx.isClassicScheme ? CLASSIC_EDGE_COLOR : resolveToken('--color-gray-700', '#3D4A54');

      return {
        id: edgeId,
        source: parentNodeId,
        target: childNodeId,
        sourceHandle,
        targetHandle,
        label: 'Parent',
        type: ctx.edgeType,
        animated: isSelected,
        ...buildEdgeVisuals({
          color: parentEdgeColor,
          isSelected,
          selectedStrokeWidth: 3,
          unselectedStrokeWidth: 3,
          unselectedFontWeight: 600,
        }),
      };
    })
    .filter((e): e is Edge => e !== null);
}

interface RealizationVisibility {
  visibleCapabilityIds: Set<string>;
  componentIdsOnCanvas: Set<string>;
  allRealizations: CapabilityRealization[];
}

function isRealizationVisible(realization: CapabilityRealization, visibility: RealizationVisibility): boolean {
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

function buildRealizationEdge(realization: CapabilityRealization, ctx: EdgeCreationContext): Edge {
  const edgeId = `realization-${realization.id}`;
  const isSelected = ctx.selectedEdgeId === edgeId;
  const isInherited = realization.origin === 'Inherited';
  const realizationColor = ctx.isClassicScheme ? CLASSIC_EDGE_COLOR : resolveToken('--status-positive', '#0C7A55');
  const targetNodeId = `cap-${realization.capabilityId}`;
  const { sourceHandle, targetHandle } = resolveHandles(ctx, realization.componentId, targetNodeId);

  return {
    id: edgeId,
    source: realization.componentId,
    target: targetNodeId,
    sourceHandle,
    targetHandle,
    label: isInherited ? 'Realizes (inherited)' : 'Realizes',
    type: ctx.edgeType,
    animated: isSelected,
    className: isInherited ? 'realization-edge inherited' : 'realization-edge',
    ...buildEdgeVisuals({
      color: realizationColor,
      isSelected,
      extraStyle: { strokeDasharray: '5,5', opacity: isInherited ? 0.6 : 1.0 },
      extraLabelStyle: { opacity: isInherited ? 0.8 : 1.0 },
    }),
  };
}

export function createRealizationEdges(
  capabilityRealizations: CapabilityRealization[],
  viewCapabilities: ViewCapability[],
  viewComponents: ViewComponent[],
  ctx: EdgeCreationContext,
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

const ORIGIN_RELATIONSHIP_COLOR_TOKENS: Record<OriginRelationshipType, [string, string]> = {
  AcquiredVia: ['--entity-acquired', '#5F4FC7'],
  PurchasedFrom: ['--entity-vendor', '#B23A6B'],
  BuiltBy: ['--entity-team', '#0E7A80'],
};

const originRelationshipColor = (type: OriginRelationshipType): string => {
  const [token, fallback] = ORIGIN_RELATIONSHIP_COLOR_TOKENS[type];
  return resolveToken(token, fallback);
};

const getOriginEntityNodeId = (relationshipType: OriginRelationshipType, entityId: string): string =>
  makeNodeId(RELATIONSHIP_TO_ENTITY_TYPE[relationshipType], entityId);

export function createOriginRelationshipEdges(
  originRelationships: OriginRelationship[],
  originEntityNodeIds: Set<string>,
  componentIdsOnCanvas: Set<string>,
  ctx: EdgeCreationContext,
): Edge[] {
  return originRelationships
    .filter((rel) => {
      const originNodeId = getOriginEntityNodeId(rel.relationshipType, rel.originEntityId);
      return originEntityNodeIds.has(originNodeId) && componentIdsOnCanvas.has(rel.componentId);
    })
    .map((rel) => {
      const edgeId = `origin-${rel.relationshipType}-${rel.componentId}`;
      const isSelected = ctx.selectedEdgeId === edgeId;
      const edgeColor = ctx.isClassicScheme ? CLASSIC_EDGE_COLOR : originRelationshipColor(rel.relationshipType);
      const label = ORIGIN_RELATIONSHIP_LABELS[rel.relationshipType];

      const sourceNodeId = rel.componentId;
      const targetNodeId = getOriginEntityNodeId(rel.relationshipType, rel.originEntityId);
      const { sourceHandle, targetHandle } = resolveHandles(ctx, sourceNodeId, targetNodeId);

      return {
        id: edgeId,
        source: sourceNodeId,
        target: targetNodeId,
        sourceHandle,
        targetHandle,
        label,
        type: ctx.edgeType,
        animated: isSelected,
        ...buildEdgeVisuals({ color: edgeColor, isSelected }),
      };
    });
}
