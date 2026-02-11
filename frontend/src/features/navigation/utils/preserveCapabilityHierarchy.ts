import type { Capability } from '../../../api/types';

export function preserveCapabilityHierarchy(
  filteredCapabilities: Capability[],
  allCapabilities: Capability[]
): Capability[] {
  if (filteredCapabilities.length === 0) {
    return [];
  }

  const allById = new Map(allCapabilities.map((c) => [c.id, c]));
  const resultIds = new Set(filteredCapabilities.map((c) => c.id));

  for (const cap of filteredCapabilities) {
    let current = cap;
    while (current.parentId) {
      if (resultIds.has(current.parentId)) break;
      const parent = allById.get(current.parentId);
      if (!parent) break;
      resultIds.add(parent.id);
      current = parent;
    }
  }

  return allCapabilities.filter((c) => resultIds.has(c.id));
}
