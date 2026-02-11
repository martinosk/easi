import { describe, it, expect, beforeEach } from 'vitest';
import { buildCapability, resetIdCounter } from '../../../test/helpers';
import { toCapabilityId } from '../../../api/types';
import { preserveCapabilityHierarchy } from './preserveCapabilityHierarchy';

describe('preserveCapabilityHierarchy', () => {
  beforeEach(() => {
    resetIdCounter();
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
    const grandparent = buildCapability({
      id: toCapabilityId('cap-gp'),
      name: 'Grandparent',
    });
    const parent = buildCapability({
      id: toCapabilityId('cap-parent'),
      name: 'Parent',
      parentId: toCapabilityId('cap-gp'),
      level: 'L2',
    });
    const grandchild = buildCapability({
      id: toCapabilityId('cap-gc'),
      name: 'Grandchild',
      parentId: toCapabilityId('cap-parent'),
      level: 'L3',
    });

    const allCapabilities = [grandparent, parent, grandchild];
    const filtered = [grandchild];

    const result = preserveCapabilityHierarchy(filtered, allCapabilities);

    expect(result).toContainEqual(grandchild);
    expect(result).toContainEqual(parent);
    expect(result).toContainEqual(grandparent);
    expect(result).toHaveLength(3);
  });

  it('should not duplicate already-included parent capabilities', () => {
    const parent = buildCapability({ id: toCapabilityId('cap-parent'), name: 'Parent' });
    const childA = buildCapability({
      id: toCapabilityId('cap-child-a'),
      name: 'Child A',
      parentId: toCapabilityId('cap-parent'),
      level: 'L2',
    });
    const childB = buildCapability({
      id: toCapabilityId('cap-child-b'),
      name: 'Child B',
      parentId: toCapabilityId('cap-parent'),
      level: 'L2',
    });

    const allCapabilities = [parent, childA, childB];
    const filtered = [parent, childA, childB];

    const result = preserveCapabilityHierarchy(filtered, allCapabilities);

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
    const parent = buildCapability({ id: toCapabilityId('cap-parent'), name: 'Shared Parent' });
    const childA = buildCapability({
      id: toCapabilityId('cap-child-a'),
      name: 'Child A',
      parentId: toCapabilityId('cap-parent'),
      level: 'L2',
    });
    const childB = buildCapability({
      id: toCapabilityId('cap-child-b'),
      name: 'Child B',
      parentId: toCapabilityId('cap-parent'),
      level: 'L2',
    });

    const allCapabilities = [parent, childA, childB];
    const filtered = [childA, childB];

    const result = preserveCapabilityHierarchy(filtered, allCapabilities);

    expect(result).toContainEqual(childA);
    expect(result).toContainEqual(childB);
    expect(result).toContainEqual(parent);
    expect(result).toHaveLength(3);
  });
});
