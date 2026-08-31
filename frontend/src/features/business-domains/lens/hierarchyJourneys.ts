import type { Capability } from '../../../api/types';
import type { CapabilityJourney } from '../../journeys/types';
import { byTargetPeriodThenName } from '../../journeys/utils/period';

export interface CapabilityHierarchyJourneys {
  descendants: CapabilityJourney[];
  ancestors: CapabilityJourney[];
}

export const NO_HIERARCHY_JOURNEYS: CapabilityHierarchyJourneys = { descendants: [], ancestors: [] };

export type JourneyLookup = (capabilityId: string) => CapabilityJourney[];

export interface BuildHierarchyJourneysParams {
  capabilityId: string;
  capabilities: Capability[];
  getJourneys: JourneyLookup;
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

function isActive(journey: CapabilityJourney): boolean {
  return journey.status === 'planned' || journey.status === 'in-flight';
}

function collectDescendantJourneys(
  rootId: string,
  children: Map<string, Capability[]>,
  getJourneys: JourneyLookup,
): CapabilityJourney[] {
  const journeys: CapabilityJourney[] = [];
  const visited = new Set<string>([rootId]);
  const pending = [...(children.get(rootId) ?? [])];

  for (let cursor = 0; cursor < pending.length; cursor++) {
    const capabilityId = String(pending[cursor].id);
    if (visited.has(capabilityId)) continue;
    visited.add(capabilityId);

    journeys.push(...getJourneys(capabilityId));
    pending.push(...(children.get(capabilityId) ?? []));
  }

  return journeys.sort(byTargetPeriodThenName);
}

function collectAncestorJourneys(
  capability: Capability,
  byId: Map<string, Capability>,
  getJourneys: JourneyLookup,
): CapabilityJourney[] {
  const parentOf = (child: Capability) => (child.parentId ? byId.get(String(child.parentId)) : undefined);
  const journeys: CapabilityJourney[] = [];
  const visited = new Set<string>([String(capability.id)]);
  let current = parentOf(capability);

  while (current) {
    const currentId = String(current.id);
    if (visited.has(currentId)) break;
    visited.add(currentId);

    journeys.push(...getJourneys(currentId).filter(isActive));
    current = parentOf(current);
  }

  return journeys;
}

export function buildHierarchyJourneys({
  capabilityId,
  capabilities,
  getJourneys,
}: BuildHierarchyJourneysParams): CapabilityHierarchyJourneys {
  const byId = indexById(capabilities);
  const capability = byId.get(capabilityId);
  if (!capability) return NO_HIERARCHY_JOURNEYS;

  return {
    descendants: collectDescendantJourneys(capabilityId, indexChildren(capabilities), getJourneys),
    ancestors: collectAncestorJourneys(capability, byId, getJourneys),
  };
}
