import { describe, expect, it, vi } from 'vitest';
import type { ViewId } from '../../../api/types';
import { saveDraft, type DraftSaveApi, type DraftSaveInput } from './saveDraft';

function makeApi(): DraftSaveApi & { calls: string[] } {
  const calls: string[] = [];
  const stub = (name: string) =>
    vi.fn(async (...args: unknown[]) => {
      calls.push(`${name}:${JSON.stringify(args)}`);
    });
  return {
    calls,
    addComponent: stub('addComponent'),
    addCapability: stub('addCapability'),
    addOriginEntity: stub('addOriginEntity'),
    removeComponent: stub('removeComponent'),
    removeCapability: stub('removeCapability'),
    removeOriginEntity: stub('removeOriginEntity'),
    updateComponentPosition: stub('updateComponentPosition'),
    updateCapabilityPosition: stub('updateCapabilityPosition'),
    updateOriginEntityPosition: stub('updateOriginEntityPosition'),
  };
}

const viewId = 'view-1' as ViewId;
const emptyInput: DraftSaveInput = {
  viewId,
  additions: [],
  removals: [],
  positionDeltas: [],
};

describe('saveDraft', () => {
  it('makes no calls when draft is empty', async () => {
    const api = makeApi();
    const result = await saveDraft(api, emptyInput);

    expect(api.calls).toEqual([]);
    expect(result.successCount).toBe(0);
    expect(result.failures).toEqual([]);
  });

  describe.each([
    {
      operation: 'add' as const,
      input: {
        additions: [
          { id: 'comp-1', type: 'component' as const, x: 1, y: 2 },
          { id: 'cap-1', type: 'capability' as const, x: 3, y: 4 },
          { id: 'org-1', type: 'originEntity' as const, x: 5, y: 6 },
        ],
      },
      expectations: (api: ReturnType<typeof makeApi>) => {
        expect(api.addComponent).toHaveBeenCalledWith(viewId, 'comp-1', 1, 2);
        expect(api.addCapability).toHaveBeenCalledWith(viewId, 'cap-1', 3, 4);
        expect(api.addOriginEntity).toHaveBeenCalledWith(viewId, 'org-1', 5, 6);
      },
    },
    {
      operation: 'remove' as const,
      input: {
        removals: [
          { id: 'comp-1', type: 'component' as const },
          { id: 'cap-1', type: 'capability' as const },
          { id: 'org-1', type: 'originEntity' as const },
        ],
      },
      expectations: (api: ReturnType<typeof makeApi>) => {
        expect(api.removeComponent).toHaveBeenCalledWith(viewId, 'comp-1');
        expect(api.removeCapability).toHaveBeenCalledWith(viewId, 'cap-1');
        expect(api.removeOriginEntity).toHaveBeenCalledWith(viewId, 'org-1');
      },
    },
    {
      operation: 'position' as const,
      input: {
        positionDeltas: [
          { id: 'comp-1', type: 'component' as const, x: 10, y: 20 },
          { id: 'cap-1', type: 'capability' as const, x: 30, y: 40 },
          { id: 'org-1', type: 'originEntity' as const, x: 50, y: 60 },
        ],
      },
      expectations: (api: ReturnType<typeof makeApi>) => {
        expect(api.updateComponentPosition).toHaveBeenCalledWith(viewId, 'comp-1', 10, 20);
        expect(api.updateCapabilityPosition).toHaveBeenCalledWith(viewId, 'cap-1', 30, 40);
        expect(api.updateOriginEntityPosition).toHaveBeenCalledWith(viewId, 'org-1', 50, 60);
      },
    },
  ])('dispatches $operation calls per entity type', ({ input, expectations }) => {
    it('routes to the right API method for each type', async () => {
      const api = makeApi();
      const result = await saveDraft(api, { ...emptyInput, ...input });
      expectations(api);
      expect(result.successCount).toBe(3);
      expect(result.failures).toEqual([]);
    });
  });

  it('does not separately update position for newly-added entities (add already includes position)', async () => {
    const api = makeApi();
    await saveDraft(api, {
      ...emptyInput,
      additions: [{ id: 'comp-1', type: 'component', x: 1, y: 2 }],
      positionDeltas: [{ id: 'comp-1', type: 'component', x: 1, y: 2 }],
    });

    expect(api.addComponent).toHaveBeenCalledOnce();
    expect(api.updateComponentPosition).not.toHaveBeenCalled();
  });

  it('runs removals before additions (deterministic order)', async () => {
    const api = makeApi();
    await saveDraft(api, {
      viewId,
      removals: [{ id: 'comp-rm', type: 'component' }],
      additions: [{ id: 'comp-add', type: 'component', x: 0, y: 0 }],
      positionDeltas: [{ id: 'comp-pos', type: 'component', x: 0, y: 0 }],
    });

    const removeIdx = api.calls.findIndex((c) => c.startsWith('removeComponent'));
    const addIdx = api.calls.findIndex((c) => c.startsWith('addComponent'));
    const posIdx = api.calls.findIndex((c) => c.startsWith('updateComponentPosition'));
    expect(removeIdx).toBeLessThan(addIdx);
    expect(addIdx).toBeLessThan(posIdx);
  });

  it('captures failures and continues with the remaining operations', async () => {
    const api = makeApi();
    api.addCapability = vi.fn(async () => {
      throw new Error('boom');
    });

    const result = await saveDraft(api, {
      ...emptyInput,
      additions: [
        { id: 'comp-1', type: 'component', x: 0, y: 0 },
        { id: 'cap-1', type: 'capability', x: 0, y: 0 },
        { id: 'org-1', type: 'originEntity', x: 0, y: 0 },
      ],
    });

    expect(api.addComponent).toHaveBeenCalled();
    expect(api.addCapability).toHaveBeenCalled();
    expect(api.addOriginEntity).toHaveBeenCalled();
    expect(result.successCount).toBe(2);
    expect(result.failures).toEqual([
      { entity: { id: 'cap-1', type: 'capability' }, operation: 'add', message: 'boom' },
    ]);
  });
});
