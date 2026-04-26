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

export function computeOrphans(
  data: DynamicGraphData,
  included: ReadonlyArray<EntityRef>,
  removedId: string,
  filters: DynamicFilters,
): string[] {
  const includedIds = new Set(included.map((e) => e.id));
  const remainingIds = new Set(included.filter((e) => e.id !== removedId).map((e) => e.id));

  const hasNeighborIn = (entity: EntityRef, set: ReadonlySet<string>): boolean => {
    for (const n of getNeighbors(data, entity)) {
      if (!filters.edges[n.edgeType]) continue;
      if (!filters.types[n.type]) continue;
      if (n.id === entity.id) continue;
      if (set.has(n.id)) return true;
    }
    return false;
  };

  const orphans: string[] = [];
  for (const e of included) {
    if (e.id === removedId) continue;
    if (!hasNeighborIn(e, includedIds)) continue;
    if (!hasNeighborIn(e, remainingIds)) orphans.push(e.id);
  }
  return orphans;
}

export function getUnexpandedByEdgeType(
  data: DynamicGraphData,
  entity: EntityRef,
  included: ReadonlySet<string>,
  filters: DynamicFilters,
): UnexpandedByEdgeType {
  const out: UnexpandedByEdgeType = { relation: [], realization: [], parentage: [], origin: [] };
  for (const n of getNeighbors(data, entity)) {
    if (!filters.edges[n.edgeType]) continue;
    if (!filters.types[n.type]) continue;
    if (included.has(n.id)) continue;
    out[n.edgeType].push(n.id);
  }
  return out;
}

export function getNeighbors(data: DynamicGraphData, entity: EntityRef): EntityNeighbor[] {
  const out: EntityNeighbor[] = [];

  if (entity.type === 'component') {
    for (const r of data.relations) {
      if (r.sourceComponentId === entity.id) {
        out.push({ id: r.targetComponentId, type: 'component', edgeType: 'relation' });
      } else if (r.targetComponentId === entity.id) {
        out.push({ id: r.sourceComponentId, type: 'component', edgeType: 'relation' });
      }
    }
    for (const rz of data.realizations) {
      if (rz.componentId === entity.id) {
        out.push({ id: rz.capabilityId, type: 'capability', edgeType: 'realization' });
      }
    }
    for (const or of data.originRelationships) {
      if (or.componentId === entity.id) {
        out.push({ id: or.originEntityId, type: 'originEntity', edgeType: 'origin' });
      }
    }
  }

  if (entity.type === 'capability') {
    for (const rz of data.realizations) {
      if (rz.capabilityId === entity.id) {
        out.push({ id: rz.componentId, type: 'component', edgeType: 'realization' });
      }
    }
    for (const cap of data.capabilities) {
      if (cap.id === entity.id && cap.parentId) {
        out.push({ id: cap.parentId, type: 'capability', edgeType: 'parentage' });
      } else if (cap.parentId === entity.id) {
        out.push({ id: cap.id, type: 'capability', edgeType: 'parentage' });
      }
    }
  }

  if (entity.type === 'originEntity') {
    for (const or of data.originRelationships) {
      if (or.originEntityId === entity.id) {
        out.push({ id: or.componentId, type: 'component', edgeType: 'origin' });
      }
    }
  }

  return out;
}
