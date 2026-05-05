import { beforeEach, describe, expect, it } from 'vitest';
import {
  toAcquiredEntityId,
  toCapabilityId,
  toComponentId,
  toInternalTeamId,
  toVendorId,
  toViewId,
} from '../../../api/types';
import {
  buildAcquiredEntity,
  buildCapability,
  buildComponent,
  buildInternalTeam,
  buildVendor,
  buildView,
  resetIdCounter,
} from '../../../test/helpers';
import type { FilterableArtifacts } from './filterByCreator';
import { filterByCreator, filterEntitiesByCreator } from './filterByCreator';

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
    type OriginCase = {
      label: string;
      key: keyof FilterableArtifacts;
      keptId: string;
      droppedId: string;
      keptCreator: string;
      droppedCreator: string;
      selected: string;
      build: (id: string) => { id: string };
    };

    const cases: OriginCase[] = [
      {
        label: 'acquired entities',
        key: 'acquiredEntities',
        keptId: 'ae-1',
        droppedId: 'ae-2',
        keptCreator: USER_ALICE,
        droppedCreator: USER_BOB,
        selected: USER_ALICE,
        build: (id) => buildAcquiredEntity({ id: toAcquiredEntityId(id) }),
      },
      {
        label: 'vendors',
        key: 'vendors',
        keptId: 'vendor-1',
        droppedId: 'vendor-2',
        keptCreator: USER_BOB,
        droppedCreator: USER_BOB,
        selected: USER_ALICE,
        build: (id) => buildVendor({ id: toVendorId(id) }),
      },
      {
        label: 'internal teams',
        key: 'internalTeams',
        keptId: 'team-1',
        droppedId: 'team-2',
        keptCreator: USER_CAROL,
        droppedCreator: USER_ALICE,
        selected: USER_CAROL,
        build: (id) => buildInternalTeam({ id: toInternalTeamId(id) }),
      },
    ];

    it.each(cases)('should filter $label by creator', (c) => {
      const kept = c.build(c.keptId);
      const dropped = c.build(c.droppedId);
      const artifacts = buildArtifacts({ [c.key]: [kept, dropped] });
      const creatorMap = buildCreatorMap([
        [c.keptId, c.keptCreator],
        [c.droppedId, c.droppedCreator],
      ]);

      const result = filterByCreator(artifacts, [c.selected], creatorMap);

      const expected = c.selected === c.keptCreator ? [kept] : [];
      expect(result[c.key]).toEqual(expected);
    });
  });

  describe('artifacts not in creatorMap', () => {
    it('should filter out artifacts that have no entry in creatorMap when filter is active', () => {
      const comp1 = buildComponent({ id: toComponentId('comp-1'), name: 'Known Creator' });
      const comp2 = buildComponent({ id: toComponentId('comp-2'), name: 'Unknown Creator' });

      const artifacts = buildArtifacts({
        components: [comp1, comp2],
      });

      const creatorMap = buildCreatorMap([['comp-1', USER_ALICE]]);

      const result = filterByCreator(artifacts, [USER_ALICE], creatorMap);

      expect(result.components).toEqual([comp1]);
      expect(result.components).not.toContainEqual(expect.objectContaining({ name: 'Unknown Creator' }));
    });
  });

  describe('filterEntitiesByCreator (used for views)', () => {
    it('should return all items unchanged when no creators are selected', () => {
      const v1 = buildView({ id: toViewId('view-1') });
      const v2 = buildView({ id: toViewId('view-2') });
      const creatorMap = buildCreatorMap([
        ['view-1', USER_ALICE],
        ['view-2', USER_BOB],
      ]);

      const result = filterEntitiesByCreator([v1, v2], [], creatorMap);

      expect(result).toEqual([v1, v2]);
    });

    it('should return only views created by the selected creator', () => {
      const v1 = buildView({ id: toViewId('view-1'), name: 'Alice view' });
      const v2 = buildView({ id: toViewId('view-2'), name: 'Bob view' });
      const v3 = buildView({ id: toViewId('view-3'), name: 'Alice second view' });
      const creatorMap = buildCreatorMap([
        ['view-1', USER_ALICE],
        ['view-2', USER_BOB],
        ['view-3', USER_ALICE],
      ]);

      const result = filterEntitiesByCreator([v1, v2, v3], [USER_ALICE], creatorMap);

      expect(result).toEqual([v1, v3]);
    });

    it('should drop views with no entry in creatorMap when filter is active', () => {
      const v1 = buildView({ id: toViewId('view-1'), name: 'Known' });
      const v2 = buildView({ id: toViewId('view-2'), name: 'Unknown' });
      const creatorMap = buildCreatorMap([['view-1', USER_ALICE]]);

      const result = filterEntitiesByCreator([v1, v2], [USER_ALICE], creatorMap);

      expect(result).toEqual([v1]);
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
