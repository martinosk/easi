import type { CapabilityId, CapabilityRealization } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';

export type RealizationsLookup = (capabilityId: CapabilityId) => CapabilityRealization[];

function realizationMatches(realizations: CapabilityRealization[], query: string): boolean {
  return realizations.some((realization) =>
    (realization.componentName ?? realization.componentId).toLowerCase().includes(query),
  );
}

export function nodeMatchesSearch(
  node: CapabilityTreeNode,
  query: string,
  getRealizations: RealizationsLookup,
): boolean {
  if (!query) return true;
  const normalized = query.toLowerCase();

  if (node.capability.name.toLowerCase().includes(normalized)) return true;
  if (realizationMatches(getRealizations(node.capability.id), normalized)) return true;

  return node.children.some((child) => nodeMatchesSearch(child, normalized, getRealizations));
}
