import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { View } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { selectDirtyForView } from '../../../store/slices/dynamicModeSlice';
import { buildSnapshotFromView, useDynamicSnapshot } from './useDynamicSnapshot';

let currentViewState: { currentView: View | null; currentViewId: string | null } = {
  currentView: null,
  currentViewId: null,
};

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => currentViewState,
}));

vi.mock('../../../utils/hateoas', () => ({
  canEdit: () => true,
}));

function makeView(id: string, componentIds: string[]): View {
  return {
    id,
    name: id,
    components: componentIds.map((cid) => ({ componentId: cid, x: 0, y: 0 })),
    capabilities: [],
    originEntities: [],
    edgeType: 'default',
    colorScheme: 'maturity',
    _links: { self: { href: `/v/${id}`, method: 'GET' }, edit: { href: `/v/${id}`, method: 'PUT' } },
  } as unknown as View;
}

function setActiveView(view: View | null) {
  currentViewState = { currentView: view, currentViewId: view?.id ?? null };
}

describe('buildSnapshotFromView', () => {
  function viewAt(componentId: string, x: number, y: number): View {
    const view = makeView('view-a', [componentId]);
    view.components = [{ componentId, x, y }] as View['components'];
    return view;
  }

  it('seeds positions from view coordinates', () => {
    const snapshot = buildSnapshotFromView(viewAt('comp-1', 12, 34));

    expect(snapshot.positions['comp-1']).toEqual({ x: 12, y: 34 });
  });

  it('seeds capability and origin entity positions from view coordinates', () => {
    const view = makeView('view-a', []);
    view.capabilities = [{ capabilityId: 'cap-1', x: 55, y: 66 }] as View['capabilities'];
    view.originEntities = [{ originEntityId: 'oe-1', x: 77, y: 88 }] as View['originEntities'];

    const snapshot = buildSnapshotFromView(view);

    expect(snapshot.positions['cap-1']).toEqual({ x: 55, y: 66 });
    expect(snapshot.positions['oe-1']).toEqual({ x: 77, y: 88 });
    expect(snapshot.entities).toEqual([
      { id: 'cap-1', type: 'capability' },
      { id: 'oe-1', type: 'originEntity' },
    ]);
  });

  it('seeds no position for an element the view has no coordinates for', () => {
    const view = makeView('view-a', ['comp-1']);
    view.components = [{ componentId: 'comp-1' }] as unknown as View['components'];

    const snapshot = buildSnapshotFromView(view);

    expect(snapshot.entities).toEqual([{ id: 'comp-1', type: 'component' }]);
    expect(snapshot.positions['comp-1']).toBeUndefined();
  });
});

describe('useDynamicSnapshot — per-view drafts', () => {
  beforeEach(() => {
    setActiveView(null);
    useAppStore.setState({
      dynamicOriginal: null,
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicPositions: {},
      draftsByView: {},
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('builds a fresh snapshot for a brand-new view', () => {
    const viewA = makeView('view-a', ['comp-1']);
    setActiveView(viewA);

    renderHook(() => useDynamicSnapshot());

    expect(useAppStore.getState().dynamicViewId).toBe('view-a');
    expect(useAppStore.getState().dynamicEntities).toEqual([{ id: 'comp-1', type: 'component' }]);
    expect(useAppStore.getState().draftsByView['view-a']).toBeDefined();
  });

  it('preserves dirty state for view A when switching to view B', () => {
    const viewA = makeView('view-a', ['comp-1']);
    const viewB = makeView('view-b', ['comp-2']);
    setActiveView(viewA);
    const { rerender } = renderHook(() => useDynamicSnapshot());

    act(() => {
      useAppStore.getState().draftAddEntities([{ id: 'comp-x', type: 'component' }], { 'comp-x': { x: 5, y: 5 } });
    });

    setActiveView(viewB);
    rerender();

    expect(selectDirtyForView(useAppStore.getState(), 'view-a')).toBe(true);
    expect(selectDirtyForView(useAppStore.getState(), 'view-b')).toBe(false);
    expect(useAppStore.getState().dynamicViewId).toBe('view-b');
  });

  it('hydrates view A scalars when switching back to it', () => {
    const viewA = makeView('view-a', ['comp-1']);
    const viewB = makeView('view-b', ['comp-2']);
    setActiveView(viewA);
    const { rerender } = renderHook(() => useDynamicSnapshot());

    act(() => {
      useAppStore.getState().draftAddEntities([{ id: 'comp-x', type: 'component' }], { 'comp-x': { x: 5, y: 5 } });
    });
    setActiveView(viewB);
    rerender();

    setActiveView(viewA);
    rerender();

    const state = useAppStore.getState();
    expect(state.dynamicViewId).toBe('view-a');
    expect(state.dynamicEntities).toContainEqual({ id: 'comp-x', type: 'component' });
    expect(state.dynamicPositions['comp-x']).toEqual({ x: 5, y: 5 });
  });
});
