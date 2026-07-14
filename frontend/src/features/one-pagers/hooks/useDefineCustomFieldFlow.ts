import { useState } from 'react';
import type { DefineCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import type { CustomField, DefineCustomFieldRequest, OnePagerConfiguration, OnePagerSubjectType } from '../types';
import { useDefineCustomField, useSetNumberFieldBounds } from './useOnePagerMutations';

function buildRequest(data: DefineCustomFieldFormData, version: number): DefineCustomFieldRequest {
  return {
    name: data.name,
    fieldType: data.fieldType,
    required: data.required,
    helpText: data.helpText,
    options: data.fieldType === 'selection' ? data.options : undefined,
    version,
  };
}

function findDefinedNumberField(configuration: OnePagerConfiguration, name: string): CustomField | undefined {
  return configuration.customFields.find((field) => field.type === 'number' && field.active && field.name === name);
}

function hasBoundsToApply(data: DefineCustomFieldFormData): boolean {
  if (data.fieldType !== 'number') return false;
  return data.min !== '' || data.max !== '';
}

export function useDefineCustomFieldFlow(subjectType: OnePagerSubjectType, configuration: OnePagerConfiguration | undefined) {
  const defineField = useDefineCustomField(subjectType);
  const setBounds = useSetNumberFieldBounds(subjectType);
  const [pendingNewField, setPendingNewField] = useState<DefineCustomFieldFormData | null>(null);
  const [formKey, setFormKey] = useState(0);

  const applyBounds = (updated: OnePagerConfiguration, data: DefineCustomFieldFormData) => {
    if (!hasBoundsToApply(data)) return;
    const field = findDefinedNumberField(updated, data.name);
    if (!field || !hasLink(field, 'x-set-bounds')) return;
    setBounds.mutate({
      field,
      request: {
        min: data.min === '' ? undefined : data.min,
        max: data.max === '' ? undefined : data.max,
        version: updated.version,
      },
    });
  };

  const submit = (data: DefineCustomFieldFormData) => {
    if (!configuration) return;
    const request = buildRequest(data, configuration.version);
    defineField.mutate(
      { configuration, request },
      {
        onSuccess: (updated) => {
          setFormKey((key) => key + 1);
          applyBounds(updated, data);
        },
      },
    );
  };

  const handleSubmit = (data: DefineCustomFieldFormData) => {
    if (data.required && hasLink(configuration, 'x-impact-preview')) {
      setPendingNewField(data);
      return;
    }
    submit(data);
  };

  const confirmPendingField = () => {
    if (!pendingNewField) return;
    submit(pendingNewField);
    setPendingNewField(null);
  };

  return {
    isSaving: defineField.isPending || setBounds.isPending,
    formKey,
    pendingNewField,
    handleSubmit,
    confirmPendingField,
    cancelPendingField: () => setPendingNewField(null),
  };
}
