import type { AcquiredEntity, Capability, Component, InternalTeam, Vendor } from '../../../api/types';

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

export type FilteredArtifacts = FilterableArtifacts;

export function filterEntitiesByCreator<T extends { id: string }>(
  items: T[],
  selectedCreatorIds: string[],
  creatorMap: Map<string, string>,
): T[] {
  if (selectedCreatorIds.length === 0) {
    return items;
  }
  const selectedSet = new Set(selectedCreatorIds);
  return items.filter((item) => {
    const creatorId = creatorMap.get(item.id);
    return creatorId !== undefined && selectedSet.has(creatorId);
  });
}

export function filterByCreator(
  artifacts: FilterableArtifacts,
  selectedCreatorIds: string[],
  creatorMap: Map<string, string>,
): FilteredArtifacts {
  if (selectedCreatorIds.length === 0) {
    return artifacts;
  }

  return {
    components: filterEntitiesByCreator(artifacts.components, selectedCreatorIds, creatorMap),
    capabilities: filterEntitiesByCreator(artifacts.capabilities, selectedCreatorIds, creatorMap),
    acquiredEntities: filterEntitiesByCreator(artifacts.acquiredEntities, selectedCreatorIds, creatorMap),
    vendors: filterEntitiesByCreator(artifacts.vendors, selectedCreatorIds, creatorMap),
    internalTeams: filterEntitiesByCreator(artifacts.internalTeams, selectedCreatorIds, creatorMap),
  };
}
