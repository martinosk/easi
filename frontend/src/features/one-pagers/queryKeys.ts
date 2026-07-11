import type { OnePagerSubjectType } from './types';

export const onePagersQueryKeys = {
  all: ['onePagers'] as const,
  configurations: () => [...onePagersQueryKeys.all, 'configuration'] as const,
  configuration: (subjectType: OnePagerSubjectType) => [...onePagersQueryKeys.configurations(), subjectType] as const,
  facts: () => [...onePagersQueryKeys.all, 'facts'] as const,
  factsForSubject: (subjectType: OnePagerSubjectType, subjectId: string) =>
    [...onePagersQueryKeys.facts(), subjectType, subjectId] as const,
  views: () => [...onePagersQueryKeys.all, 'view'] as const,
  viewsForSubjectType: (subjectType: OnePagerSubjectType) => [...onePagersQueryKeys.views(), subjectType] as const,
  onePager: (subjectType: OnePagerSubjectType, subjectId: string) =>
    [...onePagersQueryKeys.viewsForSubjectType(subjectType), subjectId] as const,
};
