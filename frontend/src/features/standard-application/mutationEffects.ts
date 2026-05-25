import { auditQueryKeys } from '../audit/queryKeys';
import { standardApplicationQueryKeys } from './queryKeys';

export const standardApplicationMutationEffects = {
  set: (enterpriseCapabilityId: string) => [
    standardApplicationQueryKeys.byEnterpriseCapability(enterpriseCapabilityId),
    standardApplicationQueryKeys.historyByEnterpriseCapability(enterpriseCapabilityId),
    auditQueryKeys.history(enterpriseCapabilityId),
  ],
};
