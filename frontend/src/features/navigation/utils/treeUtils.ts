import type { Capability } from '../../../api/types';
import type { CapabilityTreeNode } from '../types';

export const getPersistedBoolean = (key: string, defaultValue: boolean): boolean => {
  const saved = localStorage.getItem(key);
  return saved !== null ? JSON.parse(saved) : defaultValue;
};

export const getPersistedSet = (key: string): Set<string> => {
  const saved = localStorage.getItem(key);
  return saved ? new Set(JSON.parse(saved)) : new Set();
};

const LEVEL_NUMBER_MAP: Record<string, number> = {
  L1: 1, L2: 2, L3: 3, L4: 4,
};

export const getLevelNumber = (level: string): number => LEVEL_NUMBER_MAP[level] ?? 1;

export const getContextMenuPosition = (e: React.MouseEvent) => {
  e.preventDefault();
  e.stopPropagation();
  return { x: e.clientX, y: e.clientY };
};

export const buildCapabilityTree = (capabilities: Capability[]): CapabilityTreeNode[] => {
  const capabilityMap = new Map<string, CapabilityTreeNode>();

  capabilities.forEach((cap) => {
    capabilityMap.set(cap.id, { capability: cap, children: [] });
  });

  const roots: CapabilityTreeNode[] = [];

  capabilities.forEach((cap) => {
    const node = capabilityMap.get(cap.id)!;
    if (cap.parentId && capabilityMap.has(cap.parentId)) {
      capabilityMap.get(cap.parentId)!.children.push(node);
    } else {
      roots.push(node);
    }
  });

  roots.sort((a, b) => a.capability.name.localeCompare(b.capability.name));

  return roots;
};

export const hasCustomColor = (
  colorScheme: string | undefined,
  customColor: string | undefined | null
): boolean =>
  colorScheme === 'custom' &&
  customColor !== undefined &&
  customColor !== null &&
  customColor !== '';
