import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import type { MapDepth } from './useMapViewState';

function toMapNode(node: CapabilityTreeNode, remainingLevels: number): CapabilityTreeNode {
  const children =
    remainingLevels > 1
      ? node.children
          .map((child) => toMapNode(child, remainingLevels - 1))
          .sort((a, b) => a.capability.name.localeCompare(b.capability.name))
      : [];
  return { capability: node.capability, children };
}

export function buildMapTree(l1Nodes: CapabilityTreeNode[], depth: MapDepth): CapabilityTreeNode[] {
  return l1Nodes.map((node) => toMapNode(node, depth)).sort((a, b) => a.capability.name.localeCompare(b.capability.name));
}
