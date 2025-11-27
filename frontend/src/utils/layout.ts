import dagre from 'dagre';
import type { Node, Edge } from '@xyflow/react';

export interface LayoutOptions {
  direction?: 'TB' | 'LR' | 'BT' | 'RL';
  nodeSpacing?: number;
  rankSpacing?: number;
}

export const calculateDagreLayout = (
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions = {}
): Node[] => {
  const {
    direction = 'TB',
    nodeSpacing = 100,
    rankSpacing = 150,
  } = options;

  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const isHorizontal = direction === 'LR' || direction === 'RL';

  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: nodeSpacing,
    ranksep: rankSpacing,
  });

  nodes.forEach((node) => {
    const width = 180;
    const height = 80;
    const layer = node.data?.layer ?? 0;
    dagreGraph.setNode(node.id, { width, height, rank: layer });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);

    return {
      ...node,
      position: {
        x: nodeWithPosition.x - (isHorizontal ? 90 : 90),
        y: nodeWithPosition.y - (isHorizontal ? 40 : 40),
      },
    };
  });

  return layoutedNodes;
};
