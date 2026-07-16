import { auditQueryKeys } from '../audit/queryKeys';
import { enterpriseCapabilitiesQueryKeys } from '../enterprise-architecture/queryKeys';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { directionQueryKeys, realizationRoleQueryKeys, timeAssessmentQueryKeys } from './queryKeys';

function compositionEffects(enterpriseCapabilityId: string) {
  return [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.composition(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    auditQueryKeys.history(enterpriseCapabilityId),
    onePagersQueryKeys.onePager('enterprise-capability', enterpriseCapabilityId),
    onePagerQualityQueryKeys.lists(),
  ];
}

export const directionMutationEffects = {
  capture: compositionEffects,
  addSource: compositionEffects,
  removeSource: compositionEffects,
  update: compositionEffects,
  propose: compositionEffects,
  agree: compositionEffects,
  reject: compositionEffects,
  revert: compositionEffects,
};

function timeAssessmentEffects() {
  return [timeAssessmentQueryKeys.all];
}

export const timeAssessmentMutationEffects = {
  assess: timeAssessmentEffects,
  remove: timeAssessmentEffects,
};

function realizationRoleEffects() {
  return [realizationRoleQueryKeys.all];
}

export const realizationRoleMutationEffects = {
  assign: realizationRoleEffects,
  clear: realizationRoleEffects,
};
