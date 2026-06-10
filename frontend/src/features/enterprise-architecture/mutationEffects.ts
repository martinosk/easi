import { enterpriseCapabilitiesQueryKeys, maturityAnalysisQueryKeys } from './queryKeys';

export const enterpriseCapabilitiesMutationEffects = {
  create: () => [enterpriseCapabilitiesQueryKeys.lists()],

  delete: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.lists(),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
  ],

  setTargetMaturity: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    maturityAnalysisQueryKeys.all,
    enterpriseCapabilitiesQueryKeys.maturityGap(enterpriseCapabilityId),
  ],
};
