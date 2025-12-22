import { useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';

export interface CapabilityTreeNode {
  capability: Capability;
  children: CapabilityTreeNode[];
}

export interface UseCapabilityTreeResult {
  tree: CapabilityTreeNode[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  orphanedL1Ids: Set<CapabilityId>;
}

function buildTree(capabilities: Capability[]): CapabilityTreeNode[] {
  const map = new Map<CapabilityId, CapabilityTreeNode>();

  capabilities.forEach((cap) => {
    map.set(cap.id, { capability: cap, children: [] });
  });

  const roots: CapabilityTreeNode[] = [];

  capabilities.forEach((cap) => {
    const node = map.get(cap.id)!;
    if (cap.parentId && map.has(cap.parentId)) {
      map.get(cap.parentId)!.children.push(node);
    } else if (cap.level === 'L1') {
      roots.push(node);
    }
  });

  return roots.sort((a, b) => a.capability.name.localeCompare(b.capability.name));
}

function findOrphanedL1s(tree: CapabilityTreeNode[]): Set<CapabilityId> {
  const orphaned = new Set<CapabilityId>();

  tree.forEach((node) => {
    if (node.capability.level === 'L1' && node.children.length === 0) {
      orphaned.add(node.capability.id);
    }
  });

  return orphaned;
}

export function useCapabilityTree(): UseCapabilityTreeResult {
  const { data: capabilities = [], isLoading, error, refetch } = useCapabilities();

  const tree = useMemo(() => buildTree(capabilities), [capabilities]);
  const orphanedL1Ids = useMemo(() => findOrphanedL1s(tree), [tree]);

  return {
    tree,
    isLoading,
    error: error ?? null,
    refetch: async () => { await refetch(); },
    orphanedL1Ids,
  };
}
