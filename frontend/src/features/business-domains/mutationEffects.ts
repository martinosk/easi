import { auditQueryKeys } from '../audit/queryKeys';
import { capabilitiesQueryKeys } from '../capabilities/queryKeys';
import { businessDomainsQueryKeys, strategyImportanceQueryKeys } from './queryKeys';

export const strategyImportanceMutationEffects = {
  set: (domainId: string, capabilityId: string) => [
    strategyImportanceQueryKeys.byDomainAndCapability(domainId, capabilityId),
    strategyImportanceQueryKeys.byDomain(domainId),
    strategyImportanceQueryKeys.byCapability(capabilityId),
  ],

  update: (domainId: string, capabilityId: string) => [
    strategyImportanceQueryKeys.byDomainAndCapability(domainId, capabilityId),
    strategyImportanceQueryKeys.byDomain(domainId),
    strategyImportanceQueryKeys.byCapability(capabilityId),
  ],

  remove: (domainId: string, capabilityId: string) => [
    strategyImportanceQueryKeys.byDomainAndCapability(domainId, capabilityId),
    strategyImportanceQueryKeys.byDomain(domainId),
    strategyImportanceQueryKeys.byCapability(capabilityId),
  ],
};

export const businessDomainsMutationEffects = {
  create: () => [businessDomainsQueryKeys.lists()],

  delete: (domainId: string) => [
    businessDomainsQueryKeys.lists(),
    businessDomainsQueryKeys.detail(domainId),
  ],

  update: (domainId: string) => [
    businessDomainsQueryKeys.lists(),
    businessDomainsQueryKeys.detail(domainId),
    auditQueryKeys.history(domainId),
  ],

  associateCapability: (domainId: string, capabilityId: string) => [
    businessDomainsQueryKeys.capabilities(domainId),
    businessDomainsQueryKeys.realizations(domainId, 4),
    businessDomainsQueryKeys.detail(domainId),
    businessDomainsQueryKeys.lists(),
    capabilitiesQueryKeys.detail(capabilityId),
  ],

  dissociateCapability: (domainId: string, capabilityId: string) => [
    businessDomainsQueryKeys.capabilities(domainId),
    businessDomainsQueryKeys.realizations(domainId, 4),
    businessDomainsQueryKeys.detail(domainId),
    businessDomainsQueryKeys.lists(),
    capabilitiesQueryKeys.detail(capabilityId),
  ],
};
