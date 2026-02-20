import { describe, it, expect, beforeEach } from 'vitest';
import {
  buildComponent,
  buildCapability,
  buildAcquiredEntity,
  buildVendor,
  buildInternalTeam,
  buildOriginRelationship,
  resetIdCounter,
} from '../../../test/helpers';
import {
  toComponentId,
  toCapabilityId,
  toAcquiredEntityId,
  toVendorId,
  toInternalTeamId,
  toOriginRelationshipId,
  toBusinessDomainId,
} from '../../../api/types';
import type { FilterableArtifacts } from './filterByCreator';
import {
  computeVisibleArtifactIds,
  filterByDomain,
  UNASSIGNED_DOMAIN,
} from './filterByDomain';
import type { DomainFilterData } from './filterByDomain';

function buildArtifacts(overrides: Partial<FilterableArtifacts> = {}): FilterableArtifacts {
  return {
    components: [],
    capabilities: [],
    acquiredEntities: [],
    vendors: [],
    internalTeams: [],
    ...overrides,
  };
}

function buildDomainFilterData(overrides: Partial<DomainFilterData> = {}): DomainFilterData {
  return {
    domainCapabilityIds: new Map(),
    allCapabilities: [],
    domainComponentIds: new Map(),
    originRelationships: [],
    allDomainIds: [],
    ...overrides,
  };
}

const DOMAIN_A = toBusinessDomainId('domain-a');
const DOMAIN_B = toBusinessDomainId('domain-b');

describe('filterByDomain', () => {
  beforeEach(() => {
    resetIdCounter();
  });

  describe('when selectedDomainIds is empty (filter inactive)', () => {
    it('should return all artifacts unchanged', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1') });
      const cap1 = buildCapability({ id: toCapabilityId('cap-1') });
      const ae1 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-1') });
      const vendor1 = buildVendor({ id: toVendorId('vendor-1') });
      const team1 = buildInternalTeam({ id: toInternalTeamId('team-1') });

      const artifacts = buildArtifacts({
        components: [comp1],
        capabilities: [cap1],
        acquiredEntities: [ae1],
        vendors: [vendor1],
        internalTeams: [team1],
      });

      const data = buildDomainFilterData();

      const result = filterByDomain(artifacts, [], data);

      expect(result.components).toEqual([comp1]);
      expect(result.capabilities).toEqual([cap1]);
      expect(result.acquiredEntities).toEqual([ae1]);
      expect(result.vendors).toEqual([vendor1]);
      expect(result.internalTeams).toEqual([team1]);
    });
  });

  describe('capability filtering by domain selection', () => {
    it.each([
      {
        scenario: 'shows capabilities directly assigned to the selected domain',
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-1', 'cap-2']]]),
        allDomainIds: [DOMAIN_A],
        selectedDomainIds: [DOMAIN_A],
        expectedCapIds: ['cap-1', 'cap-2'],
      },
      {
        scenario: 'returns union of domain and unassigned capabilities',
        domainCapabilityIds: new Map([
          [DOMAIN_A, ['cap-1']],
          [DOMAIN_B, ['cap-3']],
        ]),
        allDomainIds: [DOMAIN_A, DOMAIN_B],
        selectedDomainIds: [DOMAIN_A, UNASSIGNED_DOMAIN],
        expectedCapIds: ['cap-1', 'cap-2'],
      },
    ])(
      '$scenario',
      ({ domainCapabilityIds, allDomainIds, selectedDomainIds, expectedCapIds }) => {
        const cap1 = buildCapability({ id: toCapabilityId('cap-1') });
        const cap2 = buildCapability({ id: toCapabilityId('cap-2') });
        const cap3 = buildCapability({ id: toCapabilityId('cap-3') });
        const allCaps = [cap1, cap2, cap3];

        const artifacts = buildArtifacts({ capabilities: allCaps });
        const data = buildDomainFilterData({
          domainCapabilityIds,
          allCapabilities: allCaps,
          allDomainIds,
        });

        const result = filterByDomain(artifacts, selectedDomainIds, data);

        const expectedCaps = allCaps.filter((c) => expectedCapIds.includes(c.id));
        expect(result.capabilities).toEqual(expectedCaps);
      },
    );
  });

  describe('descendant capability expansion', () => {
    it('should include descendants transitively and exclude unrelated capabilities', () => {
      const capRoot = buildCapability({ id: toCapabilityId('cap-root'), name: 'Root' });
      const capChild = buildCapability({
        id: toCapabilityId('cap-child'),
        name: 'Child',
        parentId: toCapabilityId('cap-root'),
        level: 'L2',
      });
      const capGrandchild = buildCapability({
        id: toCapabilityId('cap-grandchild'),
        name: 'Grandchild',
        parentId: toCapabilityId('cap-child'),
        level: 'L3',
      });
      const capUnrelated = buildCapability({ id: toCapabilityId('cap-other'), name: 'Unrelated' });

      const artifacts = buildArtifacts({
        capabilities: [capRoot, capChild, capGrandchild, capUnrelated],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-root']]]),
        allCapabilities: [capRoot, capChild, capGrandchild, capUnrelated],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.capabilities).toEqual([capRoot, capChild, capGrandchild]);
    });
  });

  describe('components from domain realizations', () => {
    it('should show components linked to the selected domain via domainComponentIds', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Payment Service' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'Order Service' });
      const comp3 = buildComponent({ id: toComponentId('comp-3'), name: 'Unrelated Service' });

      const artifacts = buildArtifacts({
        components: [comp1, comp2, comp3],
      });

      const data = buildDomainFilterData({
        domainComponentIds: new Map([[DOMAIN_A, ['comp-1', 'comp-2']]]),
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.components).toEqual([comp1, comp2]);
    });
  });

  describe('origin entities linked to visible components', () => {
    it.each([
      {
        entityType: 'acquired entities' as const,
        buildIncluded: () => buildAcquiredEntity({ id: toAcquiredEntityId('ae-1'), name: 'Alpha' }),
        buildExcluded: () => buildAcquiredEntity({ id: toAcquiredEntityId('ae-2'), name: 'Beta' }),
        artifactKey: 'acquiredEntities' as const,
        originEntityId: 'ae-1',
        relationshipType: 'AcquiredVia' as const,
      },
      {
        entityType: 'vendors' as const,
        buildIncluded: () => buildVendor({ id: toVendorId('vendor-1'), name: 'Alpha' }),
        buildExcluded: () => buildVendor({ id: toVendorId('vendor-2'), name: 'Beta' }),
        artifactKey: 'vendors' as const,
        originEntityId: 'vendor-1',
        relationshipType: 'PurchasedFrom' as const,
      },
      {
        entityType: 'internal teams' as const,
        buildIncluded: () => buildInternalTeam({ id: toInternalTeamId('team-1'), name: 'Alpha' }),
        buildExcluded: () => buildInternalTeam({ id: toInternalTeamId('team-2'), name: 'Beta' }),
        artifactKey: 'internalTeams' as const,
        originEntityId: 'team-1',
        relationshipType: 'BuiltBy' as const,
      },
    ])(
      'should show $entityType linked to visible components',
      ({ buildIncluded, buildExcluded, artifactKey, originEntityId, relationshipType }) => {
        const comp1 = buildComponent({ id: toComponentId('comp-1') });
        const included = buildIncluded();
        const excluded = buildExcluded();

        const artifacts = buildArtifacts({
          components: [comp1],
          [artifactKey]: [included, excluded],
        });

        const data = buildDomainFilterData({
          domainComponentIds: new Map([[DOMAIN_A, ['comp-1']]]),
          originRelationships: [
            buildOriginRelationship({
              id: toOriginRelationshipId('or-1'),
              componentId: toComponentId('comp-1'),
              relationshipType,
              originEntityId,
            }),
          ],
          allDomainIds: [DOMAIN_A],
        });

        const result = filterByDomain(artifacts, [DOMAIN_A], data);

        expect(result[artifactKey]).toEqual([included]);
      },
    );
  });

  describe('multiple domains (union)', () => {
    it('should return the union of artifacts from all selected domains', () => {
      const cap1 = buildCapability({ id: toCapabilityId('cap-1'), name: 'Cap in Domain A' });
      const cap2 = buildCapability({ id: toCapabilityId('cap-2'), name: 'Cap in Domain B' });
      const cap3 = buildCapability({ id: toCapabilityId('cap-3'), name: 'Cap in neither' });

      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Comp in Domain A' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'Comp in Domain B' });

      const artifacts = buildArtifacts({
        capabilities: [cap1, cap2, cap3],
        components: [comp1, comp2],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([
          [DOMAIN_A, ['cap-1']],
          [DOMAIN_B, ['cap-2']],
        ]),
        allCapabilities: [cap1, cap2, cap3],
        domainComponentIds: new Map([
          [DOMAIN_A, ['comp-1']],
          [DOMAIN_B, ['comp-2']],
        ]),
        allDomainIds: [DOMAIN_A, DOMAIN_B],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A, DOMAIN_B], data);

      expect(result.capabilities).toEqual([cap1, cap2]);
      expect(result.components).toEqual([comp1, comp2]);
    });
  });

  describe('unassigned domain filter', () => {
    it('should show capabilities not assigned to any domain directly or via ancestry', () => {
      const capAssigned = buildCapability({ id: toCapabilityId('cap-assigned'), name: 'Assigned' });
      const capChildOfAssigned = buildCapability({
        id: toCapabilityId('cap-child'),
        name: 'Child of Assigned',
        parentId: toCapabilityId('cap-assigned'),
        level: 'L2',
      });
      const capOrphan = buildCapability({ id: toCapabilityId('cap-orphan'), name: 'Orphan' });

      const artifacts = buildArtifacts({
        capabilities: [capAssigned, capChildOfAssigned, capOrphan],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-assigned']]]),
        allCapabilities: [capAssigned, capChildOfAssigned, capOrphan],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [UNASSIGNED_DOMAIN], data);

      expect(result.capabilities).toEqual([capOrphan]);
    });

    it('should show components not realizing any domain-assigned capability', () => {
      const compAssigned = buildComponent({ id: toComponentId('comp-assigned'), name: 'Assigned' });
      const compOrphan = buildComponent({ id: toComponentId('comp-orphan'), name: 'Orphan' });

      const artifacts = buildArtifacts({
        components: [compAssigned, compOrphan],
      });

      const data = buildDomainFilterData({
        domainComponentIds: new Map([[DOMAIN_A, ['comp-assigned']]]),
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [UNASSIGNED_DOMAIN], data);

      expect(result.components).toEqual([compOrphan]);
    });

    it('should show origin entities not linked to any domain-reachable component', () => {
      const compAssigned = buildComponent({ id: toComponentId('comp-assigned') });
      const compOrphan = buildComponent({ id: toComponentId('comp-orphan') });

      const aeLinked = buildAcquiredEntity({ id: toAcquiredEntityId('ae-linked'), name: 'Linked' });
      const aeOrphan = buildAcquiredEntity({ id: toAcquiredEntityId('ae-orphan'), name: 'Orphan' });

      const artifacts = buildArtifacts({
        components: [compAssigned, compOrphan],
        acquiredEntities: [aeLinked, aeOrphan],
      });

      const data = buildDomainFilterData({
        domainComponentIds: new Map([[DOMAIN_A, ['comp-assigned']]]),
        originRelationships: [
          buildOriginRelationship({
            id: toOriginRelationshipId('or-1'),
            componentId: toComponentId('comp-assigned'),
            originEntityId: 'ae-linked',
          }),
          buildOriginRelationship({
            id: toOriginRelationshipId('or-2'),
            componentId: toComponentId('comp-orphan'),
            originEntityId: 'ae-orphan',
          }),
        ],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [UNASSIGNED_DOMAIN], data);

      expect(result.acquiredEntities).toEqual([aeOrphan]);
    });
  });

  describe('full traversal: domain -> capabilities -> components -> origin entities', () => {
    it('should traverse from domain through capability descendants to components to origin entities', () => {
      const capRoot = buildCapability({ id: toCapabilityId('cap-root'), name: 'Root Capability' });
      const capChild = buildCapability({
        id: toCapabilityId('cap-child'),
        name: 'Child Capability',
        parentId: toCapabilityId('cap-root'),
        level: 'L2',
      });

      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Service A' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'Service B' });

      const ae1 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-1'), name: 'Acquired Corp' });
      const v1 = buildVendor({ id: toVendorId('vendor-1'), name: 'Vendor Inc' });
      const team1 = buildInternalTeam({ id: toInternalTeamId('team-1'), name: 'Platform Team' });

      const capUnrelated = buildCapability({ id: toCapabilityId('cap-unrelated'), name: 'Unrelated' });
      const compUnrelated = buildComponent({ id: toComponentId('comp-unrelated'), name: 'Unrelated Service' });
      const aeUnrelated = buildAcquiredEntity({ id: toAcquiredEntityId('ae-unrelated'), name: 'Unrelated AE' });

      const artifacts = buildArtifacts({
        capabilities: [capRoot, capChild, capUnrelated],
        components: [comp1, comp2, compUnrelated],
        acquiredEntities: [ae1, aeUnrelated],
        vendors: [v1],
        internalTeams: [team1],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-root']]]),
        allCapabilities: [capRoot, capChild, capUnrelated],
        domainComponentIds: new Map([[DOMAIN_A, ['comp-1', 'comp-2']]]),
        originRelationships: [
          buildOriginRelationship({
            id: toOriginRelationshipId('or-1'),
            componentId: toComponentId('comp-1'),
            originEntityId: 'ae-1',
          }),
          buildOriginRelationship({
            id: toOriginRelationshipId('or-2'),
            componentId: toComponentId('comp-2'),
            relationshipType: 'PurchasedFrom',
            originEntityId: 'vendor-1',
          }),
          buildOriginRelationship({
            id: toOriginRelationshipId('or-3'),
            componentId: toComponentId('comp-1'),
            relationshipType: 'BuiltBy',
            originEntityId: 'team-1',
          }),
          buildOriginRelationship({
            id: toOriginRelationshipId('or-4'),
            componentId: toComponentId('comp-unrelated'),
            originEntityId: 'ae-unrelated',
          }),
        ],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.capabilities).toEqual([capRoot, capChild]);
      expect(result.components).toEqual([comp1, comp2]);
      expect(result.acquiredEntities).toEqual([ae1]);
      expect(result.vendors).toEqual([v1]);
      expect(result.internalTeams).toEqual([team1]);
    });
  });
});

describe('computeVisibleArtifactIds', () => {
  beforeEach(() => {
    resetIdCounter();
  });

  it('should return IDs of all directly assigned capabilities for a domain', () => {
    const cap1 = buildCapability({ id: toCapabilityId('cap-1') });
    const cap2 = buildCapability({ id: toCapabilityId('cap-2') });

    const data = buildDomainFilterData({
      domainCapabilityIds: new Map([[DOMAIN_A, ['cap-1', 'cap-2']]]),
      allCapabilities: [cap1, cap2],
      allDomainIds: [DOMAIN_A],
    });

    const visible = computeVisibleArtifactIds([DOMAIN_A], data);

    expect(visible.has('cap-1')).toBe(true);
    expect(visible.has('cap-2')).toBe(true);
  });

  it('should include descendant capability IDs transitively', () => {
    const capRoot = buildCapability({ id: toCapabilityId('cap-root') });
    const capChild = buildCapability({
      id: toCapabilityId('cap-child'),
      parentId: toCapabilityId('cap-root'),
    });
    const capGrandchild = buildCapability({
      id: toCapabilityId('cap-grandchild'),
      parentId: toCapabilityId('cap-child'),
    });

    const data = buildDomainFilterData({
      domainCapabilityIds: new Map([[DOMAIN_A, ['cap-root']]]),
      allCapabilities: [capRoot, capChild, capGrandchild],
      allDomainIds: [DOMAIN_A],
    });

    const visible = computeVisibleArtifactIds([DOMAIN_A], data);

    expect(visible.has('cap-root')).toBe(true);
    expect(visible.has('cap-child')).toBe(true);
    expect(visible.has('cap-grandchild')).toBe(true);
  });

  it('should include component IDs from domainComponentIds', () => {
    const data = buildDomainFilterData({
      domainComponentIds: new Map([[DOMAIN_A, ['comp-1', 'comp-2']]]),
      allDomainIds: [DOMAIN_A],
    });

    const visible = computeVisibleArtifactIds([DOMAIN_A], data);

    expect(visible.has('comp-1')).toBe(true);
    expect(visible.has('comp-2')).toBe(true);
  });

  it('should include origin entity IDs linked to visible components', () => {
    const data = buildDomainFilterData({
      domainComponentIds: new Map([[DOMAIN_A, ['comp-1']]]),
      originRelationships: [
        buildOriginRelationship({
          componentId: toComponentId('comp-1'),
          originEntityId: 'ae-1',
        }),
        buildOriginRelationship({
          componentId: toComponentId('comp-unrelated'),
          originEntityId: 'ae-2',
        }),
      ],
      allDomainIds: [DOMAIN_A],
    });

    const visible = computeVisibleArtifactIds([DOMAIN_A], data);

    expect(visible.has('ae-1')).toBe(true);
    expect(visible.has('ae-2')).toBe(false);
  });

  it('should compute unassigned as the complement of the full domain set', () => {
    const cap1 = buildCapability({ id: toCapabilityId('cap-1') });
    const cap2 = buildCapability({ id: toCapabilityId('cap-2') });
    const capOrphan = buildCapability({ id: toCapabilityId('cap-orphan') });

    const data = buildDomainFilterData({
      domainCapabilityIds: new Map([[DOMAIN_A, ['cap-1', 'cap-2']]]),
      allCapabilities: [cap1, cap2, capOrphan],
      domainComponentIds: new Map([[DOMAIN_A, ['comp-1']]]),
      originRelationships: [],
      allDomainIds: [DOMAIN_A],
    });

    const visible = computeVisibleArtifactIds([UNASSIGNED_DOMAIN], data);

    expect(visible.has('cap-1')).toBe(false);
    expect(visible.has('cap-2')).toBe(false);
    expect(visible.has('cap-orphan')).toBe(true);
    expect(visible.has('comp-1')).toBe(false);
  });
});
