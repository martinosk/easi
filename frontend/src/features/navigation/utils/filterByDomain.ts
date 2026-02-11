import type { Capability, OriginRelationship } from '../../../api/types';
import type { FilterableArtifacts, FilteredArtifacts } from './filterByCreator';

export const UNASSIGNED_DOMAIN = '__unassigned__' as const;

export interface DomainFilterData {
  domainCapabilityIds: Map<string, string[]>;
  allCapabilities: Capability[];
  domainComponentIds: Map<string, string[]>;
  originRelationships: OriginRelationship[];
  allDomainIds: string[];
}

function isUnincludedChild(cap: Capability, included: Set<string>): boolean {
  return !included.has(cap.id) && !!cap.parentId && included.has(cap.parentId);
}

function getDescendantCapabilityIds(
  directIds: Set<string>,
  allCapabilities: Capability[]
): Set<string> {
  const result = new Set(directIds);
  let added = true;
  while (added) {
    added = false;
    for (const cap of allCapabilities) {
      if (isUnincludedChild(cap, result)) {
        result.add(cap.id);
        added = true;
      }
    }
  }
  return result;
}

function getOriginEntityIds(
  visibleComponentIds: Set<string>,
  originRelationships: OriginRelationship[]
): Set<string> {
  const result = new Set<string>();
  for (const rel of originRelationships) {
    if (visibleComponentIds.has(rel.componentId)) {
      result.add(rel.originEntityId);
    }
  }
  return result;
}

function computeVisibleForDomains(
  domainIds: string[],
  data: DomainFilterData
): Set<string> {
  const visible = new Set<string>();

  for (const domainId of domainIds) {
    const directCapIds = new Set(data.domainCapabilityIds.get(domainId) ?? []);
    const expandedCapIds = getDescendantCapabilityIds(directCapIds, data.allCapabilities);
    for (const id of expandedCapIds) visible.add(id);

    const componentIds = data.domainComponentIds.get(domainId) ?? [];
    const componentIdSet = new Set(componentIds);
    for (const id of componentIds) visible.add(id);

    const originIds = getOriginEntityIds(componentIdSet, data.originRelationships);
    for (const id of originIds) visible.add(id);
  }

  return visible;
}

function collectAllKnownArtifactIds(data: DomainFilterData): Set<string> {
  const ids = new Set<string>();
  for (const cap of data.allCapabilities) ids.add(cap.id);
  for (const compIds of data.domainComponentIds.values()) {
    for (const id of compIds) ids.add(id);
  }
  for (const rel of data.originRelationships) {
    ids.add(rel.originEntityId);
  }
  return ids;
}

function computeUnassignedIds(data: DomainFilterData): Set<string> {
  const visibleFromAll = computeVisibleForDomains(data.allDomainIds, data);
  const allKnown = collectAllKnownArtifactIds(data);

  const unassigned = new Set<string>();
  for (const id of allKnown) {
    if (!visibleFromAll.has(id)) {
      unassigned.add(id);
    }
  }
  return unassigned;
}

function mergeInto(target: Set<string>, source: Set<string>): void {
  for (const id of source) target.add(id);
}

export function computeVisibleArtifactIds(
  selectedDomainIds: string[],
  data: DomainFilterData
): Set<string> {
  const hasUnassigned = selectedDomainIds.includes(UNASSIGNED_DOMAIN);
  const realDomainIds = selectedDomainIds.filter((id) => id !== UNASSIGNED_DOMAIN);

  const visible = realDomainIds.length > 0
    ? computeVisibleForDomains(realDomainIds, data)
    : new Set<string>();

  if (hasUnassigned) {
    mergeInto(visible, computeUnassignedIds(data));
  }

  return visible;
}

export function filterByDomain(
  artifacts: FilterableArtifacts,
  selectedDomainIds: string[],
  data: DomainFilterData
): FilteredArtifacts {
  if (selectedDomainIds.length === 0) {
    return artifacts;
  }

  const hasUnassigned = selectedDomainIds.includes(UNASSIGNED_DOMAIN);
  const realDomainIds = selectedDomainIds.filter((id) => id !== UNASSIGNED_DOMAIN);

  const selectedVisible = realDomainIds.length > 0
    ? computeVisibleForDomains(realDomainIds, data)
    : new Set<string>();

  const assignedToAny = hasUnassigned
    ? computeVisibleForDomains(data.allDomainIds, data)
    : new Set<string>();

  const isVisible = (id: string) =>
    selectedVisible.has(id) || (hasUnassigned && !assignedToAny.has(id));

  return {
    components: artifacts.components.filter((c) => isVisible(c.id)),
    capabilities: artifacts.capabilities.filter((c) => isVisible(c.id)),
    acquiredEntities: artifacts.acquiredEntities.filter((e) => isVisible(e.id)),
    vendors: artifacts.vendors.filter((v) => isVisible(v.id)),
    internalTeams: artifacts.internalTeams.filter((t) => isVisible(t.id)),
  };
}
