import { describe, expect, it } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { buildCapabilityRealization } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { nodeMatchesSearch } from './boardSearch';

describe('nodeMatchesSearch', () => {
  const noRealizations = () => [];

  it('matches everything when the query is empty', () => {
    const [node] = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1')]);
    expect(nodeMatchesSearch(node, '', noRealizations)).toBe(true);
  });

  it('matches by capability name, case-insensitively', () => {
    const [node] = buildCapabilityTree([cap('l1-a', 'Invoice Processing', 'L1')]);
    expect(nodeMatchesSearch(node, 'invoice', noRealizations)).toBe(true);
    expect(nodeMatchesSearch(node, 'nomatch', noRealizations)).toBe(false);
  });

  it('matches when a descendant capability name matches', () => {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1'), cap('l2-a1', 'Booking Management', 'L2', 'l1-a')]);
    expect(nodeMatchesSearch(tree[0], 'booking', noRealizations)).toBe(true);
  });

  it('matches by realising application name on the node itself', () => {
    const [node] = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1')]);
    const getRealizations = (id: ReturnType<typeof toCapabilityId>) =>
      id === node.capability.id ? [buildCapabilityRealization({ componentName: 'Phoenix Invoice Job' })] : [];

    expect(nodeMatchesSearch(node, 'phoenix', getRealizations)).toBe(true);
    expect(nodeMatchesSearch(node, 'seabook', getRealizations)).toBe(false);
  });

  it('matches by a descendant realising application name', () => {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1'), cap('l2-a1', 'Bravo', 'L2', 'l1-a')]);
    const l2Id = toCapabilityId('l2-a1');
    const getRealizations = (id: ReturnType<typeof toCapabilityId>) =>
      id === l2Id ? [buildCapabilityRealization({ componentName: 'Seabook' })] : [];

    expect(nodeMatchesSearch(tree[0], 'seabook', getRealizations)).toBe(true);
  });
});
