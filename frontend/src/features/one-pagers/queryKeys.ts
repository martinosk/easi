import type { OnePagerSubjectType } from './types';

export const onePagersQueryKeys = {
  all: ['onePagers'] as const,
  configurations: () => [...onePagersQueryKeys.all, 'configuration'] as const,
  configuration: (subjectType: OnePagerSubjectType) => [...onePagersQueryKeys.configurations(), subjectType] as const,
};
