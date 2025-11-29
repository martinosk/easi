import { useState, useEffect, useCallback, useMemo } from 'react';
import { apiClient } from '../../../api/client';
import type { Capability, CapabilityId } from '../../../api/types';

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
  const [capabilities, setCapabilities] = useState<Capability[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchCapabilities = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await apiClient.getCapabilities();
      setCapabilities(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch capabilities'));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCapabilities();
  }, [fetchCapabilities]);

  const tree = useMemo(() => buildTree(capabilities), [capabilities]);
  const orphanedL1Ids = useMemo(() => findOrphanedL1s(tree), [tree]);

  return {
    tree,
    isLoading,
    error,
    refetch: fetchCapabilities,
    orphanedL1Ids,
  };
}
