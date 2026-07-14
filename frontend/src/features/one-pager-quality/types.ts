import type { HATEOASLink, HATEOASLinks, PaginatedResponse } from '../../api/types';
import type { OnePagerSubjectType } from '../one-pagers/types';

export type { OnePagerSubjectType } from '../one-pagers/types';

export type OnePagerQualityCompleteness = 'complete' | 'incomplete' | 'not-applicable';

export const ONE_PAGER_QUALITY_SORTS = ['completeness', 'creator', 'name', 'created', 'updated'] as const;

export type OnePagerQualitySort = (typeof ONE_PAGER_QUALITY_SORTS)[number];

export type OnePagerQualityOrder = 'asc' | 'desc';

export interface OnePagerQualityRowLinks extends HATEOASLinks {
  'x-edit-grants'?: HATEOASLink;
}

export interface OnePagerQualityRow {
  subjectType: OnePagerSubjectType;
  subjectId: string;
  name: string;
  completeness: OnePagerQualityCompleteness;
  requiredCount: number;
  filledCount: number;
  missingCount: number;
  creatorId: string;
  creatorEmail: string;
  createdAt: string;
  lastUpdatedAt: string;
  _links?: OnePagerQualityRowLinks;
}

export type OnePagerQualityResponse = PaginatedResponse<OnePagerQualityRow>;
