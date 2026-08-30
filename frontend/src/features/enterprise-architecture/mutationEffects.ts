import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { compositionSummariesQueryKeys, enterpriseCapabilitiesQueryKeys, maturityAnalysisQueryKeys } from './queryKeys';

export const enterpriseCapabilitiesMutationEffects = {
  create: () => [
    enterpriseCapabilitiesQueryKeys.lists(),
    compositionSummariesQueryKeys.lists(),
    onePagersQueryKeys.completenessForSubjectType('enterprise-capability'),
  ],

  delete: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.lists(),
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    compositionSummariesQueryKeys.lists(),
    onePagersQueryKeys.completenessForSubjectType('enterprise-capability'),
  ],

  setTargetMaturity: (enterpriseCapabilityId: string) => [
    enterpriseCapabilitiesQueryKeys.detail(enterpriseCapabilityId),
    enterpriseCapabilitiesQueryKeys.lists(),
    maturityAnalysisQueryKeys.all,
    enterpriseCapabilitiesQueryKeys.maturityGap(enterpriseCapabilityId),
  ],
};
