import type { Direction, DirectionStatus, ECDirectionResponse } from '../../../features/architecture-direction/types';
import type {
  CompositionDomainGroup,
  CompositionResponse,
  EnterpriseCapability,
  IncludedCapabilityItem,
} from '../../../features/enterprise-architecture/types';
import { toEnterpriseCapabilityId } from '../../../api/types';
import { type ActiveDirection, resolveComposition, type StubCapability } from './composition';

export interface StubEnterpriseCapability {
  id: string;
  name: string;
  description?: string;
  category?: string;
  active: boolean;
  targetMaturity?: number;
  createdAt: string;
}

export interface StubDirection {
  id: string;
  enterpriseCapabilityId: string;
  type: Direction['type'];
  status: DirectionStatus;
  horizon: Direction['horizon'];
  narrative?: string;
  sourceCapabilityIds: string[];
  createdAt: string;
  updatedAt?: string;
}

interface Spec172Db {
  enterpriseCapabilities: StubEnterpriseCapability[];
  directions: StubDirection[];
  capabilities: StubCapability[];
}

const ACTIVE_STATUSES: DirectionStatus[] = ['draft', 'proposed', 'agreed'];

function emptyDb(): Spec172Db {
  return { enterpriseCapabilities: [], directions: [], capabilities: [] };
}

let db: Spec172Db = emptyDb();

export function resetSpec172Db(): void {
  db = emptyDb();
}

export function seedSpec172Db(data: Partial<Spec172Db>): void {
  if (data.enterpriseCapabilities) db.enterpriseCapabilities = data.enterpriseCapabilities;
  if (data.directions) db.directions = data.directions;
  if (data.capabilities) db.capabilities = data.capabilities;
}

export function getStubEnterpriseCapability(id: string): StubEnterpriseCapability | undefined {
  return db.enterpriseCapabilities.find((ec) => ec.id === id);
}

export function getStubCapabilities(): StubCapability[] {
  return db.capabilities;
}

function isActive(direction: StubDirection | undefined): direction is StubDirection {
  return !!direction && ACTIVE_STATUSES.includes(direction.status);
}

export function getActiveDirection(ecId: string): StubDirection | undefined {
  return db.directions.find((d) => d.enterpriseCapabilityId === ecId && isActive(d));
}

export function upsertDirection(direction: StubDirection): void {
  const index = db.directions.findIndex((d) => d.id === direction.id);
  if (index === -1) db.directions.push(direction);
  else db.directions[index] = direction;
}

export function addDirection(direction: StubDirection): void {
  db.directions.push(direction);
}

function activeDirectionsAcrossEcs(): ActiveDirection[] {
  return db.directions.filter(isActive).map((d) => ({
    ecId: d.enterpriseCapabilityId,
    ecName: getStubEnterpriseCapability(d.enterpriseCapabilityId)?.name ?? d.enterpriseCapabilityId,
    sourceCapabilityIds: d.sourceCapabilityIds,
  }));
}

export function otherActiveDirections(targetEcId: string): ActiveDirection[] {
  return activeDirectionsAcrossEcs().filter((d) => d.ecId !== targetEcId);
}

export function allActiveDirections(): ActiveDirection[] {
  return activeDirectionsAcrossEcs();
}

const link = (href: string, method: string) => ({ href, method: method as never });

export function buildEnterpriseCapabilityDto(ec: StubEnterpriseCapability): EnterpriseCapability {
  const composition = resolveComposition(ec.id, allActiveDirections(), db.capabilities);
  const included = composition.filter((c) => c.role !== 'carved-out');
  const domainCount = new Set(included.map((c) => c.businessDomainId).filter(Boolean)).size;
  const base = `/api/v1/enterprise-capabilities/${ec.id}`;
  return {
    id: toEnterpriseCapabilityId(ec.id),
    name: ec.name,
    description: ec.description ?? '',
    category: ec.category ?? '',
    active: ec.active,
    targetMaturity: ec.targetMaturity,
    includedCapabilityCount: included.length,
    domainCount,
    createdAt: ec.createdAt,
    _links: {
      self: link(base, 'GET'),
      edit: link(base, 'PUT'),
      delete: link(base, 'DELETE'),
      'x-direction': link(`${base}/direction`, 'GET'),
      'x-composition': link(`${base}/composition`, 'GET'),
    },
  };
}

function includedItem(
  ecId: string,
  item: ReturnType<typeof resolveComposition>[number],
  directionAgreed: boolean,
): IncludedCapabilityItem {
  const links: IncludedCapabilityItem['_links'] = {
    self: link(`/api/v1/capabilities/${item.capabilityId}`, 'GET'),
  };
  if (item.role === 'source' && !directionAgreed) {
    links['x-exclude'] = link(`/api/v1/enterprise-capabilities/${ecId}/direction/sources/${item.capabilityId}`, 'DELETE');
  }
  if (item.role === 'carved-out' && item.carvedOutBy) {
    links['x-owning-ec'] = link(`/api/v1/enterprise-capabilities/${item.carvedOutBy.enterpriseCapabilityId}`, 'GET');
  }
  return {
    capabilityId: item.capabilityId,
    name: item.name,
    level: item.level,
    businessDomainId: item.businessDomainId,
    businessDomainName: item.businessDomainName,
    role: item.role,
    carvedOutBy: item.carvedOutBy,
    _links: links,
  };
}

export function buildCompositionResponse(ecId: string): CompositionResponse {
  const direction = getActiveDirection(ecId);
  const agreed = direction?.status === 'agreed';
  const resolved = resolveComposition(ecId, allActiveDirections(), db.capabilities);

  const groupsByDomain = new Map<string, CompositionDomainGroup>();
  for (const item of resolved) {
    const key = item.businessDomainId ?? '__unassigned__';
    let group = groupsByDomain.get(key);
    if (!group) {
      group = { businessDomainId: item.businessDomainId, businessDomainName: item.businessDomainName, items: [] };
      groupsByDomain.set(key, group);
    }
    group.items.push(includedItem(ecId, item, agreed));
  }

  const included = resolved.filter((c) => c.role !== 'carved-out');
  const base = `/api/v1/enterprise-capabilities/${ecId}`;
  return {
    data: Array.from(groupsByDomain.values()),
    meta: {
      sourceCount: direction?.sourceCapabilityIds.length ?? 0,
      includedCount: included.length,
      carvedOutCount: resolved.length - included.length,
      domainCount: new Set(included.map((c) => c.businessDomainId).filter(Boolean)).size,
    },
    _links: {
      self: link(`${base}/composition`, 'GET'),
      up: link(base, 'GET'),
      'x-direction': link(`${base}/direction`, 'GET'),
      ...(direction ? {} : { 'x-capture-direction': link(`${base}/direction`, 'POST') }),
    },
  };
}

function directionLinks(direction: StubDirection): Direction['_links'] {
  const base = `/api/v1/enterprise-capabilities/${direction.enterpriseCapabilityId}/direction`;
  const status = direction.status;
  const editable = status === 'draft' || status === 'proposed';
  return {
    self: link(base, 'GET'),
    up: link(`/api/v1/enterprise-capabilities/${direction.enterpriseCapabilityId}`, 'GET'),
    ...(editable && { edit: link(base, 'PUT'), 'x-add-source': link(`${base}/sources`, 'POST') }),
    ...(status === 'draft' && { 'x-propose': link(`${base}/propose`, 'POST') }),
    ...(status === 'proposed' && { 'x-agree': link(`${base}/agree`, 'POST') }),
    ...(status !== 'rejected' && { 'x-reject': link(`${base}/reject`, 'POST') }),
  };
}

function directionSourceDto(capabilityId: string) {
  const cap = db.capabilities.find((c) => c.id === capabilityId);
  return {
    id: capabilityId,
    stale: !cap,
    name: cap?.name ?? null,
    businessDomainId: cap?.businessDomainId ?? undefined,
    businessDomainName: cap?.businessDomainName ?? null,
  };
}

export function buildDirectionDto(direction: StubDirection): Direction {
  return {
    id: direction.id,
    enterpriseCapabilityId: toEnterpriseCapabilityId(direction.enterpriseCapabilityId),
    type: direction.type,
    status: direction.status,
    horizon: direction.horizon,
    narrative: direction.narrative,
    sourceCapabilities: direction.sourceCapabilityIds.map(directionSourceDto),
    placements: [],
    createdAt: direction.createdAt,
    updatedAt: direction.updatedAt,
    _links: directionLinks(direction),
  };
}

export function buildEcDirectionResponse(ecId: string): ECDirectionResponse {
  const direction = getActiveDirection(ecId);
  const base = `/api/v1/enterprise-capabilities/${ecId}`;
  return {
    direction: direction ? buildDirectionDto(direction) : null,
    _links: {
      self: link(`${base}/direction`, 'GET'),
      up: link(base, 'GET'),
      'x-composition': link(`${base}/composition`, 'GET'),
      ...(direction ? {} : { 'x-capture-direction': link(`${base}/direction`, 'POST') }),
    },
  };
}
