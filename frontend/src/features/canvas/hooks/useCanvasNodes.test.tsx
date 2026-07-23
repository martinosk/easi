import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { type ComponentId, type OriginRelationship, toOriginRelationshipId } from '../../../api/types';
import { useCanvasNodes } from './useCanvasNodes';

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

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: () => ({
    data: [
      { id: 'comp-1', name: 'Component 1' },
      { id: 'comp-2', name: 'Component 2' },
    ],
  }),
}));

let mockCapabilities: { id: string; name: string; level: number }[] = [];

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: mockCapabilities }),
}));

vi.mock('../../../store/appStore', () => ({
  useAppStore: () => null,
}));

const mockCurrentView: {
  components: { componentId: ComponentId; x?: number; y?: number }[];
  capabilities: { capabilityId: string; x: number; y: number }[];
  originEntities: { originEntityId: string; x: number; y: number }[];
} = {
  components: [],
  capabilities: [],
  originEntities: [],
};

function defaultViewComponents(): { componentId: ComponentId; x?: number; y?: number }[] {
  return [
    { componentId: 'comp-1' as ComponentId, x: 10, y: 20 },
    { componentId: 'comp-2' as ComponentId, x: 30, y: 40 },
  ];
}

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

function createMockRelationship(
  componentId: string,
  originEntityId: string,
  originEntityName: string,
): OriginRelationship {
  return {
    id: toOriginRelationshipId(`rel-${originEntityId}`),
    componentId: componentId as ComponentId,
    componentName: `Component ${componentId}`,
    relationshipType: 'AcquiredVia',
    originEntityId,
    originEntityName,
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: '/test', method: 'GET' } },
  };
}

function addOriginEntityToView(originEntityId: string, x = 100, y = 100) {
  mockCurrentView.originEntities.push({ originEntityId, x, y });
}

function renderAndGetAcquiredEntityNodes() {
  const { result } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });
  return result.current.filter((n) => n.id.startsWith('acq-'));
}

describe('useCanvasNodes', () => {
  beforeEach(() => {
    mockOriginRelationships = [];
    mockCurrentView.originEntities = [];
    mockCurrentView.components = defaultViewComponents();
    mockCurrentView.capabilities = [];
    mockCapabilities = [];
  });

  describe('origin entity nodes with relationships', () => {
    it('should include origin entity when it is explicitly added to the view', () => {
      addOriginEntityToView('ae-1');

      const nodes = renderAndGetAcquiredEntityNodes();

      expect(nodes).toHaveLength(1);
      expect(nodes[0].id).toBe('acq-ae-1');
    });

    it('should NOT include origin entity when it only has a relationship (not explicitly added to view)', () => {
      mockOriginRelationships = [createMockRelationship('comp-1', 'ae-2', 'Acquired Entity 2')];

      expect(renderAndGetAcquiredEntityNodes()).toHaveLength(0);
    });

    it('should NOT include origin entity when it has no layout position AND no relationship to canvas components', () => {
      expect(renderAndGetAcquiredEntityNodes()).toHaveLength(0);
    });

    it('should NOT include origin entity when relationship is to a component NOT on canvas', () => {
      mockOriginRelationships = [createMockRelationship('comp-not-on-canvas', 'ae-1', 'Acquired Entity 1')];

      expect(renderAndGetAcquiredEntityNodes()).toHaveLength(0);
    });

    it('should include origin entity from view membership even with relationship (no duplicates)', () => {
      addOriginEntityToView('ae-1');
      mockOriginRelationships = [createMockRelationship('comp-1', 'ae-1', 'Acquired Entity 1')];

      const nodes = renderAndGetAcquiredEntityNodes();

      expect(nodes).toHaveLength(1);
      expect(nodes[0].id).toBe('acq-ae-1');
    });

    it('should only include origin entity explicitly in view, ignoring relationship-only entities', () => {
      addOriginEntityToView('ae-1');
      mockOriginRelationships = [createMockRelationship('comp-2', 'ae-2', 'Acquired Entity 2')];

      const nodes = renderAndGetAcquiredEntityNodes();

      expect(nodes).toHaveLength(1);
      expect(nodes[0].id).toBe('acq-ae-1');
    });
  });

  describe('component nodes', () => {
    it('should include components that are in the current view', () => {
      const { result } = renderHook(() => useCanvasNodes(), {
        wrapper: createWrapper(),
      });

      const componentNodes = result.current.filter((n) => !n.id.startsWith('acq-') && !n.id.startsWith('cap-'));
      expect(componentNodes).toHaveLength(2);
      expect(componentNodes.map((n) => n.id).sort()).toEqual(['comp-1', 'comp-2']);
    });
  });

  describe('node positions outside draft mode', () => {
    const positionOf = (id: string) => {
      const { result } = renderHook(() => useCanvasNodes(), { wrapper: createWrapper() });
      return result.current.find((n) => n.id === id)?.position;
    };

    it('should position components from view coordinates', () => {
      expect(positionOf('comp-1')).toEqual({ x: 10, y: 20 });
      expect(positionOf('comp-2')).toEqual({ x: 30, y: 40 });
    });

    it('should position capabilities from view coordinates', () => {
      mockCapabilities = [{ id: 'cap-1', name: 'Capability 1', level: 1 }];
      mockCurrentView.capabilities = [{ capabilityId: 'cap-1', x: 55, y: 66 }];

      expect(positionOf('cap-cap-1')).toEqual({ x: 55, y: 66 });
    });

    it('should position origin entities from view coordinates', () => {
      addOriginEntityToView('ae-1', 120, 130);

      expect(positionOf('acq-ae-1')).toEqual({ x: 120, y: 130 });
    });

    it('should place an element with no view coordinates at the canvas default position', () => {
      mockCurrentView.components = [{ componentId: 'comp-1' as ComponentId }];

      expect(positionOf('comp-1')).toEqual({ x: 400, y: 300 });
    });
  });
});
