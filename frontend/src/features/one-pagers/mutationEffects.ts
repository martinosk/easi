import { onePagersQueryKeys } from './queryKeys';
import type { OnePagerSubjectType } from './types';

export const onePagersMutationEffects = {
  configuration: (subjectType: OnePagerSubjectType) => [
    onePagersQueryKeys.configuration(subjectType),
    onePagersQueryKeys.viewsForSubjectType(subjectType),
  ],
  facts: (subjectType: OnePagerSubjectType, subjectId: string) => [
    onePagersQueryKeys.factsForSubject(subjectType, subjectId),
    onePagersQueryKeys.onePager(subjectType, subjectId),
  ],
};
