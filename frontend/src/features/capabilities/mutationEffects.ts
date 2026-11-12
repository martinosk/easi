import { realizationRoleQueryKeys, timeAssessmentQueryKeys } from '../architecture-direction/queryKeys';
import { auditQueryKeys } from '../audit/queryKeys';
import { businessDomainsQueryKeys } from '../business-domains/queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { valueStreamsQueryKeys } from '../value-streams/queryKeys';
import { viewsQueryKeys } from '../views/queryKeys';
import { capabilitiesQueryKeys } from './queryKeys';

function onePagerFreshness(capabilityId: string) {
  return [onePagersQueryKeys.onePager('capability', capabilityId), onePagerQualityQueryKeys.lists()];
}

function realizationCompleteness() {
  return [
    onePagersQueryKeys.completenessForSubjectType('application'),
    onePagersQueryKeys.completenessForSubjectType('capability'),
  ];
}

interface RealizationLink {
  capabilityId: string;
  componentId: string;
}

function realizationLinkChanged(context: RealizationLink) {
  return [
    capabilitiesQueryKeys.realizations(context.capabilityId),
    capabilitiesQueryKeys.details(),
    capabilitiesQueryKeys.byComponent(context.componentId),
    capabilitiesQueryKeys.realizationsByComponents(),
    businessDomainsQueryKeys.details(),
    ...realizationCompleteness(),
  ];
}

export const capabilitiesMutationEffects = {
  create: (context: { parentId?: string; businessDomainId?: string }) => [
    capabilitiesQueryKeys.lists(),
    ...(context.parentId ? [capabilitiesQueryKeys.children(context.parentId)] : []),
    ...(context.businessDomainId ? [businessDomainsQueryKeys.capabilities(context.businessDomainId)] : []),
    artifactCreatorsQueryKeys.all,
    onePagersQueryKeys.completenessForSubjectType('capability'),
  ],

  update: (capabilityId: string) => [
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.detail(capabilityId),
    auditQueryKeys.history(capabilityId),
    valueStreamsQueryKeys.all,
    ...onePagerFreshness(capabilityId),
    onePagersQueryKeys.completenessForSubjectType('capability'),
  ],

  cascadeDelete: (context: { id: string; parentId?: string; domainId?: string }) => [
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.details(),
    ...(context.parentId ? [capabilitiesQueryKeys.children(context.parentId)] : []),
    capabilitiesQueryKeys.realizationsByComponents(),
    capabilitiesQueryKeys.dependencies(),
    businessDomainsQueryKeys.lists(),
    businessDomainsQueryKeys.details(),
    ...(context.domainId ? [businessDomainsQueryKeys.capabilities(context.domainId)] : []),
    valueStreamsQueryKeys.all,
    viewsQueryKeys.all,
    artifactCreatorsQueryKeys.all,
    auditQueryKeys.history(context.id),
    ...realizationCompleteness(),
  ],

  assignToDomain: (context: { capabilityId: string; domainId: string }) => [
    businessDomainsQueryKeys.capabilities(context.domainId),
    businessDomainsQueryKeys.detail(context.domainId),
    capabilitiesQueryKeys.detail(context.capabilityId),
  ],

  unassignFromDomain: (context: { capabilityId: string; domainId: string }) => [
    businessDomainsQueryKeys.capabilities(context.domainId),
    businessDomainsQueryKeys.detail(context.domainId),
    capabilitiesQueryKeys.detail(context.capabilityId),
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

  linkSystem: (context: RealizationLink) => realizationLinkChanged(context),

  updateRealization: (context: { capabilityId: string; componentId: string }) => [
    capabilitiesQueryKeys.realizations(context.capabilityId),
    capabilitiesQueryKeys.byComponent(context.componentId),
    capabilitiesQueryKeys.realizationsByComponents(),
    businessDomainsQueryKeys.details(),
  ],

  deleteRealization: (context: RealizationLink) => [
    ...realizationLinkChanged(context),
    timeAssessmentQueryKeys.all,
    realizationRoleQueryKeys.all,
  ],

  addExpert: (capabilityId: string) => [
    capabilitiesQueryKeys.detail(capabilityId),
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.expertRoles(),
    auditQueryKeys.history(capabilityId),
    ...onePagerFreshness(capabilityId),
  ],

  removeExpert: (capabilityId: string) => [
    capabilitiesQueryKeys.detail(capabilityId),
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.expertRoles(),
    auditQueryKeys.history(capabilityId),
    ...onePagerFreshness(capabilityId),
  ],

  addTag: (capabilityId: string) => [
    capabilitiesQueryKeys.detail(capabilityId),
    capabilitiesQueryKeys.lists(),
    auditQueryKeys.history(capabilityId),
  ],
};
