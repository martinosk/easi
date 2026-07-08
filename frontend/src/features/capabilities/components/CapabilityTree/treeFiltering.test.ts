import { describe, expect, it } from 'vitest';
import { buildCapabilityAt as cap } from '../../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../hooks/useCapabilityTree';
import { collectExpandableIds, filterTreeByName } from './treeFiltering';

const tree = buildCapabilityTree([
  cap('a', 'Customer Management', 'L1'),
  cap('a1', 'Onboarding', 'L2', 'a'),
  cap('a1x', 'Identity Checks', 'L3', 'a1'),
  cap('b', 'Billing', 'L1'),
  cap('b1', 'Invoicing', 'L2', 'b'),
]);

describe('filterTreeByName', () => {
  it('returns the tree unchanged for a blank query', () => {
    expect(filterTreeByName(tree, '  ')).toBe(tree);
  });

  it('keeps matching nodes with their ancestors', () => {
    const filtered = filterTreeByName(tree, 'identity');

    expect(filtered.map((n) => n.capability.name)).toEqual(['Customer Management']);
    expect(filtered[0].children.map((n) => n.capability.name)).toEqual(['Onboarding']);
    expect(filtered[0].children[0].children.map((n) => n.capability.name)).toEqual(['Identity Checks']);
  });

  it('matches case-insensitively on name only, ignoring descriptions', () => {
    const billing = tree.find((n) => n.capability.name === 'Billing')!;
    const withDescription = tree.map((n) =>
      n === billing ? { ...n, capability: { ...n.capability, description: 'identity handling' }, children: [] } : n,
    );

    const filtered = filterTreeByName(withDescription, 'IDENTITY');

    expect(filtered.map((n) => n.capability.name)).toEqual(['Customer Management']);
  });

  it('drops non-matching children of a matching node', () => {
    const filtered = filterTreeByName(tree, 'billing');

    expect(filtered.map((n) => n.capability.name)).toEqual(['Billing']);
    expect(filtered[0].children).toEqual([]);
  });
});

describe('collectExpandableIds', () => {
  it('collects ids of every node that has children', () => {
    expect(collectExpandableIds(tree)).toEqual(new Set(['a', 'a1', 'b']));
  });
});
