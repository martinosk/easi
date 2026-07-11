import { useState } from 'react';
import type { DefineCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import type { DefineCustomFieldRequest, OnePagerConfiguration, OnePagerSubjectType } from '../types';
import { useDefineCustomField } from './useOnePagerMutations';

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

export function useDefineCustomFieldFlow(subjectType: OnePagerSubjectType, configuration: OnePagerConfiguration | undefined) {
  const defineField = useDefineCustomField(subjectType);
  const [pendingNewField, setPendingNewField] = useState<DefineCustomFieldFormData | null>(null);
  const [formKey, setFormKey] = useState(0);

  const submit = (data: DefineCustomFieldFormData) => {
    if (!configuration) return;
    const request = buildRequest(data, configuration.version);
    defineField.mutate({ configuration, request }, { onSuccess: () => setFormKey((key) => key + 1) });
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
    isSaving: defineField.isPending,
    formKey,
    pendingNewField,
    handleSubmit,
    confirmPendingField,
    cancelPendingField: () => setPendingNewField(null),
  };
}
