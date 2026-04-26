import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId, View, ViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useCanvasDragDrop } from './useCanvasDragDrop';

vi.mock('../../views/hooks/useViews', () => ({
  useView: vi.fn(),
}));

vi.mock('../context/CanvasLayoutContext', () => ({
  useCanvasLayoutContext: () => ({
    positions: {},
    isLoading: false,
    error: null,
    updateComponentPosition: vi.fn(),
    updateCapabilityPosition: vi.fn(),
    batchUpdatePositions: vi.fn(),
    getPositionForElement: () => null,
    refetch: vi.fn(),
  }),
}));

const { useView } = await import('../../views/hooks/useViews');
const mockUseView = vi.mocked(useView);

const editableLinks = {
  self: { href: '/api/v1/views/v1', method: 'GET' as const },
  edit: { href: '/api/v1/views/v1', method: 'PATCH' as const },
};

const baseView = (overrides: Partial<View> = {}): View => ({
  id: 'v1' as ViewId,
  name: 'Test',
  isDefault: true,
  isPrivate: false,
  components: [],
  capabilities: [],
  originEntities: [],
  createdAt: '2024-01-01T00:00:00Z',
  _links: editableLinks,
  ...overrides,
});

function setView(view: View | null) {
  mockUseView.mockReturnValue({
    data: view ?? undefined,
    isLoading: false,
    error: null,
  } as ReturnType<typeof useView>);
}

function resetStore() {
  act(() => {
    useAppStore.setState({
      currentViewId: null,
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicPositions: {},
      dynamicOriginal: null,
      draftsByView: {},
    });
  });
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildDataTransfer(componentId: string): DataTransfer {
  const data: Record<string, string> = { componentId };
  return {
    getData: (key: string) => data[key] ?? '',
    setData: vi.fn(),
    types: Object.keys(data),
    dropEffect: 'copy',
    effectAllowed: 'copy',
  } as unknown as DataTransfer;
}

function buildDropEvent(componentId: string): React.DragEvent {
  return {
    preventDefault: vi.fn(),
    clientX: 100,
    clientY: 200,
    dataTransfer: buildDataTransfer(componentId),
  } as unknown as React.DragEvent;
}

function fakeReactFlowInstance() {
  return {
    screenToFlowPosition: ({ x, y }: { x: number; y: number }) => ({ x, y }),
  } as unknown as Parameters<typeof useCanvasDragDrop>[0];
}

function setupDrop(
  view: View,
  storeState: Partial<ReturnType<typeof useAppStore.getState>>,
) {
  setView(view);
  act(() => {
    useAppStore.setState(storeState);
  });
  const { result } = renderHook(() => useCanvasDragDrop(fakeReactFlowInstance()), { wrapper: createWrapper() });
  return result;
}

function dropOn(result: ReturnType<typeof setupDrop>, componentId: string) {
  act(() => {
    result.current.onDrop(buildDropEvent(componentId));
  });
}

describe('useCanvasDragDrop.onDrop', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  afterEach(() => {
    resetStore();
  });

  it('adds dropped entity to draft when dynamic mode is already active for current view', () => {
    const result = setupDrop(baseView(), {
      currentViewId: 'v1' as ViewId,
      dynamicViewId: 'v1' as ViewId,
      dynamicOriginal: { entities: [], positions: {} },
      dynamicEntities: [],
    });

    dropOn(result, 'c1');

    const state = useAppStore.getState();
    expect(state.dynamicViewId).toBe('v1');
    expect(state.dynamicEntities).toEqual([{ id: 'c1', type: 'component' }]);
  });

  it('enters dynamic mode for current view when dropping before snapshot effect has run (bug 1)', () => {
    const result = setupDrop(baseView(), {
      currentViewId: 'v1' as ViewId,
      dynamicViewId: null,
      dynamicOriginal: null,
      dynamicEntities: [],
    });

    dropOn(result, 'c1');

    const state = useAppStore.getState();
    expect(state.dynamicViewId).toBe('v1');
    expect(state.dynamicEntities).toEqual([{ id: 'c1', type: 'component' }]);
    expect(state.dynamicOriginal).not.toBeNull();
  });

  it('rebases draft to current view if a different view was previously active', () => {
    const result = setupDrop(
      baseView({
        id: 'v1' as ViewId,
        components: [{ componentId: 'existing' as ComponentId, x: 0, y: 0 }],
      }),
      {
        currentViewId: 'v1' as ViewId,
        dynamicViewId: 'v2' as ViewId,
        dynamicOriginal: { entities: [{ id: 'rogue', type: 'component' }], positions: {} },
        dynamicEntities: [{ id: 'rogue', type: 'component' }],
      },
    );

    dropOn(result, 'c1');

    const state = useAppStore.getState();
    expect(state.dynamicViewId).toBe('v1');
    expect(state.dynamicEntities.map((e) => e.id).sort()).toEqual(['c1', 'existing']);
  });

  it('is a no-op when current view is not editable', () => {
    const result = setupDrop(baseView({ _links: { self: editableLinks.self } }), {
      currentViewId: 'v1' as ViewId,
    });

    dropOn(result, 'c1');

    expect(useAppStore.getState().dynamicEntities).toEqual([]);
  });
});
