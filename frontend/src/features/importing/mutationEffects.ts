import { capabilitiesQueryKeys } from '../capabilities/queryKeys';
import { componentsQueryKeys } from '../components/queryKeys';
import { businessDomainsQueryKeys } from '../business-domains/queryKeys';
import { maturityAnalysisQueryKeys } from '../enterprise-architecture/queryKeys';

export const importsMutationEffects = {
  completed: () => [
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.realizationsByComponents(),
    componentsQueryKeys.lists(),
    businessDomainsQueryKeys.lists(),
    maturityAnalysisQueryKeys.unlinked(),
  ],
};
