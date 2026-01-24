import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useCanvasNodes } from './useCanvasNodes';
import type { OriginRelationship, ComponentId } from '../../../api/types';

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

const mockLayoutPositions: Record<string, { x: number; y: number }> = {};

vi.mock('../context/CanvasLayoutContext', () => ({
  useCanvasLayoutContext: () => ({ positions: mockLayoutPositions }),
}));

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: () => ({
    data: [
      { id: 'comp-1', name: 'Component 1' },
      { id: 'comp-2', name: 'Component 2' },
    ],
  }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: [] }),
}));

vi.mock('../../../store/appStore', () => ({
  useAppStore: () => null,
}));

const mockCurrentView = {
  components: [{ componentId: 'comp-1' as ComponentId }, { componentId: 'comp-2' as ComponentId }],
  capabilities: [],
};

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentView: mockCurrentView }),
}));

const mockAcquiredEntities = [
  { id: 'ae-1', name: 'Acquired Entity 1' },
  { id: 'ae-2', name: 'Acquired Entity 2' },
];

vi.mock('../../origin-entities/hooks/useAcquiredEntities', () => ({
  useAcquiredEntitiesQuery: () => ({ data: mockAcquiredEntities }),
}));

vi.mock('../../origin-entities/hooks/useVendors', () => ({
  useVendorsQuery: () => ({ data: [] }),
}));

vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: () => ({ data: [] }),
}));

let mockOriginRelationships: OriginRelationship[] = [];

vi.mock('../../origin-entities/hooks/useOriginRelationships', () => ({
  useOriginRelationshipsQuery: () => ({ data: mockOriginRelationships }),
}));

describe('useCanvasNodes', () => {
  beforeEach(() => {
    Object.keys(mockLayoutPositions).forEach((key) => delete mockLayoutPositions[key]);
    mockOriginRelationships = [];
  });

  describe('origin entity nodes with relationships', () => {
    it('should include origin entity when it has a saved layout position', () => {
      mockLayoutPositions['acq-ae-1'] = { x: 100, y: 100 };
      mockOriginRelationships = [];

      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const acquiredEntityNodes = result.current.filter((n) => n.id.startsWith('acq-'));
      expect(acquiredEntityNodes).toHaveLength(1);
      expect(acquiredEntityNodes[0].id).toBe('acq-ae-1');
    });

    it('should include origin entity when it has a relationship to a component on canvas (even without layout position)', () => {
      mockOriginRelationships = [
        {
          id: 'rel-1' as any,
          componentId: 'comp-1' as ComponentId,
          componentName: 'Component 1',
          relationshipType: 'AcquiredVia',
          originEntityId: 'ae-2',
          originEntityName: 'Acquired Entity 2',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/test', method: 'GET' } },
        },
      ];

      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const acquiredEntityNodes = result.current.filter((n) => n.id.startsWith('acq-'));
      expect(acquiredEntityNodes).toHaveLength(1);
      expect(acquiredEntityNodes[0].id).toBe('acq-ae-2');
    });

    it('should NOT include origin entity when it has no layout position AND no relationship to canvas components', () => {
      mockOriginRelationships = [];

      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const acquiredEntityNodes = result.current.filter((n) => n.id.startsWith('acq-'));
      expect(acquiredEntityNodes).toHaveLength(0);
    });

    it('should NOT include origin entity when relationship is to a component NOT on canvas', () => {
      mockOriginRelationships = [
        {
          id: 'rel-1' as any,
          componentId: 'comp-not-on-canvas' as ComponentId,
          componentName: 'Not On Canvas',
          relationshipType: 'AcquiredVia',
          originEntityId: 'ae-1',
          originEntityName: 'Acquired Entity 1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/test', method: 'GET' } },
        },
      ];

      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const acquiredEntityNodes = result.current.filter((n) => n.id.startsWith('acq-'));
      expect(acquiredEntityNodes).toHaveLength(0);
    });

    it('should include origin entity from both layout position AND relationship (no duplicates)', () => {
      mockLayoutPositions['acq-ae-1'] = { x: 100, y: 100 };
      mockOriginRelationships = [
        {
          id: 'rel-1' as any,
          componentId: 'comp-1' as ComponentId,
          componentName: 'Component 1',
          relationshipType: 'AcquiredVia',
          originEntityId: 'ae-1',
          originEntityName: 'Acquired Entity 1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/test', method: 'GET' } },
        },
      ];

      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const acquiredEntityNodes = result.current.filter((n) => n.id.startsWith('acq-'));
      expect(acquiredEntityNodes).toHaveLength(1);
      expect(acquiredEntityNodes[0].id).toBe('acq-ae-1');
    });

    it('should include multiple origin entities from mixed sources', () => {
      mockLayoutPositions['acq-ae-1'] = { x: 100, y: 100 };
      mockOriginRelationships = [
        {
          id: 'rel-1' as any,
          componentId: 'comp-2' as ComponentId,
          componentName: 'Component 2',
          relationshipType: 'AcquiredVia',
          originEntityId: 'ae-2',
          originEntityName: 'Acquired Entity 2',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/test', method: 'GET' } },
        },
      ];

      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const acquiredEntityNodes = result.current.filter((n) => n.id.startsWith('acq-'));
      expect(acquiredEntityNodes).toHaveLength(2);
      expect(acquiredEntityNodes.map((n) => n.id).sort()).toEqual(['acq-ae-1', 'acq-ae-2']);
    });
  });

  describe('component nodes', () => {
    it('should include components that are in the current view', () => {
      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const componentNodes = result.current.filter(
        (n) => !n.id.startsWith('acq-') && !n.id.startsWith('cap-')
      );
      expect(componentNodes).toHaveLength(2);
      expect(componentNodes.map((n) => n.id).sort()).toEqual(['comp-1', 'comp-2']);
    });
  });
});
