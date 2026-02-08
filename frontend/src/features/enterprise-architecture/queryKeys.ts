export const enterpriseCapabilitiesQueryKeys = {
  all: ['enterpriseCapabilities'] as const,
  lists: () => [...enterpriseCapabilitiesQueryKeys.all, 'list'] as const,
  details: () => [...enterpriseCapabilitiesQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...enterpriseCapabilitiesQueryKeys.details(), id] as const,
  links: (id: string) => [...enterpriseCapabilitiesQueryKeys.detail(id), 'links'] as const,
  strategicImportance: (id: string) => [...enterpriseCapabilitiesQueryKeys.detail(id), 'strategicImportance'] as const,
  maturityGap: (id: string) => [...enterpriseCapabilitiesQueryKeys.detail(id), 'maturityGap'] as const,
  linkStatuses: () => ['linkStatuses'] as const,
};

export const maturityAnalysisQueryKeys = {
  all: ['maturityAnalysis'] as const,
  candidates: (sortBy?: string) => [...maturityAnalysisQueryKeys.all, 'candidates', sortBy] as const,
  unlinked: (filters?: { businessDomainId?: string; search?: string }) =>
    [...maturityAnalysisQueryKeys.all, 'unlinked', filters] as const,
};

export const strategicFitAnalysisQueryKeys = {
  all: ['strategicFitAnalysis'] as const,
  byPillar: (pillarId: string) =>
    [...strategicFitAnalysisQueryKeys.all, 'byPillar', pillarId] as const,
};

export const timeSuggestionsQueryKeys = {
  all: ['timeSuggestions'] as const,
  list: (filters?: { capabilityId?: string; componentId?: string }) =>
    [...timeSuggestionsQueryKeys.all, 'list', filters] as const,
};
