import { describe, expect, it } from 'vitest';
import { buildCapabilityJourney } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { buildJourneyIndex, capabilityHasChange, l1HasChange, summaryCounts } from './journeyIndex';

function moveJourney(capabilityId: string, targetParentId: string | null, targetDomainId: string) {
  return buildCapabilityJourney({
    id: `mv-${capabilityId}`,
    capabilityId,
    kind: 'move',
    status: 'planned',
    move: {
      targetDomainId,
      targetDomainName: 'Group functions',
      targetDomainStale: false,
      targetParentId,
      targetParentName: targetParentId ? 'Accounts payable' : '',
      targetParentStale: false,
      resultingName: 'Freight invoicing',
    },
  });
}

describe('buildJourneyIndex', () => {
  it('exposes the board journey per capability, ignoring abandoned', () => {
    const index = buildJourneyIndex({
      journeys: [
        buildCapabilityJourney({ id: 'j1', capabilityId: 'a', status: 'in-flight' }),
        buildCapabilityJourney({ id: 'j2', capabilityId: 'b', status: 'abandoned' }),
      ],
      capabilityDomainNames: new Map(),
    });
    expect(index.getJourney('a')?.id).toBe('j1');
    expect(index.getJourney('b')).toBeUndefined();
  });

  it('indexes arriving moves under their target parent', () => {
    const index = buildJourneyIndex({
      journeys: [moveJourney('inv', 'ap', 'gf')],
      capabilityDomainNames: new Map([['inv', 'Ferry freight']]),
    });
    expect(index.getArrivingMovesForParent('ap').map((j) => j.capabilityId)).toEqual(['inv']);
    expect(index.getArrivingMovesForDomain('gf')).toEqual([]);
    expect(index.sourceDomainName('inv')).toBe('Ferry freight');
  });

  it('indexes arriving moves with no target parent at the domain top level', () => {
    const index = buildJourneyIndex({
      journeys: [moveJourney('inv', null, 'gf')],
      capabilityDomainNames: new Map(),
    });
    expect(index.getArrivingMovesForDomain('gf').map((j) => j.capabilityId)).toEqual(['inv']);
    expect(index.getArrivingMovesForParent('ap')).toEqual([]);
  });
});

describe('summaryCounts', () => {
  it('counts settled (done or steady), in flight, and not started over the rendered cards', () => {
    const [transport] = buildCapabilityTree(
      [
        cap('te', 'Transport execution', 'L1'),
        cap('tp', 'Transport planning', 'L2', 'te'),
        cap('tt', 'Track & trace', 'L2', 'te'),
      ],
      { orphanRoots: 'any-level' },
    );
    const [invoicing] = buildCapabilityTree([cap('inv', 'Invoicing', 'L1')], { orphanRoots: 'any-level' });

    const index = buildJourneyIndex({
      journeys: [
        buildCapabilityJourney({ capabilityId: 'tp', kind: 'consolidation', status: 'planned' }),
        buildCapabilityJourney({ capabilityId: 'tt', kind: 'consolidation', status: 'done' }),
        moveJourney('inv', 'ap', 'gf'),
      ],
      capabilityDomainNames: new Map(),
    });

    expect(summaryCounts([transport, invoicing], index.getJourney)).toEqual({
      settled: 1,
      inFlight: 0,
      notStarted: 2,
    });
  });
});

describe('change helpers', () => {
  it('capabilityHasChange is true for a journey or an arriving move', () => {
    const index = buildJourneyIndex({
      journeys: [buildCapabilityJourney({ capabilityId: 'a', status: 'in-flight' }), moveJourney('inv', 'ap', 'gf')],
      capabilityDomainNames: new Map(),
    });
    expect(capabilityHasChange('a', index)).toBe(true);
    expect(capabilityHasChange('ap', index)).toBe(true);
    expect(capabilityHasChange('unrelated', index)).toBe(false);
  });

  it('l1HasChange is true when any descendant has a change', () => {
    const [group] = buildCapabilityTree([cap('fin', 'Finance', 'L1'), cap('ap', 'Accounts payable', 'L2', 'fin')], {
      orphanRoots: 'any-level',
    });
    const withArriving = buildJourneyIndex({
      journeys: [moveJourney('inv', 'ap', 'gf')],
      capabilityDomainNames: new Map(),
    });
    const withoutChange = buildJourneyIndex({ journeys: [], capabilityDomainNames: new Map() });

    expect(l1HasChange(group, withArriving)).toBe(true);
    expect(l1HasChange(group, withoutChange)).toBe(false);
  });
});
