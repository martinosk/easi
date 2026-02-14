import { valueStreamsQueryKeys } from './queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';

export const valueStreamsMutationEffects = {
  create: () => [
    valueStreamsQueryKeys.lists(),
    artifactCreatorsQueryKeys.all,
  ],

  delete: (valueStreamId: string) => [
    valueStreamsQueryKeys.lists(),
    valueStreamsQueryKeys.detail(valueStreamId),
    artifactCreatorsQueryKeys.all,
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
