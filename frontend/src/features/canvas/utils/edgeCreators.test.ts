import { describe, it, expect } from 'vitest';
import type { Node } from '@xyflow/react';
import { MarkerType } from '@xyflow/react';
import { createOriginRelationshipEdges, type EdgeCreationContext } from './edgeCreators';
import type { OriginRelationship, OriginRelationshipId, ComponentId, HATEOASLinks } from '../../../api/types';

describe('createOriginRelationshipEdges', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createMockOriginRelationship = (overrides = {}): OriginRelationship => ({
    id: 'rel-123' as OriginRelationshipId,
    componentId: 'comp-456' as ComponentId,
    componentName: 'SAP HR',
    relationshipType: 'AcquiredVia',
    originEntityId: 'ae-789',
    originEntityName: 'TechCorp',
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  const createMockNode = (id: string, position = { x: 0, y: 0 }): Node => ({
    id,
    position,
    data: {},
  });

  const createEdgeContext = (nodes: Node[], overrides = {}): EdgeCreationContext => ({
    nodes,
    selectedEdgeId: null,
    edgeType: 'default',
    isClassicScheme: false,
    ...overrides,
  });

  const callWithDefaults = (ctxOverrides = {}) => {
    const relationship = createMockOriginRelationship();
    const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
    const originEntityNodeIds = new Set(['acq-ae-789']);
    const componentIds = new Set(['comp-456']);
    const ctx = createEdgeContext(nodes, ctxOverrides);
    return createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);
  };

  const relationshipTypeConfig = [
    { type: 'AcquiredVia', prefix: 'acq', entityId: 'ae-123', label: 'Acquired via', color: '#8b5cf6' },
    { type: 'PurchasedFrom', prefix: 'vendor', entityId: 'v-456', label: 'Purchased from', color: '#ec4899' },
    { type: 'BuiltBy', prefix: 'team', entityId: 'it-789', label: 'Built by', color: '#14b8a6' },
  ] as const;

  const callWithRelType = (config: typeof relationshipTypeConfig[number], ctxOverrides = {}) => {
    const relationship = createMockOriginRelationship({
      relationshipType: config.type,
      originEntityId: config.entityId,
      componentId: 'comp-456',
    });
    const nodeId = `${config.prefix}-${config.entityId}`;
    const nodes = [createMockNode(nodeId), createMockNode('comp-456')];
    const originEntityNodeIds = new Set([nodeId]);
    const componentIds = new Set(['comp-456']);
    const ctx = createEdgeContext(nodes, ctxOverrides);
    return createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);
  };

  describe('relationship filtering', () => {
    interface FilteringScenario {
      nodeIds: string[];
      originEntityNodeIds: string[];
      componentIds: string[];
      relOverrides?: Record<string, unknown>;
    }

    const callForFiltering = (scenario: FilteringScenario) => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'AcquiredVia',
        originEntityId: 'ae-123',
        componentId: 'comp-456',
        ...scenario.relOverrides,
      });
      const nodes = scenario.nodeIds.map(id => createMockNode(id));
      const ctx = createEdgeContext(nodes);
      return createOriginRelationshipEdges(
        [relationship],
        new Set(scenario.originEntityNodeIds),
        new Set(scenario.componentIds),
        ctx,
      );
    };

    it('should include edges when both origin entity and component are on canvas', () => {
      const edges = callForFiltering({
        nodeIds: ['acq-ae-123', 'comp-456'],
        originEntityNodeIds: ['acq-ae-123'],
        componentIds: ['comp-456'],
      });
      expect(edges).toHaveLength(1);
    });

    it('should exclude edges when origin entity is not on canvas', () => {
      const edges = callForFiltering({
        nodeIds: ['comp-456'],
        originEntityNodeIds: [],
        componentIds: ['comp-456'],
      });
      expect(edges).toHaveLength(0);
    });

    /**
     * Edge case: relationship exists but origin entity node is not on canvas
     *
     * This can happen if:
     * - Origin entity was removed from canvas but relationship still exists
     * - Edge creator correctly filters these out
     *
     * Note: This was previously a bug where origin entities wouldn't show on canvas
     * after creating a relationship because they had no saved layout position.
     * Fixed in useCanvasNodes.ts by also including origin entities that have
     * relationships to components on the canvas (with fallback positioning).
     */
    it('should not create edge when origin entity node is not on canvas', () => {
      const edges = callForFiltering({
        nodeIds: ['comp-456'],
        originEntityNodeIds: [],
        componentIds: ['comp-456'],
        relOverrides: { originEntityId: 'ae-new-entity' },
      });
      expect(edges).toHaveLength(0);
    });

    it('should exclude edges when component is not on canvas', () => {
      const edges = callForFiltering({
        nodeIds: ['acq-ae-123'],
        originEntityNodeIds: ['acq-ae-123'],
        componentIds: [],
      });
      expect(edges).toHaveLength(0);
    });
  });

  describe('relationship type to node ID mapping', () => {
    it.each(relationshipTypeConfig)(
      'should map $type to $prefix- prefix',
      (config) => {
        const edges = callWithRelType(config);
        expect(edges[0].target).toBe(`${config.prefix}-${config.entityId}`);
      }
    );
  });

  describe('edge labels', () => {
    it.each(relationshipTypeConfig)(
      'should use "$label" label for $type relationships',
      (config) => {
        const edges = callWithRelType(config);
        expect(edges[0].label).toBe(config.label);
      }
    );
  });

  describe('edge colors', () => {
    it.each(relationshipTypeConfig)(
      'should use correct color for $type relationships',
      (config) => {
        const edges = callWithRelType(config);
        expect(edges[0].style?.stroke).toBe(config.color);
      }
    );

    it('should use black color in classic scheme', () => {
      const edges = callWithDefaults({ isClassicScheme: true });
      expect(edges[0].style?.stroke).toBe('#000000');
    });
  });

  describe('edge properties', () => {
    it('should create edge with correct ID format', () => {
      const edges = callWithDefaults();
      expect(edges[0].id).toBe('origin-AcquiredVia-comp-456');
    });

    it('should set source to component ID', () => {
      const relationship = createMockOriginRelationship({ componentId: 'comp-target' as ComponentId });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-target')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-target']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].source).toBe('comp-target');
    });

    it('should include arrow marker at end', () => {
      const edges = callWithDefaults();
      expect(edges[0].markerEnd).toEqual({ type: MarkerType.ArrowClosed, color: '#8b5cf6' });
    });

    it('should use specified edge type', () => {
      const edges = callWithDefaults({ edgeType: 'smoothstep' });
      expect(edges[0].type).toBe('smoothstep');
    });
  });

  describe('edge selection', () => {
    it('should animate edge when selected', () => {
      const edges = callWithDefaults({ selectedEdgeId: 'origin-AcquiredVia-comp-456' });
      expect(edges[0].animated).toBe(true);
    });

    it('should not animate edge when not selected', () => {
      const edges = callWithDefaults({ selectedEdgeId: 'some-other-edge' });
      expect(edges[0].animated).toBe(false);
    });

    it('should use thicker stroke when selected', () => {
      const edges = callWithDefaults({ selectedEdgeId: 'origin-AcquiredVia-comp-456' });
      expect(edges[0].style?.strokeWidth).toBe(3);
    });

    it('should use normal stroke when not selected', () => {
      const edges = callWithDefaults({ selectedEdgeId: null });
      expect(edges[0].style?.strokeWidth).toBe(2);
    });
  });

  describe('multiple relationships', () => {
    it('should create edges for all visible relationships', () => {
      const relationships = [
        createMockOriginRelationship({
          id: 'rel-1' as OriginRelationshipId,
          relationshipType: 'AcquiredVia',
          originEntityId: 'ae-1',
          componentId: 'comp-1' as ComponentId,
        }),
        createMockOriginRelationship({
          id: 'rel-2' as OriginRelationshipId,
          relationshipType: 'PurchasedFrom',
          originEntityId: 'v-1',
          componentId: 'comp-1' as ComponentId,
        }),
        createMockOriginRelationship({
          id: 'rel-3' as OriginRelationshipId,
          relationshipType: 'BuiltBy',
          originEntityId: 'it-1',
          componentId: 'comp-1' as ComponentId,
        }),
      ];
      const nodes = [
        createMockNode('acq-ae-1'),
        createMockNode('vendor-v-1'),
        createMockNode('team-it-1'),
        createMockNode('comp-1'),
      ];
      const originEntityNodeIds = new Set(['acq-ae-1', 'vendor-v-1', 'team-it-1']);
      const componentIds = new Set(['comp-1']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges(relationships, originEntityNodeIds, componentIds, ctx);

      expect(edges).toHaveLength(3);
      expect(edges[0].label).toBe('Acquired via');
      expect(edges[1].label).toBe('Purchased from');
      expect(edges[2].label).toBe('Built by');
    });

    it('should only create edges for relationships where both nodes are on canvas', () => {
      const relationships = [
        createMockOriginRelationship({
          id: 'rel-1' as OriginRelationshipId,
          relationshipType: 'AcquiredVia',
          originEntityId: 'ae-1',
          componentId: 'comp-1' as ComponentId,
        }),
        createMockOriginRelationship({
          id: 'rel-2' as OriginRelationshipId,
          relationshipType: 'PurchasedFrom',
          originEntityId: 'v-1',
          componentId: 'comp-2' as ComponentId,
        }),
      ];
      const nodes = [createMockNode('acq-ae-1'), createMockNode('comp-1')];
      const originEntityNodeIds = new Set(['acq-ae-1']);
      const componentIds = new Set(['comp-1']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges(relationships, originEntityNodeIds, componentIds, ctx);

      expect(edges).toHaveLength(1);
      expect(edges[0].id).toBe('origin-AcquiredVia-comp-1');
    });
  });
});
