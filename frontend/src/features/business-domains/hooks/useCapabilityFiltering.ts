import { useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import type { CapabilityTreeNode } from './useCapabilityTree';

export function useCapabilityFiltering(
  tree: CapabilityTreeNode[],
  capabilities: Capability[]
) {
  const allCapabilities = useMemo(() => {
    const flatten = (nodes: typeof tree): Capability[] => {
      return nodes.flatMap((node) => [node.capability, ...flatten(node.children)]);
    };
    return flatten(tree);
  }, [tree]);

  const assignedCapabilityIds = useMemo(
    () => new Set<CapabilityId>(capabilities.map((c) => c.id)),
    [capabilities]
  );

  const capabilitiesWithDescendants = useMemo(() => {
    if (capabilities.length === 0 || tree.length === 0) return capabilities;

    const assignedL1Ids = new Set(capabilities.filter((c) => c.level === 'L1').map((c) => c.id));
    const result: Capability[] = [];

    const collectDescendants = (nodes: typeof tree) => {
      for (const node of nodes) {
        if (assignedL1Ids.has(node.capability.id)) {
          const addAll = (n: typeof tree[0]) => {
            result.push(n.capability);
            n.children.forEach(addAll);
          };
          addAll(node);
        } else {
          collectDescendants(node.children);
        }
      }
    };

    collectDescendants(tree);
    return result;
  }, [capabilities, tree]);

  return {
    allCapabilities,
    assignedCapabilityIds,
    capabilitiesWithDescendants,
  };
}
