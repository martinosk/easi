export const acquiredEntitiesQueryKeys = {
  all: ['acquiredEntities'] as const,
  lists: () => [...acquiredEntitiesQueryKeys.all, 'list'] as const,
  details: () => [...acquiredEntitiesQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...acquiredEntitiesQueryKeys.details(), id] as const,
  relationships: (id: string) => [...acquiredEntitiesQueryKeys.detail(id), 'relationships'] as const,
};

export const vendorsQueryKeys = {
  all: ['vendors'] as const,
  lists: () => [...vendorsQueryKeys.all, 'list'] as const,
  details: () => [...vendorsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...vendorsQueryKeys.details(), id] as const,
  relationships: (id: string) => [...vendorsQueryKeys.detail(id), 'relationships'] as const,
};

export const internalTeamsQueryKeys = {
  all: ['internalTeams'] as const,
  lists: () => [...internalTeamsQueryKeys.all, 'list'] as const,
  details: () => [...internalTeamsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...internalTeamsQueryKeys.details(), id] as const,
  relationships: (id: string) => [...internalTeamsQueryKeys.detail(id), 'relationships'] as const,
};

export const originRelationshipsQueryKeys = {
  all: ['originRelationships'] as const,
  lists: () => [...originRelationshipsQueryKeys.all, 'list'] as const,
};
