import { describe, expect, it } from 'vitest';
import { buildCapability } from '../../../test/helpers/entityBuilders';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { buildMapTree } from './mapLayout';

function node(name: string, level: 'L1' | 'L2' | 'L3' | 'L4', children: CapabilityTreeNode[] = []): CapabilityTreeNode {
  return { capability: buildCapability({ name, level }), children };
}

const l3 = node('Deep', 'L3');
const l2b = node('Beta', 'L2', [l3]);
const l2a = node('Alpha', 'L2');
const l1 = node('Zeta Root', 'L1', [l2b, l2a]);
const otherL1 = node('Alpha Root', 'L1');

describe('buildMapTree', () => {
  it('orders L1 roots and nested children alphabetically by name', () => {
    const tree = buildMapTree([l1, otherL1], 4);
    expect(tree.map((n) => n.capability.name)).toEqual(['Alpha Root', 'Zeta Root']);
    expect(tree[1].children.map((n) => n.capability.name)).toEqual(['Alpha', 'Beta']);
  });

  it('prunes all children at depth 1', () => {
    const tree = buildMapTree([l1], 1);
    expect(tree[0].children).toEqual([]);
  });

  it('keeps L2 but prunes L3 at depth 2', () => {
    const tree = buildMapTree([l1], 2);
    const beta = tree[0].children.find((n) => n.capability.name === 'Beta');
    expect(beta).toBeDefined();
    expect(beta?.children).toEqual([]);
  });

  it('keeps the full hierarchy at depth 4', () => {
    const tree = buildMapTree([l1], 4);
    const beta = tree[0].children.find((n) => n.capability.name === 'Beta');
    expect(beta?.children.map((n) => n.capability.name)).toEqual(['Deep']);
  });

  it('does not mutate the input nodes', () => {
    buildMapTree([l1], 1);
    expect(l1.children).toHaveLength(2);
    expect(l1.children[0].capability.name).toBe('Beta');
  });
});
