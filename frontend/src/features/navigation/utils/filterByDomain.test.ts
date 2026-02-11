import { describe, it, expect, beforeEach } from 'vitest';
import { buildComponent, buildCapability, resetIdCounter } from '../../../test/helpers';
import {
  toComponentId,
  toCapabilityId,
  toAcquiredEntityId,
  toVendorId,
  toInternalTeamId,
  toBusinessDomainId,
  toOriginRelationshipId,
} from '../../../api/types';
import type {
  AcquiredEntity,
  Vendor,
  InternalTeam,
  OriginRelationship,
  HATEOASLinks,
} from '../../../api/types';
import type { FilterableArtifacts } from './filterByCreator';
import {
  computeVisibleArtifactIds,
  filterByDomain,
  UNASSIGNED_DOMAIN,
} from './filterByDomain';
import type { DomainFilterData } from './filterByDomain';

function buildLinks(href: string): HATEOASLinks {
  return {
    self: { href, method: 'GET' },
    edit: { href, method: 'PUT' },
    delete: { href, method: 'DELETE' },
  };
}

function buildAcquiredEntity(overrides: Partial<AcquiredEntity> = {}): AcquiredEntity {
  return {
    id: toAcquiredEntityId('ae-1'),
    name: 'Acquired Corp',
    integrationStatus: 'NOT_STARTED',
    componentCount: 0,
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks('/api/v1/acquired-entities/ae-1'),
    ...overrides,
  };
}

function buildVendor(overrides: Partial<Vendor> = {}): Vendor {
  return {
    id: toVendorId('vendor-1'),
    name: 'Vendor Inc',
    componentCount: 0,
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks('/api/v1/vendors/vendor-1'),
    ...overrides,
  };
}

function buildInternalTeam(overrides: Partial<InternalTeam> = {}): InternalTeam {
  return {
    id: toInternalTeamId('team-1'),
    name: 'Platform Team',
    componentCount: 0,
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks('/api/v1/internal-teams/team-1'),
    ...overrides,
  };
}

function buildOriginRelationship(overrides: Partial<OriginRelationship> = {}): OriginRelationship {
  return {
    id: toOriginRelationshipId('or-1'),
    componentId: toComponentId('comp-1'),
    componentName: 'Component comp-1',
    relationshipType: 'AcquiredVia',
    originEntityId: 'ae-1',
    originEntityName: 'Acquired Corp',
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks('/api/v1/origin-relationships/or-1'),
    ...overrides,
  };
}

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

  describe('directly assigned capabilities', () => {
    it('should show capabilities directly assigned to the selected domain', () => {
      const cap1 = buildCapability({ id: toCapabilityId('cap-1'), name: 'Customer Mgmt' });
      const cap2 = buildCapability({ id: toCapabilityId('cap-2'), name: 'Order Processing' });
      const cap3 = buildCapability({ id: toCapabilityId('cap-3'), name: 'Logistics' });

      const artifacts = buildArtifacts({
        capabilities: [cap1, cap2, cap3],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-1', 'cap-2']]]),
        allCapabilities: [cap1, cap2, cap3],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.capabilities).toEqual([cap1, cap2]);
    });
  });

  describe('descendant capability expansion', () => {
    it('should include child capabilities of a directly assigned capability', () => {
      const capParent = buildCapability({ id: toCapabilityId('cap-parent'), name: 'Parent' });
      const capChild = buildCapability({
        id: toCapabilityId('cap-child'),
        name: 'Child',
        parentId: toCapabilityId('cap-parent'),
        level: 'L2',
      });
      const capUnrelated = buildCapability({ id: toCapabilityId('cap-other'), name: 'Unrelated' });

      const artifacts = buildArtifacts({
        capabilities: [capParent, capChild, capUnrelated],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-parent']]]),
        allCapabilities: [capParent, capChild, capUnrelated],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.capabilities).toEqual([capParent, capChild]);
    });

    it('should include grandchild capabilities transitively', () => {
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

      const artifacts = buildArtifacts({
        capabilities: [capRoot, capChild, capGrandchild],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([[DOMAIN_A, ['cap-root']]]),
        allCapabilities: [capRoot, capChild, capGrandchild],
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
    it('should show origin entities whose component is visible in the selected domain', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Payment Service' });
      const ae1 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-1'), name: 'Acquired Alpha' });
      const ae2 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-2'), name: 'Acquired Beta' });

      const artifacts = buildArtifacts({
        components: [comp1],
        acquiredEntities: [ae1, ae2],
      });

      const data = buildDomainFilterData({
        domainComponentIds: new Map([[DOMAIN_A, ['comp-1']]]),
        originRelationships: [
          buildOriginRelationship({
            id: toOriginRelationshipId('or-1'),
            componentId: toComponentId('comp-1'),
            originEntityId: 'ae-1',
          }),
        ],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.acquiredEntities).toEqual([ae1]);
    });

    it('should show vendors linked to visible components', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1') });
      const v1 = buildVendor({ id: toVendorId('vendor-1'), name: 'Vendor Alpha' });
      const v2 = buildVendor({ id: toVendorId('vendor-2'), name: 'Vendor Beta' });

      const artifacts = buildArtifacts({
        components: [comp1],
        vendors: [v1, v2],
      });

      const data = buildDomainFilterData({
        domainComponentIds: new Map([[DOMAIN_A, ['comp-1']]]),
        originRelationships: [
          buildOriginRelationship({
            id: toOriginRelationshipId('or-1'),
            componentId: toComponentId('comp-1'),
            relationshipType: 'PurchasedFrom',
            originEntityId: 'vendor-1',
          }),
        ],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.vendors).toEqual([v1]);
    });

    it('should show internal teams linked to visible components', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1') });
      const team1 = buildInternalTeam({ id: toInternalTeamId('team-1'), name: 'Platform' });
      const team2 = buildInternalTeam({ id: toInternalTeamId('team-2'), name: 'Mobile' });

      const artifacts = buildArtifacts({
        components: [comp1],
        internalTeams: [team1, team2],
      });

      const data = buildDomainFilterData({
        domainComponentIds: new Map([[DOMAIN_A, ['comp-1']]]),
        originRelationships: [
          buildOriginRelationship({
            id: toOriginRelationshipId('or-1'),
            componentId: toComponentId('comp-1'),
            relationshipType: 'BuiltBy',
            originEntityId: 'team-1',
          }),
        ],
        allDomainIds: [DOMAIN_A],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A], data);

      expect(result.internalTeams).toEqual([team1]);
    });
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

  describe('combined domain + unassigned', () => {
    it('should return the union of domain artifacts and unassigned artifacts', () => {
      const capInA = buildCapability({ id: toCapabilityId('cap-in-a'), name: 'In Domain A' });
      const capOrphan = buildCapability({ id: toCapabilityId('cap-orphan'), name: 'Orphan' });
      const capInB = buildCapability({ id: toCapabilityId('cap-in-b'), name: 'In Domain B' });

      const artifacts = buildArtifacts({
        capabilities: [capInA, capOrphan, capInB],
      });

      const data = buildDomainFilterData({
        domainCapabilityIds: new Map([
          [DOMAIN_A, ['cap-in-a']],
          [DOMAIN_B, ['cap-in-b']],
        ]),
        allCapabilities: [capInA, capOrphan, capInB],
        allDomainIds: [DOMAIN_A, DOMAIN_B],
      });

      const result = filterByDomain(artifacts, [DOMAIN_A, UNASSIGNED_DOMAIN], data);

      expect(result.capabilities).toEqual([capInA, capOrphan]);
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
