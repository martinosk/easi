import type { CapabilityLevel } from '../../../api/types';
import type { CarvedOutBy, IncludedCapabilityRole } from '../../../features/enterprise-architecture/types';

export interface StubCapability {
  id: string;
  name: string;
  level: CapabilityLevel;
  parentId?: string | null;
  businessDomainId?: string | null;
  businessDomainName?: string | null;
}

export interface ActiveDirection {
  ecId: string;
  ecName: string;
  sourceCapabilityIds: string[];
}

export interface ResolvedCapability {
  capabilityId: string;
  name: string;
  level: CapabilityLevel;
  businessDomainId: string | null;
  businessDomainName: string | null;
  role: IncludedCapabilityRole;
  carvedOutBy: CarvedOutBy | null;
}

export interface EligibilityResult {
  eligible: boolean;
  ineligibilityReason: string | null;
  conflictingEnterpriseCapability: { id: string; name: string } | null;
}

function childrenIndex(caps: StubCapability[]): Map<string, StubCapability[]> {
  const index = new Map<string, StubCapability[]>();
  for (const cap of caps) {
    if (!cap.parentId) continue;
    const siblings = index.get(cap.parentId) ?? [];
    siblings.push(cap);
    index.set(cap.parentId, siblings);
  }
  return index;
}

function toResolved(cap: StubCapability, role: IncludedCapabilityRole, carvedOutBy: CarvedOutBy | null): ResolvedCapability {
  return {
    capabilityId: cap.id,
    name: cap.name,
    level: cap.level,
    businessDomainId: cap.businessDomainId ?? null,
    businessDomainName: cap.businessDomainName ?? null,
    role,
    carvedOutBy,
  };
}

interface TraversalContext {
  capById: Map<string, StubCapability>;
  children: Map<string, StubCapability[]>;
  targetSources: Set<string>;
  owningByOtherEc: Map<string, CarvedOutBy>;
  resolved: Map<string, ResolvedCapability>;
}

function ownershipByOtherEc(directions: ActiveDirection[], targetEcId: string): Map<string, CarvedOutBy> {
  const owners = new Map<string, CarvedOutBy>();
  for (const direction of directions) {
    if (direction.ecId === targetEcId) continue;
    for (const sourceId of direction.sourceCapabilityIds) {
      owners.set(sourceId, {
        enterpriseCapabilityId: direction.ecId,
        enterpriseCapabilityName: direction.ecName,
      });
    }
  }
  return owners;
}

function visitNode(capId: string, ctx: TraversalContext): void {
  if (ctx.resolved.has(capId)) return;
  const cap = ctx.capById.get(capId);
  if (!cap) return;

  const carvedOutBy = ctx.owningByOtherEc.get(capId);
  if (carvedOutBy && !ctx.targetSources.has(capId)) {
    ctx.resolved.set(capId, toResolved(cap, 'carved-out', carvedOutBy));
    return;
  }

  ctx.resolved.set(capId, toResolved(cap, ctx.targetSources.has(capId) ? 'source' : 'implicit', null));
  for (const child of ctx.children.get(capId) ?? []) {
    visitNode(child.id, ctx);
  }
}

export function resolveComposition(
  targetEcId: string,
  directions: ActiveDirection[],
  caps: StubCapability[],
): ResolvedCapability[] {
  const target = directions.find((d) => d.ecId === targetEcId);
  if (!target) return [];

  const ctx: TraversalContext = {
    capById: new Map(caps.map((c) => [c.id, c])),
    children: childrenIndex(caps),
    targetSources: new Set(target.sourceCapabilityIds),
    owningByOtherEc: ownershipByOtherEc(directions, targetEcId),
    resolved: new Map(),
  };

  for (const sourceId of target.sourceCapabilityIds) {
    visitNode(sourceId, ctx);
  }

  return Array.from(ctx.resolved.values());
}

export function evaluateEligibility(
  capabilityId: string,
  targetEcId: string,
  directions: ActiveDirection[],
): EligibilityResult {
  const conflict = directions.find(
    (d) => d.ecId !== targetEcId && d.sourceCapabilityIds.includes(capabilityId),
  );
  if (!conflict) {
    return { eligible: true, ineligibilityReason: null, conflictingEnterpriseCapability: null };
  }
  return {
    eligible: false,
    ineligibilityReason: `Already an explicit source of an active direction on '${conflict.ecName}'`,
    conflictingEnterpriseCapability: { id: conflict.ecId, name: conflict.ecName },
  };
}
