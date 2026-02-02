import dagre from 'dagre';
import type { Node, Edge } from '@xyflow/react';

export type EntityType = 'capability' | 'component' | 'originEntity';

export interface EntityLayoutMetadata {
  nodeId: string;
  entityType: EntityType;
  layer: number;
  sublayer?: number;
  weight?: number;
}

export interface LayoutStrategy {
  extractMetadata(node: Node): EntityLayoutMetadata;
  getLayer(): number;
}

class CapabilityLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 0;
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    const rawLevel = node.data?.level;
    let level: number;
    if (typeof rawLevel === 'number') {
      level = rawLevel;
    } else if (typeof rawLevel === 'string') {
      const match = rawLevel.match(/^L(\d)$/i);
      level = match ? Number(match[1]) : Number(rawLevel) || 1;
    } else {
      level = 1;
    }
    return {
      nodeId: node.id,
      entityType: 'capability',
      layer: 0,
      sublayer: Math.max(0, level - 1),
    };
  }
}

class ComponentLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 1;
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    return {
      nodeId: node.id,
      entityType: 'component',
      layer: 1,
    };
  }
}

class OriginEntityLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 2;
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    return {
      nodeId: node.id,
      entityType: 'originEntity',
      layer: 2,
    };
  }
}

const LAYOUT_STRATEGIES = new Map<EntityType, LayoutStrategy>([
  ['capability', new CapabilityLayoutStrategy()],
  ['component', new ComponentLayoutStrategy()],
  ['originEntity', new OriginEntityLayoutStrategy()],
]);

function getStrategyForNode(node: Node): LayoutStrategy {
  if (node.type === 'capability') return LAYOUT_STRATEGIES.get('capability')!;
  if (node.type === 'component') return LAYOUT_STRATEGIES.get('component')!;
  if (node.type === 'originEntity') return LAYOUT_STRATEGIES.get('originEntity')!;
  throw new Error(`No layout strategy for node type: ${node.type}`);
}

export interface AutoLayoutOptions {
  nodeSpacing?: number;
  layerSpacing?: number;
  sublayerSpacing?: number;
}

export function calculateAutoLayout(
  nodes: Node[],
  edges: Edge[],
  options: AutoLayoutOptions = {}
): Node[] {
  if (nodes.length === 0) return nodes;

  const {
    nodeSpacing = 120,
    layerSpacing = 200,
    sublayerSpacing = 100,
  } = options;

  const metadataMap = new Map<string, EntityLayoutMetadata>();
  let maxSublayer = 0;
  let hasCapabilities = false;

  for (const node of nodes) {
    const strategy = getStrategyForNode(node);
    const metadata = strategy.extractMetadata(node);
    metadataMap.set(node.id, metadata);
    if (metadata.sublayer !== undefined) {
      maxSublayer = Math.max(maxSublayer, metadata.sublayer);
    }
    if (metadata.entityType === 'capability') {
      hasCapabilities = true;
    }
  }

  const layerRankStride = Math.max(maxSublayer + 2, Math.ceil(layerSpacing / sublayerSpacing));

  const graph = new dagre.graphlib.Graph();
  graph.setDefaultEdgeLabel(() => ({}));
  graph.setGraph({
    rankdir: 'TB',
    nodesep: nodeSpacing,
    ranksep: sublayerSpacing,
    ranker: 'tight-tree',
  });

  for (const node of nodes) {
    const metadata = metadataMap.get(node.id)!;
    const rank = metadata.layer * layerRankStride + (metadata.sublayer ?? 0);
    graph.setNode(node.id, { width: 180, height: 80, rank });
  }

  for (const edge of edges) {
    if (graph.hasNode(edge.source) && graph.hasNode(edge.target)) {
      graph.setEdge(edge.source, edge.target);
    }
  }

  dagre.layout(graph);

  const nodeHeight = 80;
  const levelPitch = nodeHeight + sublayerSpacing;
  const capabilityZoneHeight = hasCapabilities ? (maxSublayer + 1) * levelPitch : 0;
  const componentCenterY = capabilityZoneHeight > 0
    ? capabilityZoneHeight + layerSpacing + nodeHeight / 2
    : nodeHeight / 2;
  const originCenterY = componentCenterY + layerSpacing + nodeHeight;

  return nodes.map((node) => {
    const nodeWithPosition = graph.node(node.id);
    if (!nodeWithPosition) return node;
    const metadata = metadataMap.get(node.id);
    let centerY = nodeWithPosition.y;
    if (metadata?.entityType === 'capability') {
      const sublayer = metadata.sublayer ?? 0;
      centerY = nodeHeight / 2 + sublayer * levelPitch;
    } else if (metadata?.entityType === 'component') {
      centerY = componentCenterY;
    } else if (metadata?.entityType === 'originEntity') {
      centerY = originCenterY;
    }
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - 90,
        y: centerY - 40,
      },
    };
  });
}