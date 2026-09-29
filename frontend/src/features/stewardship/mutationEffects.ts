import { stewardshipQueryKeys } from './queryKeys';

export const stewardshipMutationEffects = {
  assign: (domainId: string) => [stewardshipQueryKeys.domain(domainId)],

  release: (domainId: string) => [stewardshipQueryKeys.domain(domainId)],
};
