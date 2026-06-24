import { beforeEach, describe, expect, it } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { buildCapability, resetIdCounter } from '../../../test/helpers';
import { preserveCapabilityHierarchy } from './preserveCapabilityHierarchy';

describe('preserveCapabilityHierarchy', () => {
  beforeEach(() => {
    resetIdCounter();
  });

  const cap = (id: string, name: string, parentId?: string, level: 'L2' | 'L3' = 'L2') =>
    buildCapability({
      id: toCapabilityId(id),
      name,
      ...(parentId ? { parentId: toCapabilityId(parentId), level } : {}),
    });

  it('should return empty array when filtered capabilities is empty', () => {
    const allCapabilities = [
      buildCapability({ id: toCapabilityId('cap-root'), name: 'Root' }),
      buildCapability({
        id: toCapabilityId('cap-child'),
        name: 'Child',
        parentId: toCapabilityId('cap-root'),
        level: 'L2',
      }),
    ];

    const result = preserveCapabilityHierarchy([], allCapabilities);

    expect(result).toEqual([]);
  });

  it('should return filtered as-is when all are root capabilities', () => {
    const rootA = buildCapability({ id: toCapabilityId('cap-a'), name: 'Root A' });
    const rootB = buildCapability({ id: toCapabilityId('cap-b'), name: 'Root B' });

    const allCapabilities = [rootA, rootB];

    const result = preserveCapabilityHierarchy([rootA, rootB], allCapabilities);

    expect(result).toEqual([rootA, rootB]);
  });

  it('should add parent capability when child matches but parent does not', () => {
    const parent = buildCapability({ id: toCapabilityId('cap-parent'), name: 'Parent Capability' });
    const child = buildCapability({
      id: toCapabilityId('cap-child'),
      name: 'Child Capability',
      parentId: toCapabilityId('cap-parent'),
      level: 'L2',
    });

    const allCapabilities = [parent, child];
    const filtered = [child];

    const result = preserveCapabilityHierarchy(filtered, allCapabilities);

    expect(result).toContainEqual(child);
    expect(result).toContainEqual(parent);
    expect(result).toHaveLength(2);
  });

  it('should add grandparent when only grandchild matches', () => {
    const grandparent = cap('cap-gp', 'Grandparent');
    const parent = cap('cap-parent', 'Parent', 'cap-gp');
    const grandchild = cap('cap-gc', 'Grandchild', 'cap-parent', 'L3');

    const result = preserveCapabilityHierarchy([grandchild], [grandparent, parent, grandchild]);

    expect(result).toContainEqual(grandchild);
    expect(result).toContainEqual(parent);
    expect(result).toContainEqual(grandparent);
    expect(result).toHaveLength(3);
  });

  it('should not duplicate already-included parent capabilities', () => {
    const parent = cap('cap-parent', 'Parent');
    const childA = cap('cap-child-a', 'Child A', 'cap-parent');
    const childB = cap('cap-child-b', 'Child B', 'cap-parent');

    const all = [parent, childA, childB];

    const result = preserveCapabilityHierarchy(all, all);

    const parentOccurrences = result.filter((c) => c.id === parent.id);
    expect(parentOccurrences).toHaveLength(1);
    expect(result).toHaveLength(3);
  });

  it('should preserve original filtered capabilities plus structural parents only', () => {
    const root = buildCapability({ id: toCapabilityId('cap-root'), name: 'Root' });
    const siblingA = buildCapability({
      id: toCapabilityId('cap-sibling-a'),
      name: 'Sibling A',
      parentId: toCapabilityId('cap-root'),
      level: 'L2',
    });
    const siblingB = buildCapability({
      id: toCapabilityId('cap-sibling-b'),
      name: 'Sibling B',
      parentId: toCapabilityId('cap-root'),
      level: 'L2',
    });
    const unrelatedRoot = buildCapability({
      id: toCapabilityId('cap-unrelated'),
      name: 'Unrelated Root',
    });

    const allCapabilities = [root, siblingA, siblingB, unrelatedRoot];
    const filtered = [siblingA];

    const result = preserveCapabilityHierarchy(filtered, allCapabilities);

    expect(result).toContainEqual(siblingA);
    expect(result).toContainEqual(root);
    expect(result).not.toContainEqual(siblingB);
    expect(result).not.toContainEqual(unrelatedRoot);
    expect(result).toHaveLength(2);
  });

  it('should not duplicate parent when two children under same parent both match', () => {
    const parent = cap('cap-parent', 'Shared Parent');
    const childA = cap('cap-child-a', 'Child A', 'cap-parent');
    const childB = cap('cap-child-b', 'Child B', 'cap-parent');

    const result = preserveCapabilityHierarchy([childA, childB], [parent, childA, childB]);

    expect(result).toContainEqual(childA);
    expect(result).toContainEqual(childB);
    expect(result).toContainEqual(parent);
    expect(result).toHaveLength(3);
  });
});
