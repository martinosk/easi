import { ONE_PAGER_SUBJECT_TYPES, type OnePagerSubjectType } from './types';

export interface SubjectTypeTab {
  value: OnePagerSubjectType;
  label: string;
}

const SUBJECT_TYPE_LABELS: Record<OnePagerSubjectType, string> = {
  capability: 'Capability',
  application: 'Application',
  'acquired-entity': 'Acquired Entity',
  vendor: 'Vendor',
  'internal-team': 'Internal Team',
};

export const ONE_PAGER_SUBJECT_TYPE_TABS: SubjectTypeTab[] = ONE_PAGER_SUBJECT_TYPES.map((value) => ({
  value,
  label: SUBJECT_TYPE_LABELS[value],
}));

export function subjectTypeLabel(subjectType: OnePagerSubjectType): string {
  return SUBJECT_TYPE_LABELS[subjectType];
}

function pluralize(label: string): string {
  if (/[^aeiou]y$/i.test(label)) return `${label.slice(0, -1)}ies`;
  return `${label}s`;
}

export function pluralSubjectTypeLabel(subjectType: OnePagerSubjectType): string {
  return pluralize(SUBJECT_TYPE_LABELS[subjectType]);
}
