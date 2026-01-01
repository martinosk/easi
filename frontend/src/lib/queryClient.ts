import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      gcTime: 1000 * 60 * 30,
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

export const queryKeys = {
  components: {
    all: ['components'] as const,
    lists: () => [...queryKeys.components.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.components.lists(), filters] as const,
    details: () => [...queryKeys.components.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.components.details(), id] as const,
  },
  relations: {
    all: ['relations'] as const,
    lists: () => [...queryKeys.relations.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.relations.lists(), filters] as const,
    details: () => [...queryKeys.relations.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.relations.details(), id] as const,
  },
  views: {
    all: ['views'] as const,
    lists: () => [...queryKeys.views.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.views.lists(), filters] as const,
    details: () => [...queryKeys.views.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.views.details(), id] as const,
    components: (viewId: string) => [...queryKeys.views.detail(viewId), 'components'] as const,
  },
  capabilities: {
    all: ['capabilities'] as const,
    lists: () => [...queryKeys.capabilities.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.capabilities.lists(), filters] as const,
    details: () => [...queryKeys.capabilities.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.capabilities.details(), id] as const,
    children: (id: string) => [...queryKeys.capabilities.detail(id), 'children'] as const,
    dependencies: () => [...queryKeys.capabilities.all, 'dependencies'] as const,
    outgoing: (id: string) => [...queryKeys.capabilities.detail(id), 'outgoing'] as const,
    incoming: (id: string) => [...queryKeys.capabilities.detail(id), 'incoming'] as const,
    realizations: (id: string) => [...queryKeys.capabilities.detail(id), 'realizations'] as const,
    byComponent: (componentId: string) =>
      [...queryKeys.capabilities.all, 'byComponent', componentId] as const,
  },
  businessDomains: {
    all: ['businessDomains'] as const,
    lists: () => [...queryKeys.businessDomains.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.businessDomains.lists(), filters] as const,
    details: () => [...queryKeys.businessDomains.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.businessDomains.details(), id] as const,
    capabilities: (id: string) => [...queryKeys.businessDomains.detail(id), 'capabilities'] as const,
    capabilitiesByLink: (link: string) =>
      [...queryKeys.businessDomains.all, 'capabilitiesByLink', link] as const,
    realizations: (id: string, depth?: number) =>
      [...queryKeys.businessDomains.detail(id), 'realizations', depth] as const,
  },
  layouts: {
    all: ['layouts'] as const,
    detail: (contextType: string, contextRef: string) =>
      [...queryKeys.layouts.all, contextType, contextRef] as const,
  },
  metadata: {
    all: ['metadata'] as const,
    maturityLevels: () => [...queryKeys.metadata.all, 'maturityLevels'] as const,
    maturityScale: () => [...queryKeys.metadata.all, 'maturityScale'] as const,
    statuses: () => [...queryKeys.metadata.all, 'statuses'] as const,
    ownershipModels: () => [...queryKeys.metadata.all, 'ownershipModels'] as const,
    strategyPillars: () => [...queryKeys.metadata.all, 'strategyPillars'] as const,
    strategyPillarsConfig: () => [...queryKeys.metadata.all, 'strategyPillarsConfig'] as const,
    version: () => [...queryKeys.metadata.all, 'version'] as const,
  },
  releases: {
    all: ['releases'] as const,
    lists: () => [...queryKeys.releases.all, 'list'] as const,
    latest: () => [...queryKeys.releases.all, 'latest'] as const,
    detail: (version: string) => [...queryKeys.releases.all, 'detail', version] as const,
  },
  strategyImportance: {
    all: ['strategyImportance'] as const,
    byDomainAndCapability: (domainId: string, capabilityId: string) =>
      [...queryKeys.strategyImportance.all, 'byDomainAndCapability', domainId, capabilityId] as const,
    byDomain: (domainId: string) =>
      [...queryKeys.strategyImportance.all, 'byDomain', domainId] as const,
    byCapability: (capabilityId: string) =>
      [...queryKeys.strategyImportance.all, 'byCapability', capabilityId] as const,
  },
  enterpriseCapabilities: {
    all: ['enterpriseCapabilities'] as const,
    lists: () => [...queryKeys.enterpriseCapabilities.all, 'list'] as const,
    details: () => [...queryKeys.enterpriseCapabilities.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.enterpriseCapabilities.details(), id] as const,
    links: (id: string) => [...queryKeys.enterpriseCapabilities.detail(id), 'links'] as const,
    strategicImportance: (id: string) => [...queryKeys.enterpriseCapabilities.detail(id), 'strategicImportance'] as const,
    maturityGap: (id: string) => [...queryKeys.enterpriseCapabilities.detail(id), 'maturityGap'] as const,
  },
  maturityAnalysis: {
    all: ['maturityAnalysis'] as const,
    candidates: (sortBy?: string) => [...queryKeys.maturityAnalysis.all, 'candidates', sortBy] as const,
    unlinked: (filters?: { businessDomainId?: string; search?: string }) =>
      [...queryKeys.maturityAnalysis.all, 'unlinked', filters] as const,
  },
  fitScores: {
    all: ['fitScores'] as const,
    byComponent: (componentId: string) =>
      [...queryKeys.fitScores.all, 'byComponent', componentId] as const,
  },
  fitComparisons: {
    all: ['fitComparisons'] as const,
    byContext: (componentId: string, capabilityId: string, businessDomainId: string) =>
      [...queryKeys.fitComparisons.all, componentId, capabilityId, businessDomainId] as const,
  },
  strategicFitAnalysis: {
    all: ['strategicFitAnalysis'] as const,
    byPillar: (pillarId: string) =>
      [...queryKeys.strategicFitAnalysis.all, 'byPillar', pillarId] as const,
  },
} as const;
