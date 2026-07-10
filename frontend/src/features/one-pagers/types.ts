import type { HATEOASLink, HATEOASLinks } from '../../api/types';
import { onePagerFieldTypeValues } from '../../lib/schemas/onePagerConfiguration';

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
}

export interface CustomField {
  id: string;
  name: string;
  type: OnePagerFieldType;
  required: boolean;
  helpText: string;
  active: boolean;
  options?: SelectionOption[];
  _links?: CustomFieldLinks;
}

export interface BuiltInFieldLinks extends HATEOASLinks {
  'x-include'?: HATEOASLink;
  'x-exclude'?: HATEOASLink;
}

export interface BuiltInField {
  id: string;
  label: string;
  included: boolean;
  _links?: BuiltInFieldLinks;
}

export interface OnePagerConfigurationLinks extends HATEOASLinks {
  'x-define-custom-field'?: HATEOASLink;
  'x-reorder'?: HATEOASLink;
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

export interface ReorderFieldsRequest {
  order: FieldRef[];
  version: number;
}
