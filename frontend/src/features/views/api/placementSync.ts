import { ApiError, type ViewId } from '../../../api/types';

const pendingAdds = new Map<string, Set<Promise<unknown>>>();

function placementKey(viewId: ViewId, elementId: string): string {
  return `${viewId}:${elementId}`;
}

function register(key: string, pending: Promise<unknown>): void {
  const set = pendingAdds.get(key) ?? new Set<Promise<unknown>>();
  set.add(pending);
  pendingAdds.set(key, set);
  const forget = () => {
    set.delete(pending);
    if (set.size === 0 && pendingAdds.get(key) === set) pendingAdds.delete(key);
  };
  pending.then(forget, forget);
}

export function trackAdd<T>(viewId: ViewId, elementId: string, add: Promise<T>): Promise<T> {
  register(placementKey(viewId, elementId), add);
  return add;
}

export function holdPlacement(viewId: ViewId, elementId: string): () => void {
  let release: () => void = () => undefined;
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  register(placementKey(viewId, elementId), held);
  return release;
}

export async function afterPendingAdds(viewId: ViewId, elementIds: string[]): Promise<void> {
  const pending = elementIds.flatMap((id) => [...(pendingAdds.get(placementKey(viewId, id)) ?? [])]);
  await Promise.allSettled(pending);
}

export async function ignoreMissingPlacement(write: () => Promise<unknown>): Promise<void> {
  try {
    await write();
  } catch (error) {
    if (error instanceof ApiError && error.statusCode === 404) return;
    throw error;
  }
}
