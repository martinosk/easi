import { describe, expect, it } from 'vitest';
import type {
  Capability,
  CapabilityId,
  CapabilityRealization,
  ComponentId,
  HATEOASLinks,
  OriginRelationship,
  OriginRelationshipId,
  RealizationId,
  Relation,
  RelationId,
} from '../../../api/types';
import { type DynamicGraphData, type EntityRef, getNeighbors, getUnexpandedByEdgeType } from './dynamicMode';

const links: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

function makeRelation(source: string, target: string): Relation {
  return {
    id: `rel-${source}-${target}` as RelationId,
    sourceComponentId: source as ComponentId,
    targetComponentId: target as ComponentId,
    relationType: 'Triggers',
    createdAt: '2024-01-01',
    _links: links,
  };
}

function makeRealization(componentId: string, capabilityId: string): CapabilityRealization {
  return {
    id: `rz-${componentId}-${capabilityId}` as RealizationId,
    capabilityId: capabilityId as CapabilityId,
    componentId: componentId as ComponentId,
    realizationLevel: 'Full',
    origin: 'Direct',
    linkedAt: '2024-01-01',
    _links: links,
  };
}

function makeCapability(id: string, parentId?: string): Capability {
  return {
    id: id as CapabilityId,
    name: `Cap ${id}`,
    level: 'L1',
    parentId: parentId as CapabilityId | undefined,
    createdAt: '2024-01-01',
    _links: links,
  };
}

function makeOriginRelationship(componentId: string, originEntityId: string): OriginRelationship {
  return {
    id: `or-${componentId}-${originEntityId}` as OriginRelationshipId,
    componentId: componentId as ComponentId,
    componentName: `Component ${componentId}`,
    relationshipType: 'AcquiredVia',
    originEntityId,
    originEntityName: `Entity ${originEntityId}`,
    createdAt: '2024-01-01',
    _links: links,
  };
}

const empty: DynamicGraphData = {
  relations: [],
  capabilities: [],
  realizations: [],
  originRelationships: [],
};

function neighbors(start: EntityRef, overrides: Partial<DynamicGraphData> = {}) {
  return getNeighbors({ ...empty, ...overrides }, start);
}

describe('getNeighbors', () => {
  it('returns empty list for entity with no neighbors', () => {
    expect(neighbors({ id: 'comp-1', type: 'component' })).toEqual([]);
  });

  it('returns target component as relation neighbor when entity is the source', () => {
    const result = neighbors({ id: 'A', type: 'component' }, { relations: [makeRelation('A', 'B')] });

    expect(result).toEqual([{ id: 'B', type: 'component', edgeType: 'relation' }]);
  });

  it('returns source component as relation neighbor when entity is the target', () => {
    const result = neighbors({ id: 'B', type: 'component' }, { relations: [makeRelation('A', 'B')] });

    expect(result).toEqual([{ id: 'A', type: 'component', edgeType: 'relation' }]);
  });

  it('returns capability as realization neighbor of a component', () => {
    const result = neighbors(
      { id: 'comp-1', type: 'component' },
      { realizations: [makeRealization('comp-1', 'cap-1')] },
    );

    expect(result).toEqual([{ id: 'cap-1', type: 'capability', edgeType: 'realization' }]);
  });

  it('returns component as realization neighbor of a capability', () => {
    const result = neighbors(
      { id: 'cap-1', type: 'capability' },
      { realizations: [makeRealization('comp-1', 'cap-1')] },
    );

    expect(result).toEqual([{ id: 'comp-1', type: 'component', edgeType: 'realization' }]);
  });

  it('returns parent as parentage neighbor of a capability', () => {
    const result = neighbors(
      { id: 'cap-child', type: 'capability' },
      { capabilities: [makeCapability('cap-child', 'cap-parent'), makeCapability('cap-parent')] },
    );

    expect(result).toEqual([{ id: 'cap-parent', type: 'capability', edgeType: 'parentage' }]);
  });

  it('returns children as parentage neighbors of a capability', () => {
    const result = neighbors(
      { id: 'cap-parent', type: 'capability' },
      {
        capabilities: [
          makeCapability('cap-parent'),
          makeCapability('cap-child-1', 'cap-parent'),
          makeCapability('cap-child-2', 'cap-parent'),
        ],
      },
    );

    expect(result).toEqual([
      { id: 'cap-child-1', type: 'capability', edgeType: 'parentage' },
      { id: 'cap-child-2', type: 'capability', edgeType: 'parentage' },
    ]);
  });

  it('returns origin entity as origin neighbor of a component', () => {
    const result = neighbors(
      { id: 'comp-1', type: 'component' },
      { originRelationships: [makeOriginRelationship('comp-1', 'vendor-1')] },
    );

    expect(result).toEqual([{ id: 'vendor-1', type: 'originEntity', edgeType: 'origin' }]);
  });

  it('returns component as origin neighbor of an origin entity', () => {
    const result = neighbors(
      { id: 'vendor-1', type: 'originEntity' },
      { originRelationships: [makeOriginRelationship('comp-1', 'vendor-1')] },
    );

    expect(result).toEqual([{ id: 'comp-1', type: 'component', edgeType: 'origin' }]);
  });
});

describe('getUnexpandedByEdgeType', () => {
  const allEdges = { relation: true, realization: true, parentage: true, origin: true };
  const allTypes = { component: true, capability: true, originEntity: true };

  it('returns empty buckets when entity has no neighbors', () => {
    const result = getUnexpandedByEdgeType(empty, { id: 'comp-1', type: 'component' }, new Set(), {
      edges: allEdges,
      types: allTypes,
    });

    expect(result).toEqual({ relation: [], realization: [], parentage: [], origin: [] });
  });

  it('groups unexpanded neighbors by edge type', () => {
    const data: Partial<DynamicGraphData> = {
      relations: [makeRelation('A', 'B'), makeRelation('A', 'C')],
      realizations: [makeRealization('A', 'cap-1')],
      originRelationships: [makeOriginRelationship('A', 'vendor-1')],
    };

    const result = getUnexpandedByEdgeType({ ...empty, ...data }, { id: 'A', type: 'component' }, new Set(), {
      edges: allEdges,
      types: allTypes,
    });

    expect(result.relation).toEqual(['B', 'C']);
    expect(result.realization).toEqual(['cap-1']);
    expect(result.origin).toEqual(['vendor-1']);
    expect(result.parentage).toEqual([]);
  });

  it('excludes already-included neighbors', () => {
    const result = getUnexpandedByEdgeType(
      { ...empty, relations: [makeRelation('A', 'B'), makeRelation('A', 'C')] },
      { id: 'A', type: 'component' },
      new Set(['B']),
      { edges: allEdges, types: allTypes },
    );

    expect(result.relation).toEqual(['C']);
  });

  it.each([
    {
      name: 'edge type',
      filters: { edges: { ...allEdges, realization: false }, types: allTypes },
    },
    {
      name: 'entity type',
      filters: { edges: allEdges, types: { ...allTypes, capability: false } },
    },
  ])('excludes neighbors filtered out by disabled $name', ({ filters }) => {
    const result = getUnexpandedByEdgeType(
      {
        ...empty,
        relations: [makeRelation('A', 'B')],
        realizations: [makeRealization('A', 'cap-1')],
      },
      { id: 'A', type: 'component' },
      new Set(),
      filters,
    );

    expect(result.relation).toEqual(['B']);
    expect(result.realization).toEqual([]);
  });
});
