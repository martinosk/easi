import { useCallback, useState } from 'react';

export interface DetailGroupLayout {
  order: readonly string[];
  collapsed: readonly string[];
  toggle: (id: string) => void;
  swap: (a: string, b: string) => void;
}

interface StoredLayout {
  order: string[];
  collapsed: string[];
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === 'string');
}

function readStored(key: string): Partial<StoredLayout> {
  try {
    const parsed: unknown = JSON.parse(localStorage.getItem(key) ?? '');
    if (typeof parsed !== 'object' || parsed === null) return {};
    const { order, collapsed } = parsed as Record<string, unknown>;
    return {
      order: isStringArray(order) ? order : undefined,
      collapsed: isStringArray(collapsed) ? collapsed : undefined,
    };
  } catch {
    return {};
  }
}

export function reconcileLayout(stored: Partial<StoredLayout>, defaultOrder: readonly string[]): StoredLayout {
  const known = new Set(defaultOrder);
  const kept = (stored.order ?? []).filter((id) => known.has(id));
  const missing = defaultOrder.filter((id) => !kept.includes(id));
  return {
    order: [...kept, ...missing],
    collapsed: (stored.collapsed ?? []).filter((id) => known.has(id)),
  };
}

function swapped(order: readonly string[], a: string, b: string): string[] {
  const next = [...order];
  const indexA = next.indexOf(a);
  const indexB = next.indexOf(b);
  if (indexA === -1 || indexB === -1) return next;
  next[indexA] = b;
  next[indexB] = a;
  return next;
}

export function useDetailGroupLayout(storageKey: string, defaultOrder: readonly string[]): DetailGroupLayout {
  const [layout, setLayout] = useState<StoredLayout>(() => reconcileLayout(readStored(storageKey), defaultOrder));

  const update = useCallback(
    (change: (current: StoredLayout) => StoredLayout) => {
      setLayout((current) => {
        const next = change(current);
        localStorage.setItem(storageKey, JSON.stringify(next));
        return next;
      });
    },
    [storageKey],
  );

  const toggle = useCallback(
    (id: string) =>
      update((current) => ({
        ...current,
        collapsed: current.collapsed.includes(id)
          ? current.collapsed.filter((item) => item !== id)
          : [...current.collapsed, id],
      })),
    [update],
  );

  const swap = useCallback(
    (a: string, b: string) => update((current) => ({ ...current, order: swapped(current.order, a, b) })),
    [update],
  );

  return { order: layout.order, collapsed: layout.collapsed, toggle, swap };
}
