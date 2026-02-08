export const componentsQueryKeys = {
  all: ['components'] as const,
  lists: () => [...componentsQueryKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) =>
    [...componentsQueryKeys.lists(), filters] as const,
  details: () => [...componentsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...componentsQueryKeys.details(), id] as const,
  origins: (id: string) => [...componentsQueryKeys.detail(id), 'origins'] as const,
  expertRoles: () => [...componentsQueryKeys.all, 'expert-roles'] as const,
};

export const fitScoresQueryKeys = {
  all: ['fitScores'] as const,
  byComponent: (componentId: string) =>
    [...fitScoresQueryKeys.all, 'byComponent', componentId] as const,
};

export const fitComparisonsQueryKeys = {
  all: ['fitComparisons'] as const,
  byContext: (componentId: string, capabilityId: string, businessDomainId: string) =>
    [...fitComparisonsQueryKeys.all, componentId, capabilityId, businessDomainId] as const,
};
