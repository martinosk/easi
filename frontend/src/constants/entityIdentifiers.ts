import type {
  AcquiredEntityId,
  VendorId,
  InternalTeamId,
  OriginRelationshipType,
} from '../api/types';

export type NodeId = string & { readonly __brand: 'NodeId' };
export type EdgeId = string & { readonly __brand: 'EdgeId' };

export function toNodeId(id: string): NodeId {
  return id as NodeId;
}

export function toEdgeId(id: string): EdgeId {
  return id as EdgeId;
}

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
  nodeId: NodeId;
}

export function getEntityType(nodeId: NodeId): NodeEntityType {
  if (nodeId.startsWith(NODE_PREFIXES.capability)) return 'capability';
  if (nodeId.startsWith(NODE_PREFIXES.acquired)) return 'acquired';
  if (nodeId.startsWith(NODE_PREFIXES.vendor)) return 'vendor';
  if (nodeId.startsWith(NODE_PREFIXES.team)) return 'team';
  return 'component';
}

export function getEntityId(nodeId: NodeId): string {
  const type = getEntityType(nodeId);
  const prefix = NODE_PREFIXES[type];
  return prefix ? nodeId.slice(prefix.length) : nodeId;
}

export function parseNodeId(nodeId: NodeId): ParsedNodeId {
  const type = getEntityType(nodeId);
  const entityId = getEntityId(nodeId);
  return { type, entityId, nodeId };
}

export function makeNodeId(entityType: NodeEntityType, id: string): NodeId {
  return (NODE_PREFIXES[entityType] + id) as NodeId;
}

export function isOriginEntity(nodeId: NodeId): boolean {
  return isOriginEntityType(getEntityType(nodeId));
}

export function isCapability(nodeId: NodeId): boolean {
  return nodeId.startsWith(NODE_PREFIXES.capability);
}

export function isComponent(nodeId: NodeId): boolean {
  return getEntityType(nodeId) === 'component';
}

const ORIGIN_ENTITY_TYPES: Set<NodeEntityType> = new Set(['acquired', 'vendor', 'team']);

function isOriginEntityType(type: NodeEntityType): type is OriginEntityType {
  return ORIGIN_ENTITY_TYPES.has(type);
}

export function getOriginEntityType(nodeId: NodeId): OriginEntityType | null {
  const type = getEntityType(nodeId);
  return isOriginEntityType(type) ? type : null;
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

export function isRealizationEdge(edgeId: EdgeId): boolean {
  return edgeId.startsWith(EDGE_PREFIXES.realization);
}

export function isParentEdge(edgeId: EdgeId): boolean {
  return edgeId.startsWith(EDGE_PREFIXES.parent);
}

export function isOriginRelationshipEdge(edgeId: EdgeId): boolean {
  return edgeId.startsWith(EDGE_PREFIXES.origin);
}

export function isRelationEdge(edgeId: EdgeId): boolean {
  return !isRealizationEdge(edgeId) && !isParentEdge(edgeId) && !isOriginRelationshipEdge(edgeId);
}
