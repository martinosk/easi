import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { CapabilityId, ComponentId, View, ViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useCurrentViewElementIds } from './useCurrentViewElementIds';

vi.mock('./useViews', () => ({
  useView: vi.fn(),
}));

const { useView } = await import('./useViews');
const mockUseView = vi.mocked(useView);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseView = (overrides: Partial<View> = {}): View => ({
  id: 'view-1' as ViewId,
  name: 'Test',
  isDefault: false,
  isPrivate: false,
  components: [],
  capabilities: [],
  originEntities: [],
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
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

interface SeededState {
  view: View;
  currentViewId?: ViewId;
  dynamicViewId?: ViewId;
  dynamicEntities?: { id: string; type: 'component' | 'capability' | 'originEntity' }[];
}

function seedView({ view, currentViewId, dynamicViewId, dynamicEntities }: SeededState) {
  act(() => {
    useAppStore.setState({
      currentViewId: currentViewId ?? null,
      dynamicViewId: dynamicViewId ?? null,
      dynamicEntities: dynamicEntities ?? [],
    });
  });
  setView(view);
}

function renderViewElementIds() {
  return renderHook(() => useCurrentViewElementIds(), { wrapper: createWrapper() });
}

describe('useCurrentViewElementIds', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  afterEach(() => {
    resetStore();
  });

  it('returns empty sets when no view is loaded', () => {
    setView(null);

    const { result } = renderHook(() => useCurrentViewElementIds(), { wrapper: createWrapper() });

    expect(result.current.components.size).toBe(0);
    expect(result.current.capabilities.size).toBe(0);
    expect(result.current.originEntities.size).toBe(0);
  });

  it('returns ids from currentView when no draft is active for it', () => {
    seedView({
      view: baseView({
        id: 'v1' as ViewId,
        components: [{ componentId: 'c1' as ComponentId, x: 0, y: 0 }],
        capabilities: [{ capabilityId: 'cap1' as CapabilityId, x: 0, y: 0 }],
        originEntities: [{ originEntityId: 'oe1', x: 0, y: 0 }],
      }),
      currentViewId: 'v1' as ViewId,
    });

    const { result } = renderViewElementIds();

    expect([...result.current.components]).toEqual(['c1']);
    expect([...result.current.capabilities]).toEqual(['cap1']);
    expect([...result.current.originEntities]).toEqual(['oe1']);
  });

  it('falls back to currentView when dynamic mode is active for a different view', () => {
    seedView({
      view: baseView({
        id: 'v1' as ViewId,
        components: [{ componentId: 'c1' as ComponentId, x: 0, y: 0 }],
      }),
      currentViewId: 'v1' as ViewId,
      dynamicViewId: 'v2' as ViewId,
      dynamicEntities: [{ id: 'rogue', type: 'component' }],
    });

    const { result } = renderViewElementIds();

    expect([...result.current.components]).toEqual(['c1']);
  });

  it('returns draft entities when dynamic mode is active for current view, including additions not in view', () => {
    seedView({
      view: baseView({
        id: 'v1' as ViewId,
        components: [{ componentId: 'c1' as ComponentId, x: 0, y: 0 }],
      }),
      currentViewId: 'v1' as ViewId,
      dynamicViewId: 'v1' as ViewId,
      dynamicEntities: [
        { id: 'c1', type: 'component' },
        { id: 'c2', type: 'component' },
        { id: 'cap-new', type: 'capability' },
      ],
    });

    const { result } = renderViewElementIds();

    expect([...result.current.components].sort()).toEqual(['c1', 'c2']);
    expect([...result.current.capabilities]).toEqual(['cap-new']);
  });

  it('excludes draft removals: entity in view but removed from draft is not in the set', () => {
    seedView({
      view: baseView({
        id: 'v1' as ViewId,
        components: [
          { componentId: 'c1' as ComponentId, x: 0, y: 0 },
          { componentId: 'c2' as ComponentId, x: 0, y: 0 },
        ],
      }),
      currentViewId: 'v1' as ViewId,
      dynamicViewId: 'v1' as ViewId,
      dynamicEntities: [{ id: 'c1', type: 'component' }],
    });

    const { result } = renderViewElementIds();

    expect([...result.current.components]).toEqual(['c1']);
  });
});
