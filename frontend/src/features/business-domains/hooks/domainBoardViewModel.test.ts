import { describe, expect, it } from 'vitest';
import type { CapabilityRealization } from '../../../api/types';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildBusinessDomain, buildCapabilityRealization } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import type { BuildDomainBoardViewModelParams } from './domainBoardViewModel';
import { buildDomainBoardViewModel, flattenViewModelCapabilities } from './domainBoardViewModel';

const domain = buildBusinessDomain({ name: 'Freight' });

function buildViewModel(overrides: Partial<BuildDomainBoardViewModelParams> = {}) {
  return buildDomainBoardViewModel({
    domain,
    assignedCapabilities: [],
    tree: [],
    realizationGroups: [],
    isLoading: false,
    ...overrides,
  });
}

function buildAlphaHierarchyViewModel() {
  const tree = buildCapabilityTree([
    cap('l1-a', 'Alpha', 'L1'),
    cap('l2-a1', 'Alpha One', 'L2', 'l1-a'),
    cap('l3-a1a', 'Alpha One A', 'L3', 'l2-a1'),
  ]);
  return buildViewModel({ assignedCapabilities: [cap('l1-a', 'Alpha', 'L1')], tree });
}

describe('buildDomainBoardViewModel', () => {
  it('includes only the L1 groups assigned to the domain', () => {
    const tree = buildCapabilityTree([
      cap('l1-a', 'Alpha', 'L1'),
      cap('l1-b', 'Bravo', 'L1'),
      cap('l2-a1', 'Alpha One', 'L2', 'l1-a'),
    ]);

    const vm = buildViewModel({ assignedCapabilities: [cap('l1-a', 'Alpha', 'L1')], tree });

    expect(vm.l1Groups).toHaveLength(1);
    expect(vm.l1Groups[0].node.capability.name).toBe('Alpha');
    expect(vm.l1Groups[0].node.children.map((c) => c.capability.name)).toEqual(['Alpha One']);
  });

  it('counts the L1 plus all its descendants as the total capability count', () => {
    const vm = buildAlphaHierarchyViewModel();

    expect(vm.totalCapabilityCount).toBe(3);
  });

  it('computes a distinct app count per L1 group, deduplicating repeated components across descendants', () => {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1'), cap('l2-a1', 'Alpha One', 'L2', 'l1-a')]);
    const groups = [
      {
        capabilityId: toCapabilityId('l1-a'),
        capabilityName: 'Alpha',
        level: 'L1' as const,
        realizations: [
          buildCapabilityRealization({ capabilityId: toCapabilityId('l1-a'), componentId: toComponentId('comp-1') }),
        ],
      },
      {
        capabilityId: toCapabilityId('l2-a1'),
        capabilityName: 'Alpha One',
        level: 'L2' as const,
        realizations: [
          buildCapabilityRealization({ capabilityId: toCapabilityId('l2-a1'), componentId: toComponentId('comp-1') }),
          buildCapabilityRealization({ capabilityId: toCapabilityId('l2-a1'), componentId: toComponentId('comp-2') }),
        ],
      },
    ];

    const vm = buildViewModel({ assignedCapabilities: [cap('l1-a', 'Alpha', 'L1')], tree, realizationGroups: groups });

    expect(vm.l1Groups[0].distinctAppCount).toBe(2);
    expect(vm.totalAppCount).toBe(2);
  });

  it('exposes a lookup for realizations by capability id, defaulting to empty', () => {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1')]);
    const groups = [
      {
        capabilityId: toCapabilityId('l1-a'),
        capabilityName: 'Alpha',
        level: 'L1' as const,
        realizations: [buildCapabilityRealization({ capabilityId: toCapabilityId('l1-a') })],
      },
    ];

    const vm = buildViewModel({ assignedCapabilities: [cap('l1-a', 'Alpha', 'L1')], tree, realizationGroups: groups });

    expect(vm.getRealizationsForCapability(toCapabilityId('l1-a'))).toHaveLength(1);
    expect(vm.getRealizationsForCapability(toCapabilityId('missing'))).toEqual([]);
  });

  it('returns no groups when nothing is assigned', () => {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1')]);

    const vm = buildViewModel({ tree });

    expect(vm.l1Groups).toEqual([]);
    expect(vm.totalCapabilityCount).toBe(0);
    expect(vm.totalAppCount).toBe(0);
  });

  function buildAssessedRealizationViewModel(realizationOverrides: Partial<CapabilityRealization> = {}) {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1')]);
    const groups = [
      {
        capabilityId: toCapabilityId('l1-a'),
        capabilityName: 'Alpha',
        level: 'L1' as const,
        realizations: [
          buildCapabilityRealization({
            capabilityId: toCapabilityId('l1-a'),
            componentId: toComponentId('comp-1'),
            ...realizationOverrides,
          }),
        ],
      },
    ];

    return buildViewModel({
      assignedCapabilities: [cap('l1-a', 'Alpha', 'L1')],
      tree,
      realizationGroups: groups,
      assessments: [
        {
          id: 'ta-1',
          capabilityId: 'l1-a',
          capabilityName: 'Alpha',
          componentId: 'comp-1',
          componentName: 'Component 1',
          grade: 'Migrate',
          rationale: '',
          assessedBy: 'user-1',
          assessedAt: '2026-01-01T00:00:00Z',
          stale: false,
          suggestion: null,
          _links: {},
        },
      ],
    });
  }

  it('attaches the current assessed grade to a Direct realization matching the assessment pair', () => {
    const vm = buildAssessedRealizationViewModel({ origin: 'Direct' });

    expect(vm.getRealizationsForCapability(toCapabilityId('l1-a'))[0].timeGrade).toBe('Migrate');
  });

  it('does not attach a grade to an Inherited realization even when the app is assessed on its source pair', () => {
    const vm = buildAssessedRealizationViewModel({ origin: 'Inherited', sourceCapabilityId: toCapabilityId('l1-source') });

    expect(vm.getRealizationsForCapability(toCapabilityId('l1-a'))[0].timeGrade).toBeUndefined();
  });

  function buildRoleEnrichedViewModel(realizationOverrides: Partial<CapabilityRealization> = {}) {
    const tree = buildCapabilityTree([cap('l1-a', 'Alpha', 'L1')]);
    const groups = [
      {
        capabilityId: toCapabilityId('l1-a'),
        capabilityName: 'Alpha',
        level: 'L1' as const,
        realizations: [
          buildCapabilityRealization({
            capabilityId: toCapabilityId('l1-a'),
            componentId: toComponentId('comp-1'),
            ...realizationOverrides,
          }),
        ],
      },
    ];

    return buildViewModel({
      assignedCapabilities: [cap('l1-a', 'Alpha', 'L1')],
      tree,
      realizationGroups: groups,
      roles: [
        {
          capabilityId: 'l1-a',
          capabilityName: 'Alpha',
          componentId: 'comp-1',
          componentName: 'Component 1',
          role: 'standard',
          assignedBy: 'user-1',
          assignedAt: '2026-01-01T00:00:00Z',
          _links: {},
        },
      ],
    });
  }

  it('attaches the current role to a Direct realization matching the role pair', () => {
    const vm = buildRoleEnrichedViewModel({ origin: 'Direct' });

    expect(vm.getRealizationsForCapability(toCapabilityId('l1-a'))[0].role).toBe('standard');
  });

  it('does not attach a role to an Inherited realization even when the app holds a role on its source pair', () => {
    const vm = buildRoleEnrichedViewModel({ origin: 'Inherited', sourceCapabilityId: toCapabilityId('l1-source') });

    expect(vm.getRealizationsForCapability(toCapabilityId('l1-a'))[0].role).toBeUndefined();
  });
});

describe('flattenViewModelCapabilities', () => {
  it('flattens every L1 group and its descendants into a flat capability list', () => {
    const vm = buildAlphaHierarchyViewModel();

    expect(flattenViewModelCapabilities(vm).map((c) => c.name)).toEqual(['Alpha', 'Alpha One', 'Alpha One A']);
  });
});
