import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useCanvasEdges } from './useCanvasEdges';
import type { Node } from '@xyflow/react';
import { toComponentId, toOriginRelationshipId, type OriginRelationship } from '../../../api/types';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: [] }),
  useRealizationsForComponents: () => ({ data: [] }),
}));

vi.mock('../../relations/hooks/useRelations', () => ({
  useRelations: () => ({ data: [] }),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({
    currentView: {
      components: [{ componentId: 'comp-1' }],
      capabilities: [],
      edgeType: 'default',
      colorScheme: 'maturity',
    },
  }),
}));

vi.mock('../../../store/appStore', () => ({
  useAppStore: () => null,
}));

vi.mock('../../origin-entities/hooks/useOriginRelationships', () => ({
  useOriginRelationshipsQuery: vi.fn(() => ({ data: [] })),
}));

describe('useCanvasEdges - Origin Entity Edge Bug', () => {
  type OriginRelationshipsQueryResult = ReturnType<
    (typeof import('../../origin-entities/hooks/useOriginRelationships'))['useOriginRelationshipsQuery']
  >;

  function createOriginRelationshipsResult(
    data: OriginRelationship[]
  ): OriginRelationshipsQueryResult {
    return { data } as OriginRelationshipsQueryResult;
  }

  /**
   * BUG REPRODUCTION: Origin entity edge not appearing after linking
   *
   * Scenario:
   * 1. Origin entity node IS on canvas (has layout position)
   * 2. Component node IS on canvas
   * 3. User links component to origin entity
   * 4. Backend creates relationship
   * 5. Frontend mutation invalidates queryKeys.originRelationships.lists()
   * 6. Query refetches new relationship data
   *
   * Expected: Edge appears connecting component → origin entity
   * Actual: Edge does NOT appear
   *
   * This test will prove whether the issue is:
   * a) Query not refetching (cache invalidation bug)
   * b) Query refetches but edges don't update (React Query timing/closure bug)
   * c) Edge filtering logic bug
   */
  it('BUG: should show edge immediately after linking when both nodes are on canvas', async () => {
    const originRelationshipsModule = await import(
      '../../origin-entities/hooks/useOriginRelationships'
    );
    const mockedUseOriginRelationshipsQuery = vi.mocked(originRelationshipsModule.useOriginRelationshipsQuery);

    // Setup: Both nodes are on canvas
    const nodes: Node[] = [
      {
        id: 'comp-1',
        position: { x: 0, y: 0 },
        data: {},
      },
      {
        id: 'acq-ae-123', // Origin entity node
        position: { x: 200, y: 0 },
        data: {},
      },
    ];

    // Initially: No relationships
    mockedUseOriginRelationshipsQuery.mockReturnValue(createOriginRelationshipsResult([]));

    const { result, rerender } = renderHook(() => useCanvasEdges(nodes), {
      wrapper: createWrapper(),
    });

    // Initially: No origin edges
    const initialOriginEdges = result.current.filter((e) => e.id.startsWith('origin-'));
    expect(initialOriginEdges).toHaveLength(0);

    // User links component to origin entity
    // Backend creates relationship
    // Mutation invalidates cache
    // Query refetches and returns new data
    const newRelationship: OriginRelationship = {
      id: toOriginRelationshipId('rel-123'),
      componentId: toComponentId('comp-1'),
      componentName: 'SAP HR',
      relationshipType: 'AcquiredVia',
      originEntityId: 'ae-123',
      originEntityName: 'TechCorp',
      createdAt: '2024-01-01T00:00:00Z',
      _links: { self: { href: '/test', method: 'GET' } },
    };

    mockedUseOriginRelationshipsQuery.mockReturnValue(createOriginRelationshipsResult([newRelationship]));

    // Trigger re-render (simulating React Query refetch)
    rerender();

    // Expected: Edge should appear
    await waitFor(() => {
      const originEdges = result.current.filter((e) => e.id.startsWith('origin-'));
      expect(originEdges).toHaveLength(1);
    });

    const originEdge = result.current.find((e) => e.id === 'origin-AcquiredVia-comp-1');
    expect(originEdge).toBeDefined();
    expect(originEdge?.source).toBe('comp-1');
    expect(originEdge?.target).toBe('acq-ae-123');
    expect(originEdge?.label).toBe('Acquired via');
  });
});
