import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useAppStore } from '../store/appStore';
import { useUnloadGuard } from './useUnloadGuard';

function dispatchBeforeUnload(): { defaultPrevented: boolean } {
  const event = new Event('beforeunload', { cancelable: true });
  window.dispatchEvent(event);
  return { defaultPrevented: event.defaultPrevented };
}

function setDirtyDraft(viewId: string) {
  useAppStore.setState({
    draftsByView: {
      [viewId]: {
        original: { entities: [], positions: {} },
        entities: [{ id: 'X', type: 'component' }],
        positions: { X: { x: 0, y: 0 } },
        filters: {
          edges: { relation: true, realization: true, parentage: true, origin: true },
          types: { component: true, capability: true, originEntity: true },
        },
      },
    },
  });
}

describe('useUnloadGuard', () => {
  beforeEach(() => {
    useAppStore.setState({
      dynamicEnabled: false,
      dynamicOriginal: null,
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicPositions: {},
      draftsByView: {},
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('does not preventDefault when no draft is dirty', () => {
    renderHook(() => useUnloadGuard());

    const { defaultPrevented } = dispatchBeforeUnload();

    expect(defaultPrevented).toBe(false);
  });

  it('preventDefaults when any view is dirty', () => {
    renderHook(() => useUnloadGuard());

    act(() => {
      setDirtyDraft('view-a');
    });

    const { defaultPrevented } = dispatchBeforeUnload();

    expect(defaultPrevented).toBe(true);
  });

  it('detaches the listener when state transitions back to clean', () => {
    renderHook(() => useUnloadGuard());

    act(() => {
      setDirtyDraft('view-a');
    });
    act(() => {
      useAppStore.setState({ draftsByView: {} });
    });

    const { defaultPrevented } = dispatchBeforeUnload();

    expect(defaultPrevented).toBe(false);
  });

  it('detaches the listener on unmount', () => {
    const { unmount } = renderHook(() => useUnloadGuard());

    act(() => {
      setDirtyDraft('view-a');
    });
    unmount();

    const { defaultPrevented } = dispatchBeforeUnload();

    expect(defaultPrevented).toBe(false);
  });
});
