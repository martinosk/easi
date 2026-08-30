import { HttpResponse, http } from 'msw';
import type { OnePagerCompletenessEntry, OnePagerSubjectType } from '../../features/one-pagers/types';

const entriesBySubjectType = new Map<OnePagerSubjectType, OnePagerCompletenessEntry[]>();

export function seedOnePagerCompleteness(subjectType: OnePagerSubjectType, entries: OnePagerCompletenessEntry[]): void {
  entriesBySubjectType.set(subjectType, entries);
}

export function resetOnePagerCompleteness(): void {
  entriesBySubjectType.clear();
}

export const onePagerCompletenessHandlers = [
  http.get('*/api/v1/one-pagers/:subjectType/completeness', ({ params }) => {
    const subjectType = params.subjectType as OnePagerSubjectType;
    return HttpResponse.json({
      data: entriesBySubjectType.get(subjectType) ?? [],
      _links: { self: { href: `/api/v1/one-pagers/${subjectType}/completeness`, method: 'GET' } },
    });
  }),
];
