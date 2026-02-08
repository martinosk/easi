export const capabilitiesQueryKeys = {
  all: ['capabilities'] as const,
  lists: () => [...capabilitiesQueryKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) =>
    [...capabilitiesQueryKeys.lists(), filters] as const,
  details: () => [...capabilitiesQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...capabilitiesQueryKeys.details(), id] as const,
  children: (id: string) => [...capabilitiesQueryKeys.detail(id), 'children'] as const,
  dependencies: () => [...capabilitiesQueryKeys.all, 'dependencies'] as const,
  outgoing: (id: string) => [...capabilitiesQueryKeys.detail(id), 'outgoing'] as const,
  incoming: (id: string) => [...capabilitiesQueryKeys.detail(id), 'incoming'] as const,
  realizations: (id: string) => [...capabilitiesQueryKeys.detail(id), 'realizations'] as const,
  byComponent: (componentId: string) =>
    [...capabilitiesQueryKeys.all, 'byComponent', componentId] as const,
  realizationsByComponents: (componentIds?: string[]) =>
    componentIds
      ? (['realizations', 'byComponents', componentIds.sort().join(',')] as const)
      : (['realizations', 'byComponents'] as const),
  expertRoles: () => [...capabilitiesQueryKeys.all, 'expert-roles'] as const,
};
