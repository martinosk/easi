import { valueStreamsQueryKeys } from './queryKeys';

export const valueStreamsMutationEffects = {
  create: () => [
    valueStreamsQueryKeys.lists(),
  ],

  delete: (valueStreamId: string) => [
    valueStreamsQueryKeys.lists(),
    valueStreamsQueryKeys.detail(valueStreamId),
  ],

  update: (valueStreamId: string) => [
    valueStreamsQueryKeys.lists(),
    valueStreamsQueryKeys.detail(valueStreamId),
  ],
};
