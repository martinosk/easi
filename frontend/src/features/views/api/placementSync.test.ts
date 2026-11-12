import { describe, expect, it, vi } from 'vitest';
import { ApiError, type ViewId } from '../../../api/types';
import { afterPendingAdds, holdPlacement, ignoreMissingPlacement, trackAdd } from './placementSync';

function deferred<T = void>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

async function flush() {
  await new Promise((r) => setTimeout(r, 0));
}

function track(viewId: string, elementId: string) {
  const add = deferred();
  trackAdd(viewId as ViewId, elementId, add.promise).catch(() => undefined);
  return add;
}

function watch(promise: Promise<void>) {
  const state = { settled: false };
  promise.then(() => {
    state.settled = true;
  });
  return state;
}

describe('placementSync', () => {
  describe('afterPendingAdds', () => {
    it('waits for a pending add of the same view and element', async () => {
      const add = track('view-same', 'el-1');
      const waiter = watch(afterPendingAdds('view-same' as ViewId, ['el-1']));

      await flush();
      expect(waiter.settled).toBe(false);

      add.resolve();
      await flush();
      expect(waiter.settled).toBe(true);
    });

    it('does not wait for a pending add of another element', async () => {
      track('view-other-el', 'el-1');

      await expect(afterPendingAdds('view-other-el' as ViewId, ['el-2'])).resolves.toBeUndefined();
    });

    it('does not wait for a pending add of the same element in another view', async () => {
      track('view-a', 'el-1');

      await expect(afterPendingAdds('view-b' as ViewId, ['el-1'])).resolves.toBeUndefined();
    });

    it('is released when the add fails', async () => {
      const add = track('view-failing', 'el-1');
      const waiter = watch(afterPendingAdds('view-failing' as ViewId, ['el-1']));

      add.reject(new ApiError('boom', 500));
      await flush();

      expect(waiter.settled).toBe(true);
    });

    it('waits for every listed element', async () => {
      const first = track('view-many', 'el-1');
      const second = track('view-many', 'el-2');
      const waiter = watch(afterPendingAdds('view-many' as ViewId, ['el-1', 'el-2']));

      first.resolve();
      await flush();
      expect(waiter.settled).toBe(false);

      second.resolve();
      await flush();
      expect(waiter.settled).toBe(true);
    });

    it('does not wait once the add has settled', async () => {
      const add = track('view-settled', 'el-1');
      add.resolve();
      await flush();

      await expect(afterPendingAdds('view-settled' as ViewId, ['el-1'])).resolves.toBeUndefined();
    });

    it('keeps waiting for a newer add when an older add for the same key settles', async () => {
      const older = track('view-newer', 'el-1');
      const newer = track('view-newer', 'el-1');

      older.resolve();
      await flush();
      const waiter = watch(afterPendingAdds('view-newer' as ViewId, ['el-1']));
      await flush();
      expect(waiter.settled).toBe(false);

      newer.resolve();
      await flush();
      expect(waiter.settled).toBe(true);
    });
  });

  describe('holdPlacement', () => {
    it('makes position saves wait until released', async () => {
      const viewId = 'view-hold' as ViewId;
      const release = holdPlacement(viewId, 'held');
      let waited = false;
      const waiting = afterPendingAdds(viewId, ['held']).then(() => {
        waited = true;
      });

      await Promise.resolve();
      expect(waited).toBe(false);
      release();
      await waiting;
      expect(waited).toBe(true);
    });
  });

  describe('trackAdd', () => {
    it('returns the add promise unchanged', () => {
      const add = Promise.resolve('done');

      expect(trackAdd('view-return' as ViewId, 'el-1', add)).toBe(add);
    });
  });

  describe('ignoreMissingPlacement', () => {
    it('swallows a 404', async () => {
      await expect(
        ignoreMissingPlacement(() => Promise.reject(new ApiError('not in view', 404))),
      ).resolves.toBeUndefined();
    });

    it.each([400, 409, 500])('rethrows a %i', async (status) => {
      const error = new ApiError('failed', status);

      await expect(ignoreMissingPlacement(() => Promise.reject(error))).rejects.toBe(error);
    });

    it('rethrows non-API errors', async () => {
      const error = new Error('network');

      await expect(ignoreMissingPlacement(() => Promise.reject(error))).rejects.toBe(error);
    });

    it('performs the write', async () => {
      const write = vi.fn().mockResolvedValue(undefined);

      await ignoreMissingPlacement(write);

      expect(write).toHaveBeenCalledOnce();
    });
  });
});
