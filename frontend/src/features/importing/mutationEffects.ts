import { businessDomainsQueryKeys } from '../business-domains/queryKeys';
import { capabilitiesQueryKeys } from '../capabilities/queryKeys';
import { componentsQueryKeys } from '../components/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { valueStreamsQueryKeys } from '../value-streams/queryKeys';

export const importsMutationEffects = {
  completed: () => [
    capabilitiesQueryKeys.lists(),
    capabilitiesQueryKeys.realizationsByComponents(),
    componentsQueryKeys.lists(),
    businessDomainsQueryKeys.lists(),
    valueStreamsQueryKeys.lists(),
    onePagersQueryKeys.completenessForSubjectType('application'),
    onePagersQueryKeys.completenessForSubjectType('capability'),
  ],
};
