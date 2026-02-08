import { enterpriseCapabilitiesQueryKeys, maturityAnalysisQueryKeys } from './queryKeys';

export const enterpriseCapabilitiesMutationEffects = {
  create: () => [
    enterpriseCapabilitiesQueryKeys.lists(),
  ],

  delete: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.lists(),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
  ],

  link: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.links(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    maturityAnalysisQueryKeys.candidates(),
    maturityAnalysisQueryKeys.unlinked(),
    enterpriseCapabilitiesQueryKeys.linkStatuses(),
  ],

  unlink: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.links(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    maturityAnalysisQueryKeys.candidates(),
    maturityAnalysisQueryKeys.unlinked(),
    enterpriseCapabilitiesQueryKeys.linkStatuses(),
  ],

  setTargetMaturity: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    maturityAnalysisQueryKeys.all,
    enterpriseCapabilitiesQueryKeys.maturityGap(enterpriseCapabilityId),
  ],
};
