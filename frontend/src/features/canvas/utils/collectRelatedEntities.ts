import type {
  Relation,
  Capability,
  CapabilityRealization,
  OriginRelationship,
} from '../../../api/types';

export type EntityType = 'component' | 'capability' | 'originEntity';

export interface EntityRef {
  id: string;
  type: EntityType;
}

export interface CollectedEntities {
  componentIds: Set<string>;
  capabilityIds: Set<string>;
  originEntityIds: Set<string>;
  truncated: boolean;
}

export interface TraversalData {
  relations: Relation[];
  capabilities: Capability[];
  realizations: CapabilityRealization[];
  originRelationships: OriginRelationship[];
}

const ENTITY_CAP = 500;

type AdjacencyMap = Map<string, EntityRef[]>;

function appendEdge(map: AdjacencyMap, fromKey: string, to: EntityRef): void {
  const list = map.get(fromKey);
  if (list) list.push(to);
  else map.set(fromKey, [to]);
}

function buildAdjacency(data: TraversalData): AdjacencyMap {
  const adj: AdjacencyMap = new Map();

  for (const r of data.relations) {
    appendEdge(adj, `component:${r.sourceComponentId}`, { id: r.targetComponentId, type: 'component' });
    appendEdge(adj, `component:${r.targetComponentId}`, { id: r.sourceComponentId, type: 'component' });
  }

  for (const rz of data.realizations) {
    appendEdge(adj, `component:${rz.componentId}`, { id: rz.capabilityId, type: 'capability' });
    appendEdge(adj, `capability:${rz.capabilityId}`, { id: rz.componentId, type: 'component' });
  }

  for (const or of data.originRelationships) {
    appendEdge(adj, `component:${or.componentId}`, { id: or.originEntityId, type: 'originEntity' });
    appendEdge(adj, `originEntity:${or.originEntityId}`, { id: or.componentId, type: 'component' });
  }

  for (const cap of data.capabilities) {
    if (!cap.parentId) continue;
    appendEdge(adj, `capability:${cap.id}`, { id: cap.parentId, type: 'capability' });
    appendEdge(adj, `capability:${cap.parentId}`, { id: cap.id, type: 'capability' });
  }

  return adj;
}

const resultSetKey: Record<EntityType, keyof Pick<CollectedEntities, 'componentIds' | 'capabilityIds' | 'originEntityIds'>> = {
  component: 'componentIds',
  capability: 'capabilityIds',
  originEntity: 'originEntityIds',
};

export function collectRelatedEntities(
  startEntity: EntityRef,
  data: TraversalData
): CollectedEntities {
  const adj = buildAdjacency(data);
  const visited = new Set<string>();
  const queue: EntityRef[] = [];
  const result: CollectedEntities = {
    componentIds: new Set(),
    capabilityIds: new Set(),
    originEntityIds: new Set(),
    truncated: false,
  };

  const enqueue = (entity: EntityRef): void => {
    const key = `${entity.type}:${entity.id}`;
    if (!visited.has(key)) {
      visited.add(key);
      queue.push(entity);
    }
  };

  enqueue(startEntity);
  let totalCount = 0;

  while (queue.length > 0) {
    if (totalCount >= ENTITY_CAP) {
      result.truncated = true;
      break;
    }

    const current = queue.shift()!;
    totalCount++;
    result[resultSetKey[current.type]].add(current.id);

    const key = `${current.type}:${current.id}`;
    const neighbors = adj.get(key);
    if (neighbors) neighbors.forEach(enqueue);
  }

  return result;
}
