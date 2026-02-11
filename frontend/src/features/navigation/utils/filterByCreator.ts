import type { Component, Capability, AcquiredEntity, Vendor, InternalTeam } from '../../../api/types';

export interface ArtifactCreator {
  aggregateId: string;
  creatorId: string;
}

export interface FilterableArtifacts {
  components: Component[];
  capabilities: Capability[];
  acquiredEntities: AcquiredEntity[];
  vendors: Vendor[];
  internalTeams: InternalTeam[];
}

export interface FilteredArtifacts extends FilterableArtifacts {}

function filterById<T extends { id: string }>(
  items: T[],
  selectedCreatorIds: string[],
  creatorMap: Map<string, string>
): T[] {
  const selectedSet = new Set(selectedCreatorIds);
  return items.filter((item) => {
    const creatorId = creatorMap.get(item.id);
    return creatorId !== undefined && selectedSet.has(creatorId);
  });
}

export function filterByCreator(
  artifacts: FilterableArtifacts,
  selectedCreatorIds: string[],
  creatorMap: Map<string, string>
): FilteredArtifacts {
  if (selectedCreatorIds.length === 0) {
    return artifacts;
  }

  return {
    components: filterById(artifacts.components, selectedCreatorIds, creatorMap),
    capabilities: filterById(artifacts.capabilities, selectedCreatorIds, creatorMap),
    acquiredEntities: filterById(artifacts.acquiredEntities, selectedCreatorIds, creatorMap),
    vendors: filterById(artifacts.vendors, selectedCreatorIds, creatorMap),
    internalTeams: filterById(artifacts.internalTeams, selectedCreatorIds, creatorMap),
  };
}
