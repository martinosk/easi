import { metadataQueryKeys } from '../../lib/appQueryKeys';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from './queryKeys';
import type { OnePagerSubjectType } from './types';

export const onePagersMutationEffects = {
  configuration: (subjectType: OnePagerSubjectType) => [
    onePagersQueryKeys.configuration(subjectType),
    onePagersQueryKeys.viewsForSubjectType(subjectType),
    onePagersQueryKeys.completenessForSubjectType(subjectType),
    onePagerQualityQueryKeys.lists(),
  ],
  attributeSchema: (subjectType: OnePagerSubjectType) => [
    metadataQueryKeys.subjectAttributeSchema(subjectType),
    ...onePagersMutationEffects.configuration(subjectType),
  ],
  facts: (subjectType: OnePagerSubjectType, subjectId: string) => [
    onePagersQueryKeys.factsForSubject(subjectType, subjectId),
    onePagersQueryKeys.onePager(subjectType, subjectId),
    onePagersQueryKeys.completenessForSubjectType(subjectType),
    onePagerQualityQueryKeys.lists(),
  ],
};
