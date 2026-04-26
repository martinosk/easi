import type { Capability, CapabilityRealization, OriginRelationship, Relation } from '../../../api/types';

export type EntityType = 'component' | 'capability' | 'originEntity';

export type EdgeType = 'relation' | 'realization' | 'parentage' | 'origin';

export interface EntityRef {
  id: string;
  type: EntityType;
}

export interface EntityNeighbor extends EntityRef {
  edgeType: EdgeType;
}

export interface DynamicGraphData {
  relations: Relation[];
  capabilities: Capability[];
  realizations: CapabilityRealization[];
  originRelationships: OriginRelationship[];
}

export interface DynamicFilters {
  edges: Record<EdgeType, boolean>;
  types: Record<EntityType, boolean>;
}

export type UnexpandedByEdgeType = Record<EdgeType, string[]>;

function relationNeighbor(r: Relation, id: string): EntityNeighbor | null {
  if (r.sourceComponentId === id) return { id: r.targetComponentId, type: 'component', edgeType: 'relation' };
  if (r.targetComponentId === id) return { id: r.sourceComponentId, type: 'component', edgeType: 'relation' };
  return null;
}

function parentageNeighbor(cap: Capability, id: string): EntityNeighbor | null {
  if (cap.id === id && cap.parentId) return { id: cap.parentId, type: 'capability', edgeType: 'parentage' };
  if (cap.parentId === id) return { id: cap.id, type: 'capability', edgeType: 'parentage' };
  return null;
}

function compactMap<T, R>(items: readonly T[], fn: (t: T) => R | null): R[] {
  const out: R[] = [];
  for (const item of items) {
    const r = fn(item);
    if (r) out.push(r);
  }
  return out;
}

function neighborsOfComponent(data: DynamicGraphData, id: string): EntityNeighbor[] {
  return [
    ...compactMap(data.relations, (r) => relationNeighbor(r, id)),
    ...compactMap<CapabilityRealization, EntityNeighbor>(data.realizations, (rz) =>
      rz.componentId === id ? { id: rz.capabilityId, type: 'capability', edgeType: 'realization' } : null,
    ),
    ...compactMap<OriginRelationship, EntityNeighbor>(data.originRelationships, (or) =>
      or.componentId === id ? { id: or.originEntityId, type: 'originEntity', edgeType: 'origin' } : null,
    ),
  ];
}

function neighborsOfCapability(data: DynamicGraphData, id: string): EntityNeighbor[] {
  return [
    ...compactMap<CapabilityRealization, EntityNeighbor>(data.realizations, (rz) =>
      rz.capabilityId === id ? { id: rz.componentId, type: 'component', edgeType: 'realization' } : null,
    ),
    ...compactMap(data.capabilities, (cap) => parentageNeighbor(cap, id)),
  ];
}

function neighborsOfOriginEntity(data: DynamicGraphData, id: string): EntityNeighbor[] {
  return compactMap<OriginRelationship, EntityNeighbor>(data.originRelationships, (or) =>
    or.originEntityId === id ? { id: or.componentId, type: 'component', edgeType: 'origin' } : null,
  );
}

const NEIGHBOR_LOOKUP: Record<EntityType, (data: DynamicGraphData, id: string) => EntityNeighbor[]> = {
  component: neighborsOfComponent,
  capability: neighborsOfCapability,
  originEntity: neighborsOfOriginEntity,
};

export function getNeighbors(data: DynamicGraphData, entity: EntityRef): EntityNeighbor[] {
  return NEIGHBOR_LOOKUP[entity.type](data, entity.id);
}

function passesFilters(n: EntityNeighbor, filters: DynamicFilters): boolean {
  return filters.edges[n.edgeType] && filters.types[n.type];
}

export function getUnexpandedByEdgeType(
  data: DynamicGraphData,
  entity: EntityRef,
  included: ReadonlySet<string>,
  filters: DynamicFilters,
): UnexpandedByEdgeType {
  const out: UnexpandedByEdgeType = { relation: [], realization: [], parentage: [], origin: [] };
  for (const n of getNeighbors(data, entity)) {
    if (passesFilters(n, filters) && !included.has(n.id)) out[n.edgeType].push(n.id);
  }
  return out;
}
