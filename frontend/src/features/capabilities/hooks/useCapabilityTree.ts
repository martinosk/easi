import { useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import { useCapabilities } from './useCapabilities';

export interface CapabilityTreeNode {
  capability: Capability;
  children: CapabilityTreeNode[];
}

export interface BuildCapabilityTreeOptions {
  orphanRoots?: 'l1-only' | 'any-level';
}

export interface UseCapabilityTreeResult {
  tree: CapabilityTreeNode[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  orphanedL1Ids: Set<CapabilityId>;
}

export function buildCapabilityTree(
  capabilities: Capability[],
  options: BuildCapabilityTreeOptions = {},
): CapabilityTreeNode[] {
  const orphanRoots = options.orphanRoots ?? 'l1-only';
  const map = new Map<CapabilityId, CapabilityTreeNode>();

  capabilities.forEach((cap) => {
    map.set(cap.id, { capability: cap, children: [] });
  });

  const roots: CapabilityTreeNode[] = [];

  capabilities.forEach((cap) => {
    const node = map.get(cap.id)!;
    if (cap.parentId && map.has(cap.parentId)) {
      map.get(cap.parentId)!.children.push(node);
    } else if (orphanRoots === 'any-level' || cap.level === 'L1') {
      roots.push(node);
    }
  });

  return roots.sort((a, b) => a.capability.name.localeCompare(b.capability.name));
}

export function findOrphanedL1Ids(tree: CapabilityTreeNode[]): Set<CapabilityId> {
  const orphaned = new Set<CapabilityId>();

  tree.forEach((node) => {
    if (node.capability.level === 'L1' && node.children.length === 0) {
      orphaned.add(node.capability.id);
    }
  });

  return orphaned;
}

export function useCapabilityTree(options: BuildCapabilityTreeOptions = {}): UseCapabilityTreeResult {
  const { data: capabilities = [], isLoading, error, refetch } = useCapabilities();
  const orphanRoots = options.orphanRoots;

  const tree = useMemo(() => buildCapabilityTree(capabilities, { orphanRoots }), [capabilities, orphanRoots]);
  const orphanedL1Ids = useMemo(() => findOrphanedL1Ids(tree), [tree]);

  return {
    tree,
    isLoading,
    error: error ?? null,
    refetch: async () => {
      await refetch();
    },
    orphanedL1Ids,
  };
}
