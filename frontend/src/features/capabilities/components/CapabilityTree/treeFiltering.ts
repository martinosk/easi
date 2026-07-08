import type { CapabilityTreeNode } from '../../hooks/useCapabilityTree';

export function filterTreeByName(tree: CapabilityTreeNode[], query: string): CapabilityTreeNode[] {
  const trimmed = query.trim().toLowerCase();
  if (!trimmed) return tree;

  const filterNodes = (nodes: CapabilityTreeNode[]): CapabilityTreeNode[] => {
    const result: CapabilityTreeNode[] = [];
    for (const node of nodes) {
      const children = filterNodes(node.children);
      if (node.capability.name.toLowerCase().includes(trimmed) || children.length > 0) {
        result.push({ ...node, children });
      }
    }
    return result;
  };

  return filterNodes(tree);
}

export function collectExpandableIds(tree: CapabilityTreeNode[]): Set<string> {
  const ids = new Set<string>();
  const walk = (nodes: CapabilityTreeNode[]) => {
    for (const node of nodes) {
      if (node.children.length > 0) {
        ids.add(node.capability.id);
        walk(node.children);
      }
    }
  };
  walk(tree);
  return ids;
}
