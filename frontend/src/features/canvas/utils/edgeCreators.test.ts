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

  describe('relationship filtering', () => {
    it('should include edges when both origin entity and component are on canvas', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'AcquiredVia',
        originEntityId: 'ae-123',
        componentId: 'comp-456',
      });
      const nodes = [createMockNode('acq-ae-123'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-123']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges).toHaveLength(1);
    });

    it('should exclude edges when origin entity is not on canvas', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'AcquiredVia',
        originEntityId: 'ae-123',
        componentId: 'comp-456',
      });
      const nodes = [createMockNode('comp-456')];
      const originEntityNodeIds = new Set<string>();
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges).toHaveLength(0);
    });

    it('should exclude edges when component is not on canvas', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'AcquiredVia',
        originEntityId: 'ae-123',
        componentId: 'comp-456',
      });
      const nodes = [createMockNode('acq-ae-123')];
      const originEntityNodeIds = new Set(['acq-ae-123']);
      const componentIds = new Set<string>();
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges).toHaveLength(0);
    });
  });

  describe('relationship type to node ID mapping', () => {
    it('should map AcquiredVia to acq- prefix', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'AcquiredVia',
        originEntityId: 'ae-123',
        componentId: 'comp-456',
      });
      const nodes = [createMockNode('acq-ae-123'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-123']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].target).toBe('acq-ae-123');
    });

    it('should map PurchasedFrom to vendor- prefix', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'PurchasedFrom',
        originEntityId: 'v-456',
        componentId: 'comp-789',
      });
      const nodes = [createMockNode('vendor-v-456'), createMockNode('comp-789')];
      const originEntityNodeIds = new Set(['vendor-v-456']);
      const componentIds = new Set(['comp-789']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].target).toBe('vendor-v-456');
    });

    it('should map BuiltBy to team- prefix', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'BuiltBy',
        originEntityId: 'it-789',
        componentId: 'comp-123',
      });
      const nodes = [createMockNode('team-it-789'), createMockNode('comp-123')];
      const originEntityNodeIds = new Set(['team-it-789']);
      const componentIds = new Set(['comp-123']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].target).toBe('team-it-789');
    });
  });

  describe('edge labels', () => {
    it('should use "Acquired via" label for AcquiredVia relationships', () => {
      const relationship = createMockOriginRelationship({ relationshipType: 'AcquiredVia' });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].label).toBe('Acquired via');
    });

    it('should use "Purchased from" label for PurchasedFrom relationships', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'PurchasedFrom',
        originEntityId: 'v-123',
      });
      const nodes = [createMockNode('vendor-v-123'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['vendor-v-123']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].label).toBe('Purchased from');
    });

    it('should use "Built by" label for BuiltBy relationships', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'BuiltBy',
        originEntityId: 'it-123',
      });
      const nodes = [createMockNode('team-it-123'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['team-it-123']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].label).toBe('Built by');
    });
  });

  describe('edge colors', () => {
    it('should use purple color for AcquiredVia relationships', () => {
      const relationship = createMockOriginRelationship({ relationshipType: 'AcquiredVia' });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].style?.stroke).toBe('#8b5cf6');
    });

    it('should use pink color for PurchasedFrom relationships', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'PurchasedFrom',
        originEntityId: 'v-123',
      });
      const nodes = [createMockNode('vendor-v-123'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['vendor-v-123']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].style?.stroke).toBe('#ec4899');
    });

    it('should use teal color for BuiltBy relationships', () => {
      const relationship = createMockOriginRelationship({
        relationshipType: 'BuiltBy',
        originEntityId: 'it-123',
      });
      const nodes = [createMockNode('team-it-123'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['team-it-123']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].style?.stroke).toBe('#14b8a6');
    });

    it('should use black color in classic scheme', () => {
      const relationship = createMockOriginRelationship({ relationshipType: 'AcquiredVia' });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes, { isClassicScheme: true });

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].style?.stroke).toBe('#000000');
    });
  });

  describe('edge properties', () => {
    it('should create edge with correct ID format', () => {
      const relationship = createMockOriginRelationship({ id: 'rel-999' as OriginRelationshipId });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].id).toBe('origin-rel-999');
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
      const relationship = createMockOriginRelationship();
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes);

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].markerEnd).toEqual({ type: MarkerType.ArrowClosed, color: '#8b5cf6' });
    });

    it('should use specified edge type', () => {
      const relationship = createMockOriginRelationship();
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes, { edgeType: 'smoothstep' });

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].type).toBe('smoothstep');
    });
  });

  describe('edge selection', () => {
    it('should animate edge when selected', () => {
      const relationship = createMockOriginRelationship({ id: 'rel-selected' as OriginRelationshipId });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes, { selectedEdgeId: 'origin-rel-selected' });

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].animated).toBe(true);
    });

    it('should not animate edge when not selected', () => {
      const relationship = createMockOriginRelationship({ id: 'rel-not-selected' as OriginRelationshipId });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes, { selectedEdgeId: 'some-other-edge' });

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].animated).toBe(false);
    });

    it('should use thicker stroke when selected', () => {
      const relationship = createMockOriginRelationship({ id: 'rel-selected' as OriginRelationshipId });
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes, { selectedEdgeId: 'origin-rel-selected' });

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

      expect(edges[0].style?.strokeWidth).toBe(3);
    });

    it('should use normal stroke when not selected', () => {
      const relationship = createMockOriginRelationship();
      const nodes = [createMockNode('acq-ae-789'), createMockNode('comp-456')];
      const originEntityNodeIds = new Set(['acq-ae-789']);
      const componentIds = new Set(['comp-456']);
      const ctx = createEdgeContext(nodes, { selectedEdgeId: null });

      const edges = createOriginRelationshipEdges([relationship], originEntityNodeIds, componentIds, ctx);

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
      expect(edges[0].id).toBe('origin-rel-1');
    });
  });
});
