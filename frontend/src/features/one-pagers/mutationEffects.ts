import { onePagersQueryKeys } from './queryKeys';
import type { OnePagerSubjectType } from './types';

export const onePagersMutationEffects = {
  configuration: (subjectType: OnePagerSubjectType) => [onePagersQueryKeys.configuration(subjectType)],
};
