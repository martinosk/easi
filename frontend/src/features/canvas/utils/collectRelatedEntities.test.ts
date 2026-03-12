import { describe, it, expect } from 'vitest';
import { collectRelatedEntities, type EntityRef, type TraversalData } from './collectRelatedEntities';
import type {
  Relation,
  Capability,
  CapabilityRealization,
  OriginRelationship,
  ComponentId,
  CapabilityId,
  RelationId,
  RealizationId,
  OriginRelationshipId,
  HATEOASLinks,
} from '../../../api/types';

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

const emptyData: TraversalData = {
  relations: [],
  capabilities: [],
  realizations: [],
  originRelationships: [],
};

function collect(start: EntityRef, overrides: Partial<TraversalData> = {}) {
  return collectRelatedEntities(start, { ...emptyData, ...overrides });
}

describe('collectRelatedEntities', () => {
  it('returns singleton when entity has no relations', () => {
    const result = collect({ id: 'comp-1', type: 'component' });

    expect(result.componentIds).toEqual(new Set(['comp-1']));
    expect(result.capabilityIds).toEqual(new Set());
    expect(result.originEntityIds).toEqual(new Set());
    expect(result.truncated).toBe(false);
  });

  it('traverses linear chain A->B->C', () => {
    const result = collect(
      { id: 'A', type: 'component' },
      { relations: [makeRelation('A', 'B'), makeRelation('B', 'C')] },
    );

    expect(result.componentIds).toEqual(new Set(['A', 'B', 'C']));
    expect(result.truncated).toBe(false);
  });

  it('handles cycle A->B->C->A', () => {
    const result = collect(
      { id: 'A', type: 'component' },
      { relations: [makeRelation('A', 'B'), makeRelation('B', 'C'), makeRelation('C', 'A')] },
    );

    expect(result.componentIds).toEqual(new Set(['A', 'B', 'C']));
    expect(result.truncated).toBe(false);
  });

  it('traverses component with realization to capability', () => {
    const result = collect(
      { id: 'comp-1', type: 'component' },
      { realizations: [makeRealization('comp-1', 'cap-1')], capabilities: [makeCapability('cap-1')] },
    );

    expect(result.componentIds).toEqual(new Set(['comp-1']));
    expect(result.capabilityIds).toEqual(new Set(['cap-1']));
  });

  it('traverses component with origin relationship', () => {
    const result = collect(
      { id: 'comp-1', type: 'component' },
      { originRelationships: [makeOriginRelationship('comp-1', 'vendor-1')] },
    );

    expect(result.componentIds).toEqual(new Set(['comp-1']));
    expect(result.originEntityIds).toEqual(new Set(['vendor-1']));
  });

  it('traverses mixed graph crossing all relation types', () => {
    const result = collect({ id: 'comp-1', type: 'component' }, {
      relations: [makeRelation('comp-1', 'comp-2')],
      capabilities: [makeCapability('cap-1'), makeCapability('cap-2', 'cap-1')],
      realizations: [makeRealization('comp-1', 'cap-1'), makeRealization('comp-2', 'cap-2')],
      originRelationships: [makeOriginRelationship('comp-2', 'acq-1')],
    });

    expect(result.componentIds).toEqual(new Set(['comp-1', 'comp-2']));
    expect(result.capabilityIds).toEqual(new Set(['cap-1', 'cap-2']));
    expect(result.originEntityIds).toEqual(new Set(['acq-1']));
    expect(result.truncated).toBe(false);
  });

  it('traverses bidirectionally from capability to component', () => {
    const result = collect(
      { id: 'cap-1', type: 'capability' },
      { realizations: [makeRealization('comp-1', 'cap-1')], capabilities: [makeCapability('cap-1')] },
    );

    expect(result.capabilityIds).toEqual(new Set(['cap-1']));
    expect(result.componentIds).toEqual(new Set(['comp-1']));
  });

  it('traverses bidirectionally from origin entity to component', () => {
    const result = collect(
      { id: 'vendor-1', type: 'originEntity' },
      { originRelationships: [makeOriginRelationship('comp-1', 'vendor-1')] },
    );

    expect(result.originEntityIds).toEqual(new Set(['vendor-1']));
    expect(result.componentIds).toEqual(new Set(['comp-1']));
  });

  it('traverses capability parent-child relationships', () => {
    const result = collect(
      { id: 'cap-1', type: 'capability' },
      { capabilities: [makeCapability('cap-1'), makeCapability('cap-2', 'cap-1'), makeCapability('cap-3', 'cap-1')] },
    );

    expect(result.capabilityIds).toEqual(new Set(['cap-1', 'cap-2', 'cap-3']));
  });

  it('caps at 500 entities with truncated flag', () => {
    const relations: Relation[] = [];
    for (let i = 0; i < 600; i++) {
      relations.push(makeRelation(`c${i}`, `c${i + 1}`));
    }

    const result = collect({ id: 'c0', type: 'component' }, { relations });
    const total = result.componentIds.size + result.capabilityIds.size + result.originEntityIds.size;
    expect(total).toBe(500);
    expect(result.truncated).toBe(true);
  });
});
