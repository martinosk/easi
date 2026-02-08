export const businessDomainsQueryKeys = {
  all: ['businessDomains'] as const,
  lists: () => [...businessDomainsQueryKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) =>
    [...businessDomainsQueryKeys.lists(), filters] as const,
  details: () => [...businessDomainsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...businessDomainsQueryKeys.details(), id] as const,
  capabilities: (id: string) => [...businessDomainsQueryKeys.detail(id), 'capabilities'] as const,
  capabilitiesByLink: (link: string) =>
    [...businessDomainsQueryKeys.all, 'capabilitiesByLink', link] as const,
  realizations: (id: string, depth?: number) =>
    [...businessDomainsQueryKeys.detail(id), 'realizations', depth] as const,
};

export const strategyImportanceQueryKeys = {
  all: ['strategyImportance'] as const,
  byDomainAndCapability: (domainId: string, capabilityId: string) =>
    [...strategyImportanceQueryKeys.all, 'byDomainAndCapability', domainId, capabilityId] as const,
  byDomain: (domainId: string) =>
    [...strategyImportanceQueryKeys.all, 'byDomain', domainId] as const,
  byCapability: (capabilityId: string) =>
    [...strategyImportanceQueryKeys.all, 'byCapability', capabilityId] as const,
};
