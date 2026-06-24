import { describe, expect, it } from 'vitest';
import type { HATEOASLinks } from '../../../api/types';
import { computeAvailableActions } from './useMultiSelectContextMenu';
import type { NodeContextMenu } from './useNodeContextMenu';

function makeNode(overrides: Partial<NodeContextMenu> = {}): NodeContextMenu {
  const nodeId = overrides.nodeId ?? 'test-id';
  return {
    x: 100,
    y: 200,
    nodeId,
    nodeName: 'Test Node',
    nodeType: 'component',
    ...overrides,
  };
}

function linksWithRemoveAndDelete(): { modelLinks: HATEOASLinks; viewElementLinks: HATEOASLinks } {
  return {
    modelLinks: { delete: { href: '/api/v1/components/1', method: 'DELETE' } },
    viewElementLinks: { 'x-remove': { href: '/api/v1/views/1/components/1', method: 'DELETE' } },
  };
}

function linksWithRemoveOnly(): { modelLinks: HATEOASLinks; viewElementLinks: HATEOASLinks } {
  return {
    modelLinks: {},
    viewElementLinks: { 'x-remove': { href: '/api/v1/views/1/components/1', method: 'DELETE' } },
  };
}

function linksWithDeleteOnly(): { modelLinks: HATEOASLinks; viewElementLinks: HATEOASLinks } {
  return {
    modelLinks: { delete: { href: '/api/v1/components/1', method: 'DELETE' } },
    viewElementLinks: {},
  };
}

function noLinks(): { modelLinks: HATEOASLinks; viewElementLinks: HATEOASLinks } {
  return { modelLinks: {}, viewElementLinks: {} };
}

type NodeLinks = { modelLinks: HATEOASLinks; viewElementLinks: HATEOASLinks };

function twoNodes(first: NodeLinks, second: NodeLinks): NodeContextMenu[] {
  return [makeNode({ nodeId: '1', ...first }), makeNode({ nodeId: '2', ...second })];
}

describe('computeAvailableActions', () => {
  it('returns empty for fewer than 2 nodes', () => {
    const single = [makeNode(linksWithRemoveAndDelete())];
    expect(computeAvailableActions(single)).toEqual([]);
  });

  it('returns empty for empty array', () => {
    expect(computeAvailableActions([])).toEqual([]);
  });

  it('returns both actions when all nodes have remove and delete', () => {
    const nodes = [
      makeNode({ nodeId: '1', ...linksWithRemoveAndDelete() }),
      makeNode({ nodeId: '2', ...linksWithRemoveAndDelete() }),
      makeNode({ nodeId: '3', ...linksWithRemoveAndDelete() }),
    ];
    const actions = computeAvailableActions(nodes);
    expect(actions).toHaveLength(2);
    expect(actions[0]).toEqual({
      type: 'removeFromView',
      label: 'Remove from View (3 items)',
      isDanger: false,
    });
    expect(actions[1]).toEqual({
      type: 'deleteFromModel',
      label: 'Delete from Model (3 items)',
      isDanger: true,
    });
  });

  it('returns only removeFromView when all can remove but not all can delete', () => {
    const actions = computeAvailableActions(twoNodes(linksWithRemoveAndDelete(), linksWithRemoveOnly()));
    expect(actions).toHaveLength(1);
    expect(actions[0].type).toBe('removeFromView');
    expect(actions[0].label).toBe('Remove from View (2 items)');
  });

  it('returns only deleteFromModel when all can delete but not all can remove', () => {
    const actions = computeAvailableActions(twoNodes(linksWithRemoveAndDelete(), linksWithDeleteOnly()));
    expect(actions).toHaveLength(1);
    expect(actions[0].type).toBe('deleteFromModel');
  });

  it('returns empty when no common actions', () => {
    const actions = computeAvailableActions(twoNodes(linksWithRemoveOnly(), linksWithDeleteOnly()));
    expect(actions).toHaveLength(0);
  });

  it('returns empty when one node has no links', () => {
    const actions = computeAvailableActions(twoNodes(linksWithRemoveAndDelete(), noLinks()));
    expect(actions).toHaveLength(0);
  });

  it('handles mixed node types', () => {
    const nodes = [
      makeNode({ nodeId: '1', nodeType: 'component', ...linksWithRemoveAndDelete() }),
      makeNode({ nodeId: '2', nodeType: 'capability', ...linksWithRemoveAndDelete() }),
      makeNode({ nodeId: '3', nodeType: 'originEntity', originEntityType: 'vendor', ...linksWithRemoveAndDelete() }),
    ];
    const actions = computeAvailableActions(nodes);
    expect(actions).toHaveLength(2);
  });
});
