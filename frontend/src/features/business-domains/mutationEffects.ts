import { businessDomainsQueryKeys, strategyImportanceQueryKeys } from './queryKeys';
import { auditQueryKeys } from '../audit/queryKeys';
import { maturityAnalysisQueryKeys } from '../enterprise-architecture/queryKeys';
import { capabilitiesQueryKeys } from '../capabilities/queryKeys';

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
  create: () => [
    businessDomainsQueryKeys.lists(),
  ],

  delete: (domainId: string) => [
    businessDomainsQueryKeys.lists(),
    businessDomainsQueryKeys.detail(domainId),
    maturityAnalysisQueryKeys.unlinked(),
  ],

  update: (domainId: string) => [
    businessDomainsQueryKeys.lists(),
    businessDomainsQueryKeys.detail(domainId),
    auditQueryKeys.history(domainId),
  ],

  associateCapability: (domainId: string, capabilityId: string) => [
    businessDomainsQueryKeys.capabilities(domainId),
    businessDomainsQueryKeys.detail(domainId),
    businessDomainsQueryKeys.lists(),
    capabilitiesQueryKeys.detail(capabilityId),
    maturityAnalysisQueryKeys.unlinked(),
  ],

  dissociateCapability: (domainId: string, capabilityId: string) => [
    businessDomainsQueryKeys.capabilities(domainId),
    businessDomainsQueryKeys.detail(domainId),
    businessDomainsQueryKeys.lists(),
    capabilitiesQueryKeys.detail(capabilityId),
    maturityAnalysisQueryKeys.unlinked(),
  ],
};
