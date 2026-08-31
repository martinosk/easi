export const enterpriseCapabilitiesQueryKeys = {
  all: ['enterpriseCapabilities'] as const,
  lists: () => [...enterpriseCapabilitiesQueryKeys.all, 'list'] as const,
  details: () => [...enterpriseCapabilitiesQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...enterpriseCapabilitiesQueryKeys.details(), id] as const,
  composition: (id: string) => [...enterpriseCapabilitiesQueryKeys.detail(id), 'composition'] as const,
  strategicImportance: (id: string) => [...enterpriseCapabilitiesQueryKeys.detail(id), 'strategicImportance'] as const,
  maturityGap: (id: string) => [...enterpriseCapabilitiesQueryKeys.detail(id), 'maturityGap'] as const,
};

export const compositionSummariesQueryKeys = {
  all: ['enterpriseCapabilityCompositions'] as const,
  lists: () => [...compositionSummariesQueryKeys.all, 'list'] as const,
};

export const maturityAnalysisQueryKeys = {
  all: ['maturityAnalysis'] as const,
  candidates: (sortBy?: string) => [...maturityAnalysisQueryKeys.all, 'candidates', sortBy] as const,
  unlinked: (filters?: { businessDomainId?: string; search?: string }) =>
    [...maturityAnalysisQueryKeys.all, 'unlinked', filters] as const,
};

export const strategicFitAnalysisQueryKeys = {
  all: ['strategicFitAnalysis'] as const,
  byPillar: (pillarId: string) => [...strategicFitAnalysisQueryKeys.all, 'byPillar', pillarId] as const,
};

export const timeSuggestionsQueryKeys = {
  all: ['timeSuggestions'] as const,
  list: (filters?: { capabilityId?: string; componentId?: string }) =>
    [...timeSuggestionsQueryKeys.all, 'list', filters] as const,
};
