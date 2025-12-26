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
    version: () => [...queryKeys.metadata.all, 'version'] as const,
  },
  releases: {
    all: ['releases'] as const,
    lists: () => [...queryKeys.releases.all, 'list'] as const,
    latest: () => [...queryKeys.releases.all, 'latest'] as const,
    detail: (version: string) => [...queryKeys.releases.all, 'detail', version] as const,
  },
} as const;
