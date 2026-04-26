import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useAppStore } from '../../../store/appStore';
import { useCanvasNodes } from './useCanvasNodes';
import { useCanvasEdges } from './useCanvasEdges';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

const mockLayoutPositions: Record<string, { x: number; y: number }> = {};

vi.mock('../context/CanvasLayoutContext', () => ({
  useCanvasLayoutContext: () => ({ positions: mockLayoutPositions }),
}));

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: () => ({
    data: [
      { id: 'comp-1', name: 'Order Service', _links: { self: { href: '/c/1', method: 'GET' } } },
      { id: 'comp-2', name: 'Payment Service', _links: { self: { href: '/c/2', method: 'GET' } } },
      { id: 'comp-3', name: 'Notification Service', _links: { self: { href: '/c/3', method: 'GET' } } },
    ],
  }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({
    data: [{ id: 'cap-1', name: 'Capability 1', _links: { self: { href: '/cap/1', method: 'GET' } } }],
  }),
  useRealizations: () => ({ data: [] }),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({
    currentView: {
      id: 'view-1',
      components: [{ componentId: 'comp-1', x: 0, y: 0 }],
      capabilities: [],
      originEntities: [],
      edgeType: 'default',
      colorScheme: 'maturity',
    },
    currentViewId: 'view-1',
  }),
}));

vi.mock('../../origin-entities/hooks/useAcquiredEntities', () => ({
  useAcquiredEntitiesQuery: () => ({ data: [] }),
}));
vi.mock('../../origin-entities/hooks/useVendors', () => ({
  useVendorsQuery: () => ({ data: [] }),
}));
vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: () => ({ data: [] }),
}));
vi.mock('../../origin-entities/hooks/useOriginRelationships', () => ({
  useOriginRelationshipsQuery: () => ({ data: [] }),
}));

vi.mock('../../relations/hooks/useRelations', () => ({
  useRelations: () => ({
    data: [
      {
        id: 'rel-1-2',
        sourceComponentId: 'comp-1',
        targetComponentId: 'comp-2',
        relationType: 'Triggers',
        name: 'calls',
      },
      {
        id: 'rel-2-3',
        sourceComponentId: 'comp-2',
        targetComponentId: 'comp-3',
        relationType: 'Triggers',
        name: 'notifies',
      },
    ],
  }),
}));

function seedDraft(state: Partial<ReturnType<typeof useAppStore.getState>>) {
  act(() => {
    useAppStore.setState(state);
  });
}

describe('Canvas in dynamic mode', () => {
  beforeEach(() => {
    Object.keys(mockLayoutPositions).forEach((key) => delete mockLayoutPositions[key]);
    seedDraft({
      dynamicViewId: 'view-1',
      dynamicEntities: [
        { id: 'comp-1', type: 'component' },
        { id: 'comp-2', type: 'component' },
      ],
      dynamicPositions: {
        'comp-1': { x: 0, y: 0 },
        'comp-2': { x: 200, y: 0 },
      },
      dynamicOriginal: {
        entities: [{ id: 'comp-1', type: 'component' }],
        positions: { 'comp-1': { x: 0, y: 0 } },
      },
    });
  });

  afterEach(() => {
    seedDraft({
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicPositions: {},
      dynamicOriginal: null,
    });
  });

  it('renders drafted entities not yet in the persisted view', () => {
    const { result } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });

    const ids = result.current.map((n) => n.id).sort();
    expect(ids).toEqual(['comp-1', 'comp-2']);
  });

  it('omits entities that have been removed from the draft', () => {
    seedDraft({
      dynamicEntities: [{ id: 'comp-2', type: 'component' }],
      dynamicPositions: { 'comp-2': { x: 200, y: 0 } },
    });

    const { result } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });

    const ids = result.current.map((n) => n.id);
    expect(ids).toEqual(['comp-2']);
  });

  it('uses draft positions over layout positions', () => {
    mockLayoutPositions['comp-1'] = { x: 0, y: 0 };
    seedDraft({
      dynamicPositions: { 'comp-1': { x: 999, y: 999 } },
    });

    const { result } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });
    const node = result.current.find((n) => n.id === 'comp-1');
    expect(node?.position).toEqual({ x: 999, y: 999 });
  });

  it('produces edges between drafted components even when they are not yet in currentView', () => {
    const { result: nodesRes } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });
    const { result: edgesRes } = renderHook(() => useCanvasEdges(nodesRes.current), {
      wrapper: createWrapper(),
    });

    const relationEdge = edgesRes.current.find((e) => e.id === 'rel-1-2');
    expect(relationEdge).toBeDefined();
    expect(relationEdge?.source).toBe('comp-1');
    expect(relationEdge?.target).toBe('comp-2');
  });

  it('does not produce relation edges that involve a non-drafted component', () => {
    const { result: nodesRes } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });
    const { result: edgesRes } = renderHook(() => useCanvasEdges(nodesRes.current), {
      wrapper: createWrapper(),
    });

    const orphanEdge = edgesRes.current.find((e) => e.id === 'rel-2-3');
    expect(orphanEdge).toBeUndefined();
  });
});
