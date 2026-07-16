import type { FactEnvelopesByField } from '../../lib/schemas/onePagerFacts';
import type { CustomField, CustomFieldView, FieldValue, OnePagerConfiguration, OnePagerFacts } from './types';

export function activeCustomFieldsInOrder(configuration: OnePagerConfiguration): CustomField[] {
  const fieldsById = new Map(configuration.customFields.map((field) => [field.id, field]));
  const ordered = configuration.displayOrder
    .filter((ref) => ref.kind === 'custom')
    .map((ref) => fieldsById.get(ref.id))
    .filter((field): field is CustomField => field?.active === true);
  const seen = new Set(ordered.map((field) => field.id));
  const unordered = configuration.customFields.filter((field) => field.active && !seen.has(field.id));
  return [...ordered, ...unordered];
}

export function envelopesByField(facts: OnePagerFacts): FactEnvelopesByField {
  const envelopes: FactEnvelopesByField = {};
  for (const fieldValue of facts.values) {
    envelopes[fieldValue.fieldId] = fieldValue.value;
  }
  return envelopes;
}

export interface SelectionItem {
  value: string;
  label: string;
}

export function selectionItems(field: CustomField, currentOptionId: string): SelectionItem[] {
  const options = field.options ?? [];
  const items = options.filter((option) => option.active).map((option) => ({ value: option.id, label: option.label }));
  if (currentOptionId && !items.some((item) => item.value === currentOptionId)) {
    const current = options.find((option) => option.id === currentOptionId);
    if (current) items.push({ value: current.id, label: `${current.label} (retired)` });
  }
  return items;
}

export interface CustomFieldViewDisplayProps {
  field: CustomField;
  fieldValue?: FieldValue;
}

export function customFieldViewDisplayProps(view: CustomFieldView): CustomFieldViewDisplayProps {
  const field: CustomField = {
    id: view.fieldId,
    name: view.name,
    type: view.type,
    required: false,
    helpText: view.helpText ?? '',
    active: true,
  };
  if (!view.value) return { field };
  const fieldValue: FieldValue = {
    fieldId: view.fieldId,
    value: view.value,
    displayText: view.displayText,
    retiredOption: view.retiredOption,
    outOfBounds: view.outOfBounds,
    modifiedAt: '',
    modifiedBy: '',
  };
  return { field, fieldValue };
}
