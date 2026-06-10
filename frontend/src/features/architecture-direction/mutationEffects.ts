import { auditQueryKeys } from '../audit/queryKeys';
import { enterpriseCapabilitiesQueryKeys } from '../enterprise-architecture/queryKeys';
import { directionQueryKeys } from './queryKeys';

function compositionEffects(enterpriseCapabilityId: string) {
  return [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.composition(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    auditQueryKeys.history(enterpriseCapabilityId),
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
};
