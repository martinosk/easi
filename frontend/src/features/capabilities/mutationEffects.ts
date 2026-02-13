import { capabilitiesQueryKeys } from './queryKeys';
import { businessDomainsQueryKeys } from '../business-domains/queryKeys';
import { maturityAnalysisQueryKeys } from '../enterprise-architecture/queryKeys';
import { auditQueryKeys } from '../audit/queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';

export const capabilitiesMutationEffects = {
  create: (context: { parentId?: string; businessDomainId?: string }) => [
    capabilitiesQueryKeys.lists(),
    ...(context.parentId ? [capabilitiesQueryKeys.children(context.parentId)] : []),
    ...(context.businessDomainId
      ? [businessDomainsQueryKeys.capabilities(context.businessDomainId)]
      : []),
    maturityAnalysisQueryKeys.unlinked(),
    artifactCreatorsQueryKeys.all,
  ],

  update: (capabilityId: string) => [
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.detail(capabilityId),
    auditQueryKeys.history(capabilityId),
  ],

  delete: (context: { id: string; parentId?: string; domainId?: string }) => [
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.detail(context.id),
    ...(context.parentId ? [capabilitiesQueryKeys.children(context.parentId)] : []),
    ...(context.domainId ? [businessDomainsQueryKeys.capabilities(context.domainId)] : []),
    businessDomainsQueryKeys.lists(),
    maturityAnalysisQueryKeys.unlinked(),
  ],

  assignToDomain: (context: { capabilityId: string; domainId: string }) => [
    businessDomainsQueryKeys.capabilities(context.domainId),
    businessDomainsQueryKeys.detail(context.domainId),
    capabilitiesQueryKeys.detail(context.capabilityId),
    maturityAnalysisQueryKeys.unlinked(),
  ],

  unassignFromDomain: (context: { capabilityId: string; domainId: string }) => [
    businessDomainsQueryKeys.capabilities(context.domainId),
    businessDomainsQueryKeys.detail(context.domainId),
    capabilitiesQueryKeys.detail(context.capabilityId),
    maturityAnalysisQueryKeys.unlinked(),
  ],

  changeParent: (context: { id: string; oldParentId?: string; newParentId?: string }) => [
    capabilitiesQueryKeys.detail(context.id),
    capabilitiesQueryKeys.details(),
    ...(context.oldParentId ? [capabilitiesQueryKeys.children(context.oldParentId)] : []),
    ...(context.newParentId ? [capabilitiesQueryKeys.children(context.newParentId)] : []),
    capabilitiesQueryKeys.lists(),
    auditQueryKeys.history(context.id),
    capabilitiesQueryKeys.realizations(context.id),
    ...(context.oldParentId ? [capabilitiesQueryKeys.realizations(context.oldParentId)] : []),
    ...(context.newParentId ? [capabilitiesQueryKeys.realizations(context.newParentId)] : []),
    capabilitiesQueryKeys.realizationsByComponents(),
    businessDomainsQueryKeys.details(),
  ],

  addDependency: (context: { sourceCapabilityId: string; targetCapabilityId: string }) => [
    capabilitiesQueryKeys.dependencies(),
    capabilitiesQueryKeys.outgoing(context.sourceCapabilityId),
    capabilitiesQueryKeys.incoming(context.targetCapabilityId),
  ],

  removeDependency: (context: { sourceCapabilityId: string; targetCapabilityId: string }) => [
    capabilitiesQueryKeys.dependencies(),
    capabilitiesQueryKeys.outgoing(context.sourceCapabilityId),
    capabilitiesQueryKeys.incoming(context.targetCapabilityId),
  ],

  linkSystem: (context: { capabilityId: string; componentId: string }) => [
    capabilitiesQueryKeys.realizations(context.capabilityId),
    capabilitiesQueryKeys.byComponent(context.componentId),
    capabilitiesQueryKeys.realizationsByComponents(),
    businessDomainsQueryKeys.details(),
  ],

  updateRealization: (context: { capabilityId: string; componentId: string }) => [
    capabilitiesQueryKeys.realizations(context.capabilityId),
    capabilitiesQueryKeys.byComponent(context.componentId),
    capabilitiesQueryKeys.realizationsByComponents(),
    businessDomainsQueryKeys.details(),
  ],

  deleteRealization: (context: { capabilityId: string; componentId: string }) => [
    capabilitiesQueryKeys.realizations(context.capabilityId),
    capabilitiesQueryKeys.byComponent(context.componentId),
    capabilitiesQueryKeys.realizationsByComponents(),
    businessDomainsQueryKeys.details(),
  ],

  addExpert: (capabilityId: string) => [
    capabilitiesQueryKeys.detail(capabilityId),
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.expertRoles(),
    auditQueryKeys.history(capabilityId),
  ],

  removeExpert: (capabilityId: string) => [
    capabilitiesQueryKeys.detail(capabilityId),
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.expertRoles(),
    auditQueryKeys.history(capabilityId),
  ],

  addTag: (capabilityId: string) => [
    capabilitiesQueryKeys.detail(capabilityId),
    capabilitiesQueryKeys.lists(),
    auditQueryKeys.history(capabilityId),
  ],
};
