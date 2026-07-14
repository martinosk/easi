import type { HATEOASLink, HATEOASLinks } from '../../api/types';
import { onePagerFieldTypeValues } from '../../lib/schemas/onePagerConfiguration';
import type { ValueEnvelope } from '../../lib/schemas/onePagerFacts';

export type { ValueEnvelope } from '../../lib/schemas/onePagerFacts';

export const ONE_PAGER_SUBJECT_TYPES = [
  'capability',
  'enterprise-capability',
  'application',
  'acquired-entity',
  'vendor',
  'internal-team',
] as const;

export type OnePagerSubjectType = (typeof ONE_PAGER_SUBJECT_TYPES)[number];

export const ONE_PAGER_FIELD_TYPES = onePagerFieldTypeValues;

export type OnePagerFieldType = (typeof ONE_PAGER_FIELD_TYPES)[number];

export type FieldRefKind = 'builtIn' | 'custom';

export interface FieldRef {
  kind: FieldRefKind;
  id: string;
}

export interface SelectionOptionLinks extends HATEOASLinks {
  'x-retire'?: HATEOASLink;
}

export interface SelectionOption {
  id: string;
  label: string;
  active: boolean;
  _links?: SelectionOptionLinks;
}

export interface CustomFieldLinks extends HATEOASLinks {
  'x-rename'?: HATEOASLink;
  'x-set-requirement'?: HATEOASLink;
  'x-retire'?: HATEOASLink;
  'x-reactivate'?: HATEOASLink;
  'x-add-option'?: HATEOASLink;
  'x-set-bounds'?: HATEOASLink;
}

export interface CustomField {
  id: string;
  name: string;
  type: OnePagerFieldType;
  required: boolean;
  helpText: string;
  active: boolean;
  options?: SelectionOption[];
  min?: number;
  max?: number;
  _links?: CustomFieldLinks;
}

export interface BuiltInFieldLinks extends HATEOASLinks {
  'x-include'?: HATEOASLink;
  'x-exclude'?: HATEOASLink;
  'x-set-requirement'?: HATEOASLink;
}

export interface BuiltInField {
  id: string;
  label: string;
  included: boolean;
  required: boolean;
  _links?: BuiltInFieldLinks;
}

export type ImpactPreviewFieldKind = 'custom' | 'builtIn';

export interface OnePagerConfigurationLinks extends HATEOASLinks {
  'x-define-custom-field'?: HATEOASLink;
  'x-reorder'?: HATEOASLink;
  'x-impact-preview'?: HATEOASLink;
}

export interface OnePagerConfiguration {
  id: string;
  subjectType: OnePagerSubjectType;
  builtInFields: BuiltInField[];
  customFields: CustomField[];
  displayOrder: FieldRef[];
  version: number;
  createdAt: string;
  modifiedAt: string;
  modifiedBy: string;
  _links: OnePagerConfigurationLinks;
}

export interface VersionRequest {
  version: number;
}

export interface DefineCustomFieldRequest {
  name: string;
  fieldType: OnePagerFieldType;
  required: boolean;
  helpText: string;
  options?: string[];
  version: number;
}

export interface RenameCustomFieldRequest {
  name: string;
  fieldType: OnePagerFieldType;
  helpText: string;
  version: number;
}

export interface ChangeRequirementRequest {
  required: boolean;
  version: number;
}

export interface AddSelectionOptionRequest {
  label: string;
  version: number;
}

export interface SetNumberFieldBoundsRequest {
  min?: number;
  max?: number;
  version: number;
}

export interface ReorderFieldsRequest {
  order: FieldRef[];
  version: number;
}

export interface OnePagerImpactPreview {
  subjectType: OnePagerSubjectType;
  fieldId?: string;
  affectedSubjectCount: number;
}

export interface FieldValueLinks extends HATEOASLinks {
  'x-record'?: HATEOASLink;
  'x-clear'?: HATEOASLink;
}

export interface FieldValue {
  fieldId: string;
  value: ValueEnvelope;
  displayText: string;
  retiredOption?: boolean;
  outOfBounds?: boolean;
  modifiedAt: string;
  modifiedBy: string;
  _links?: FieldValueLinks;
}

export interface OnePagerFactsLinks extends HATEOASLinks {
  'x-record'?: HATEOASLink;
}

export interface OnePagerFacts {
  subjectType: OnePagerSubjectType;
  subjectId: string;
  values: FieldValue[];
  _links: OnePagerFactsLinks;
}

export interface RecordFieldValueRequest {
  value: ValueEnvelope;
}

export interface OnePagerViewLinks extends HATEOASLinks {
  'x-subject'?: HATEOASLink;
  'x-record'?: HATEOASLink;
}

export interface BuiltInTextValue {
  type: 'text';
  text: string;
}

export interface BuiltInDateValue {
  type: 'date';
  date: string;
}

export interface BuiltInMaturityValue {
  type: 'maturity';
  maturity: {
    value: number;
    section?: string;
  };
}

export interface BuiltInExpertView {
  name: string;
  role: string;
  contact: string;
}

export interface BuiltInExpertsValue {
  type: 'experts';
  experts: BuiltInExpertView[];
}

export interface BuiltInReference {
  id: string;
  label: string;
  subjectType?: OnePagerSubjectType;
}

export interface BuiltInReferencesValue {
  type: 'references';
  references: BuiltInReference[];
}

export type BuiltInValue =
  | BuiltInTextValue
  | BuiltInDateValue
  | BuiltInMaturityValue
  | BuiltInExpertsValue
  | BuiltInReferencesValue;

export interface BuiltInFieldView {
  id: string;
  label: string;
  value: BuiltInValue | null;
}

export interface CustomFieldView {
  fieldId: string;
  name: string;
  type: OnePagerFieldType;
  helpText?: string;
  value: ValueEnvelope | null;
  displayText: string;
  retiredOption?: boolean;
  outOfBounds?: boolean;
}

export interface OnePagerViewBuiltInField {
  kind: 'builtIn';
  builtIn: BuiltInFieldView;
}

export interface OnePagerViewCustomField {
  kind: 'custom';
  custom: CustomFieldView;
}

export type OnePagerViewField = OnePagerViewBuiltInField | OnePagerViewCustomField;

export interface OnePagerMissingField {
  fieldId: string;
  name: string;
}

export interface OnePagerCompleteness {
  requiredCount: number;
  filledCount: number;
  missingFields: OnePagerMissingField[];
}

export interface OnePagerView {
  subjectType: OnePagerSubjectType;
  subjectId: string;
  subjectName: string;
  fields: OnePagerViewField[];
  completeness: OnePagerCompleteness;
  _links: OnePagerViewLinks;
}
