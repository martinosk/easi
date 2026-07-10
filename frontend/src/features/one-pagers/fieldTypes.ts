import { ONE_PAGER_FIELD_TYPES, type OnePagerFieldType } from './types';

const FIELD_TYPE_LABELS: Record<OnePagerFieldType, string> = {
  text: 'Text',
  number: 'Number',
  date: 'Date',
  link: 'Link',
  selection: 'Selection',
  'contact-person': 'Contact Person',
};

export const ONE_PAGER_FIELD_TYPE_OPTIONS = ONE_PAGER_FIELD_TYPES.map((value) => ({
  value,
  label: FIELD_TYPE_LABELS[value],
}));
