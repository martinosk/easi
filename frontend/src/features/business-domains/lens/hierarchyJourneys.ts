import type { Capability } from '../../../api/types';
import type { CapabilityJourney, TargetPeriod } from '../../journeys/types';

export interface CapabilityHierarchyJourneys {
  descendants: CapabilityJourney[];
  ancestors: CapabilityJourney[];
}

export const NO_HIERARCHY_JOURNEYS: CapabilityHierarchyJourneys = { descendants: [], ancestors: [] };

export type JourneyLookup = (capabilityId: string) => CapabilityJourney | undefined;

export interface BuildHierarchyJourneysParams {
  capabilityId: string;
  capabilities: Capability[];
  getJourney: JourneyLookup;
}

function indexById(capabilities: Capability[]): Map<string, Capability> {
  return new Map(capabilities.map((capability) => [String(capability.id), capability]));
}

function indexChildren(capabilities: Capability[]): Map<string, Capability[]> {
  const children = new Map<string, Capability[]>();
  for (const capability of capabilities) {
    if (!capability.parentId) continue;
    const parentId = String(capability.parentId);
    const siblings = children.get(parentId) ?? [];
    siblings.push(capability);
    children.set(parentId, siblings);
  }
  return children;
}

function periodRank(period: TargetPeriod | null): number {
  return period ? period.year * 4 + period.quarter : Number.POSITIVE_INFINITY;
}

function byTargetPeriodThenName(first: CapabilityJourney, second: CapabilityJourney): number {
  const firstRank = periodRank(first.targetPeriod);
  const secondRank = periodRank(second.targetPeriod);
  if (firstRank !== secondRank) return firstRank - secondRank;
  return first.capabilityName.localeCompare(second.capabilityName);
}

function isActive(journey: CapabilityJourney): boolean {
  return journey.status === 'planned' || journey.status === 'in-flight';
}

function collectDescendantJourneys(
  rootId: string,
  children: Map<string, Capability[]>,
  getJourney: JourneyLookup,
): CapabilityJourney[] {
  const journeys: CapabilityJourney[] = [];
  const visited = new Set<string>([rootId]);
  const pending = [...(children.get(rootId) ?? [])];

  for (let cursor = 0; cursor < pending.length; cursor++) {
    const capabilityId = String(pending[cursor].id);
    if (visited.has(capabilityId)) continue;
    visited.add(capabilityId);

    const journey = getJourney(capabilityId);
    if (journey) journeys.push(journey);
    pending.push(...(children.get(capabilityId) ?? []));
  }

  return journeys.sort(byTargetPeriodThenName);
}

function collectAncestorJourneys(
  capability: Capability,
  byId: Map<string, Capability>,
  getJourney: JourneyLookup,
): CapabilityJourney[] {
  const parentOf = (child: Capability) => (child.parentId ? byId.get(String(child.parentId)) : undefined);
  const journeys: CapabilityJourney[] = [];
  const visited = new Set<string>([String(capability.id)]);
  let current = parentOf(capability);

  while (current) {
    const currentId = String(current.id);
    if (visited.has(currentId)) break;
    visited.add(currentId);

    const journey = getJourney(currentId);
    if (journey && isActive(journey)) journeys.push(journey);
    current = parentOf(current);
  }

  return journeys;
}

export function buildHierarchyJourneys({
  capabilityId,
  capabilities,
  getJourney,
}: BuildHierarchyJourneysParams): CapabilityHierarchyJourneys {
  const byId = indexById(capabilities);
  const capability = byId.get(capabilityId);
  if (!capability) return NO_HIERARCHY_JOURNEYS;

  return {
    descendants: collectDescendantJourneys(capabilityId, indexChildren(capabilities), getJourney),
    ancestors: collectAncestorJourneys(capability, byId, getJourney),
  };
}
