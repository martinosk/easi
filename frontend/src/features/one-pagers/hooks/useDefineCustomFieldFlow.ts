import { useState } from 'react';
import type { DefineSubjectAttributeRequest, SubjectAttributeSchema } from '../../../api/types';
import type { DefineCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import type { OnePagerConfiguration, OnePagerSubjectType } from '../types';
import { useDefineAttribute } from './useAttributeSchemaMutations';
import { useRequirementAfterDefinition } from './useRequirementAfterDefinition';

function bound(value: number | ''): number | undefined {
  return value === '' ? undefined : value;
}

export function buildDefineRequest(data: DefineCustomFieldFormData, version: number): DefineSubjectAttributeRequest {
  const isNumber = data.fieldType === 'number';
  return {
    name: data.name,
    type: data.fieldType,
    helpText: data.helpText,
    options: data.fieldType === 'selection' ? data.options : undefined,
    min: isNumber ? bound(data.min) : undefined,
    max: isNumber ? bound(data.max) : undefined,
    version,
  };
}

function definedAttributeId(schema: SubjectAttributeSchema, name: string): string | undefined {
  return schema.attributes.find((attribute) => attribute.active && attribute.name === name)?.id;
}

export function useDefineCustomFieldFlow(
  subjectType: OnePagerSubjectType,
  configuration: OnePagerConfiguration | undefined,
  schema: SubjectAttributeSchema | undefined,
) {
  const defineAttribute = useDefineAttribute(subjectType);
  const requirement = useRequirementAfterDefinition(subjectType, configuration);
  const [pendingNewField, setPendingNewField] = useState<DefineCustomFieldFormData | null>(null);
  const [formKey, setFormKey] = useState(0);

  const submit = (data: DefineCustomFieldFormData) => {
    if (!schema) return;
    defineAttribute.mutate(
      { schema, request: buildDefineRequest(data, schema.version) },
      {
        onSuccess: (updated) => {
          setFormKey((key) => key + 1);
          if (data.required) requirement.awaitRequirementFor(definedAttributeId(updated, data.name));
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
    isSaving: defineAttribute.isPending || requirement.isPending,
    formKey,
    pendingNewField,
    handleSubmit,
    confirmPendingField,
    cancelPendingField: () => setPendingNewField(null),
  };
}
