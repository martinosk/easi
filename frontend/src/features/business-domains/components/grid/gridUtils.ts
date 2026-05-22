import type { Capability, CapabilityId, Position } from '../../../../api/types';

export interface CapabilityNode {
  capability: Capability;
  children: CapabilityNode[];
}

export interface PositionMap {
  [capabilityId: string]: Position;
}

export function buildTree(capabilities: Capability[]): CapabilityNode[] {
  const byId = new Map<CapabilityId, Capability>();
  const childrenMap = new Map<CapabilityId | undefined, Capability[]>();

  for (const cap of capabilities) {
    byId.set(cap.id, cap);
    const parentId = cap.parentId;
    if (!childrenMap.has(parentId)) {
      childrenMap.set(parentId, []);
    }
    childrenMap.get(parentId)!.push(cap);
  }

  function buildNode(cap: Capability): CapabilityNode {
    const children = (childrenMap.get(cap.id) || []).sort((a, b) => a.name.localeCompare(b.name)).map(buildNode);
    return { capability: cap, children };
  }

  const l1Caps = capabilities.filter((c) => c.level === 'L1');
  return l1Caps.sort((a, b) => a.name.localeCompare(b.name)).map(buildNode);
}

export function levelToNumber(level: Capability['level']): number {
  return parseInt(level.substring(1), 10);
}

export function compareNodesByPosition(a: CapabilityNode, b: CapabilityNode, positions: PositionMap): number {
  const posA = positions[a.capability.id];
  const posB = positions[b.capability.id];

  if (posA && posB) {
    if (posA.y !== posB.y) return posA.y - posB.y;
    return posA.x - posB.x;
  }
  if (posA) return -1;
  if (posB) return 1;
  return a.capability.name.localeCompare(b.capability.name);
}
