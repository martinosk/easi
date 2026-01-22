import type {
  AcquiredEntityId,
  VendorId,
  InternalTeamId,
  OriginRelationshipType,
} from '../api/types';

export const NODE_PREFIXES = {
  component: '',
  capability: 'cap-',
  acquired: 'acq-',
  vendor: 'vendor-',
  team: 'team-',
} as const;

export const EDGE_PREFIXES = {
  parent: 'parent-',
  realization: 'realization-',
  origin: 'origin-',
} as const;

export type NodeEntityType = keyof typeof NODE_PREFIXES;
export type OriginEntityType = 'acquired' | 'vendor' | 'team';

export interface ParsedNodeId {
  type: NodeEntityType;
  entityId: string;
  nodeId: string;
}

export function getEntityType(nodeId: string): NodeEntityType {
  if (nodeId.startsWith(NODE_PREFIXES.capability)) return 'capability';
  if (nodeId.startsWith(NODE_PREFIXES.acquired)) return 'acquired';
  if (nodeId.startsWith(NODE_PREFIXES.vendor)) return 'vendor';
  if (nodeId.startsWith(NODE_PREFIXES.team)) return 'team';
  return 'component';
}

export function getEntityId(nodeId: string): string {
  const type = getEntityType(nodeId);
  const prefix = NODE_PREFIXES[type];
  return prefix ? nodeId.slice(prefix.length) : nodeId;
}

export function parseNodeId(nodeId: string): ParsedNodeId {
  const type = getEntityType(nodeId);
  const entityId = getEntityId(nodeId);
  return { type, entityId, nodeId };
}

export function makeNodeId(entityType: NodeEntityType, id: string): string {
  return NODE_PREFIXES[entityType] + id;
}

export function isOriginEntity(nodeId: string): boolean {
  const type = getEntityType(nodeId);
  return type === 'acquired' || type === 'vendor' || type === 'team';
}

export function isCapability(nodeId: string): boolean {
  return nodeId.startsWith(NODE_PREFIXES.capability);
}

export function isComponent(nodeId: string): boolean {
  return getEntityType(nodeId) === 'component';
}

export function getOriginEntityType(nodeId: string): OriginEntityType | null {
  const type = getEntityType(nodeId);
  if (type === 'acquired' || type === 'vendor' || type === 'team') {
    return type;
  }
  return null;
}

export type OriginEntityIdMap = {
  acquired: AcquiredEntityId;
  vendor: VendorId;
  team: InternalTeamId;
};

export function toTypedEntityId<T extends OriginEntityType>(
  _type: T,
  id: string
): OriginEntityIdMap[T] {
  return id as OriginEntityIdMap[T];
}

export const ORIGIN_RELATIONSHIP_TYPE_MAP: Record<OriginEntityType, OriginRelationshipType> = {
  acquired: 'AcquiredVia',
  vendor: 'PurchasedFrom',
  team: 'BuiltBy',
};

export const ORIGIN_RELATIONSHIP_LABELS: Record<OriginRelationshipType, string> = {
  AcquiredVia: 'Acquired via',
  PurchasedFrom: 'Purchased from',
  BuiltBy: 'Built by',
};

export function isRealizationEdge(edgeId: string): boolean {
  return edgeId.startsWith(EDGE_PREFIXES.realization);
}

export function isParentEdge(edgeId: string): boolean {
  return edgeId.startsWith(EDGE_PREFIXES.parent);
}

export function isOriginRelationshipEdge(edgeId: string): boolean {
  return edgeId.startsWith(EDGE_PREFIXES.origin);
}

export function isRelationEdge(edgeId: string): boolean {
  return !isRealizationEdge(edgeId) && !isParentEdge(edgeId) && !isOriginRelationshipEdge(edgeId);
}
