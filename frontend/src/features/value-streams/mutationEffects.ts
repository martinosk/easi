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

  addStage: (valueStreamId: string) => [
    valueStreamsQueryKeys.lists(),
    valueStreamsQueryKeys.detail(valueStreamId),
  ],

  updateStage: (valueStreamId: string) => [
    valueStreamsQueryKeys.detail(valueStreamId),
  ],

  deleteStage: (valueStreamId: string) => [
    valueStreamsQueryKeys.lists(),
    valueStreamsQueryKeys.detail(valueStreamId),
  ],

  reorderStages: (valueStreamId: string) => [
    valueStreamsQueryKeys.detail(valueStreamId),
  ],

  addStageCapability: (valueStreamId: string) => [
    valueStreamsQueryKeys.detail(valueStreamId),
  ],

  removeStageCapability: (valueStreamId: string) => [
    valueStreamsQueryKeys.detail(valueStreamId),
  ],
};
