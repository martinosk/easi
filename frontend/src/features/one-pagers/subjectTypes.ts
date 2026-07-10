import { ONE_PAGER_SUBJECT_TYPES, type OnePagerSubjectType } from './types';

export interface SubjectTypeTab {
  value: OnePagerSubjectType;
  label: string;
}

const SUBJECT_TYPE_LABELS: Record<OnePagerSubjectType, string> = {
  capability: 'Capability',
  'enterprise-capability': 'Enterprise Capability',
  application: 'Application',
  'acquired-entity': 'Acquired Entity',
  vendor: 'Vendor',
  'internal-team': 'Internal Team',
};

export const ONE_PAGER_SUBJECT_TYPE_TABS: SubjectTypeTab[] = ONE_PAGER_SUBJECT_TYPES.map((value) => ({
  value,
  label: SUBJECT_TYPE_LABELS[value],
}));
