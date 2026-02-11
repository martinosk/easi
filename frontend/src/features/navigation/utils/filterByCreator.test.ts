import { describe, it, expect, beforeEach } from 'vitest';
import { buildComponent, buildCapability, resetIdCounter } from '../../../test/helpers';
import {
  toComponentId,
  toCapabilityId,
  toAcquiredEntityId,
  toVendorId,
  toInternalTeamId,
} from '../../../api/types';
import type {
  AcquiredEntity,
  Vendor,
  InternalTeam,
  HATEOASLinks,
} from '../../../api/types';
import { filterByCreator } from './filterByCreator';
import type { FilterableArtifacts } from './filterByCreator';

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

describe('filterByCreator', () => {
  beforeEach(() => {
    resetIdCounter();
  });

  const USER_ALICE = 'user-alice';
  const USER_BOB = 'user-bob';
  const USER_CAROL = 'user-carol';

  function buildCreatorMap(entries: Array<[string, string]>): Map<string, string> {
    return new Map(entries);
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

  describe('when selectedCreatorIds is empty (filter inactive)', () => {
    it('should return all artifacts unchanged', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1') });
      const comp2 = buildComponent({ id: toComponentId('comp-2') });
      const cap1 = buildCapability({ id: toCapabilityId('cap-1') });
      const ae1 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-1') });
      const vendor1 = buildVendor({ id: toVendorId('vendor-1') });
      const team1 = buildInternalTeam({ id: toInternalTeamId('team-1') });

      const artifacts = buildArtifacts({
        components: [comp1, comp2],
        capabilities: [cap1],
        acquiredEntities: [ae1],
        vendors: [vendor1],
        internalTeams: [team1],
      });

      const creatorMap = buildCreatorMap([
        ['comp-1', USER_ALICE],
        ['comp-2', USER_BOB],
        ['cap-1', USER_ALICE],
        ['ae-1', USER_BOB],
        ['vendor-1', USER_CAROL],
        ['team-1', USER_ALICE],
      ]);

      const result = filterByCreator(artifacts, [], creatorMap);

      expect(result.components).toEqual([comp1, comp2]);
      expect(result.capabilities).toEqual([cap1]);
      expect(result.acquiredEntities).toEqual([ae1]);
      expect(result.vendors).toEqual([vendor1]);
      expect(result.internalTeams).toEqual([team1]);
    });
  });

  describe('filtering components by creator', () => {
    it('should return only components created by the selected creator', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Payment Service' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'Order Service' });
      const comp3 = buildComponent({ id: toComponentId('comp-3'), name: 'Shipping Service' });

      const artifacts = buildArtifacts({
        components: [comp1, comp2, comp3],
      });

      const creatorMap = buildCreatorMap([
        ['comp-1', USER_ALICE],
        ['comp-2', USER_BOB],
        ['comp-3', USER_ALICE],
      ]);

      const result = filterByCreator(artifacts, [USER_ALICE], creatorMap);

      expect(result.components).toEqual([comp1, comp3]);
    });
  });

  describe('filtering capabilities by creator', () => {
    it('should return only capabilities created by the selected creator (flat, no hierarchy)', () => {
      const cap1 = buildCapability({ id: toCapabilityId('cap-1'), name: 'Customer Mgmt' });
      const cap2 = buildCapability({
        id: toCapabilityId('cap-2'),
        name: 'Order Processing',
        parentId: toCapabilityId('cap-1'),
        level: 'L2',
      });
      const cap3 = buildCapability({ id: toCapabilityId('cap-3'), name: 'Logistics' });

      const artifacts = buildArtifacts({
        capabilities: [cap1, cap2, cap3],
      });

      const creatorMap = buildCreatorMap([
        ['cap-1', USER_BOB],
        ['cap-2', USER_ALICE],
        ['cap-3', USER_BOB],
      ]);

      const result = filterByCreator(artifacts, [USER_ALICE], creatorMap);

      expect(result.capabilities).toEqual([cap2]);
    });
  });

  describe('filtering origin entity types by creator', () => {
    it('should filter acquired entities by creator', () => {
      const ae1 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-1'), name: 'Acquired A' });
      const ae2 = buildAcquiredEntity({ id: toAcquiredEntityId('ae-2'), name: 'Acquired B' });

      const artifacts = buildArtifacts({
        acquiredEntities: [ae1, ae2],
      });

      const creatorMap = buildCreatorMap([
        ['ae-1', USER_ALICE],
        ['ae-2', USER_BOB],
      ]);

      const result = filterByCreator(artifacts, [USER_ALICE], creatorMap);

      expect(result.acquiredEntities).toEqual([ae1]);
    });

    it('should filter vendors by creator', () => {
      const v1 = buildVendor({ id: toVendorId('vendor-1'), name: 'Vendor Alpha' });
      const v2 = buildVendor({ id: toVendorId('vendor-2'), name: 'Vendor Beta' });

      const artifacts = buildArtifacts({
        vendors: [v1, v2],
      });

      const creatorMap = buildCreatorMap([
        ['vendor-1', USER_BOB],
        ['vendor-2', USER_BOB],
      ]);

      const result = filterByCreator(artifacts, [USER_ALICE], creatorMap);

      expect(result.vendors).toEqual([]);
    });

    it('should filter internal teams by creator', () => {
      const team1 = buildInternalTeam({ id: toInternalTeamId('team-1'), name: 'Platform' });
      const team2 = buildInternalTeam({ id: toInternalTeamId('team-2'), name: 'Mobile' });

      const artifacts = buildArtifacts({
        internalTeams: [team1, team2],
      });

      const creatorMap = buildCreatorMap([
        ['team-1', USER_CAROL],
        ['team-2', USER_ALICE],
      ]);

      const result = filterByCreator(artifacts, [USER_CAROL], creatorMap);

      expect(result.internalTeams).toEqual([team1]);
    });
  });

  describe('artifacts not in creatorMap', () => {
    it('should filter out artifacts that have no entry in creatorMap when filter is active', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Known Creator' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'Unknown Creator' });

      const artifacts = buildArtifacts({
        components: [comp1, comp2],
      });

      const creatorMap = buildCreatorMap([
        ['comp-1', USER_ALICE],
      ]);

      const result = filterByCreator(artifacts, [USER_ALICE], creatorMap);

      expect(result.components).toEqual([comp1]);
      expect(result.components).not.toContainEqual(
        expect.objectContaining({ name: 'Unknown Creator' })
      );
    });
  });

  describe('filtering by multiple creators (union)', () => {
    it('should return artifacts created by any of the selected creators', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'By Alice' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'By Bob' });
      const comp3 = buildComponent({ id: toComponentId('comp-3'), name: 'By Carol' });

      const cap1 = buildCapability({ id: toCapabilityId('cap-1'), name: 'Cap by Alice' });
      const cap2 = buildCapability({ id: toCapabilityId('cap-2'), name: 'Cap by Carol' });

      const artifacts = buildArtifacts({
        components: [comp1, comp2, comp3],
        capabilities: [cap1, cap2],
      });

      const creatorMap = buildCreatorMap([
        ['comp-1', USER_ALICE],
        ['comp-2', USER_BOB],
        ['comp-3', USER_CAROL],
        ['cap-1', USER_ALICE],
        ['cap-2', USER_CAROL],
      ]);

      const result = filterByCreator(artifacts, [USER_ALICE, USER_BOB], creatorMap);

      expect(result.components).toEqual([comp1, comp2]);
      expect(result.capabilities).toEqual([cap1]);
    });
  });
});
