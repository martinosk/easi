import { auditQueryKeys } from '../audit/queryKeys';
import { directionQueryKeys } from './queryKeys';

export const directionMutationEffects = {
  capture: (enterpriseCapabilityId: string) => [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    auditQueryKeys.history(enterpriseCapabilityId),
  ],

  update: (enterpriseCapabilityId: string) => [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    auditQueryKeys.history(enterpriseCapabilityId),
  ],

  propose: (enterpriseCapabilityId: string) => [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    auditQueryKeys.history(enterpriseCapabilityId),
  ],

  agree: (enterpriseCapabilityId: string) => [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    auditQueryKeys.history(enterpriseCapabilityId),
  ],

  reject: (enterpriseCapabilityId: string) => [
    directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    auditQueryKeys.history(enterpriseCapabilityId),
  ],
};
