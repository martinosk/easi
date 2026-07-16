import { zodResolver } from '@hookform/resolvers/zod';
import { useMemo } from 'react';
import { useForm, useWatch, type Control, type Resolver } from 'react-hook-form';
import {
  buildOnePagerFactsSchema,
  factEnvelope,
  factFormDefaults,
  isFactValueEmpty,
  type FactFormValue,
  type OnePagerFactsFormValues,
} from '../../../lib/schemas/onePagerFacts';
import { envelopesByField } from '../factFields';
import type { CustomField, FieldValue, OnePagerFacts } from '../types';
import { useSaveOnePagerFacts, type FieldValueRecord } from './useSaveOnePagerFacts';

export interface OnePagerFactsSubmissionPlan {
  records: FieldValueRecord[];
  clears: FieldValue[];
}

export function planOnePagerFactsSubmission(
  fields: CustomField[],
  facts: OnePagerFacts,
  data: OnePagerFactsFormValues,
  dirtyFields: Record<string, unknown>,
): OnePagerFactsSubmissionPlan {
  const valuesByFieldId = new Map(facts.values.map((fieldValue) => [fieldValue.fieldId, fieldValue]));
  const records: FieldValueRecord[] = [];
  const clears: FieldValue[] = [];
  for (const field of fields) {
    if (!dirtyFields[field.id]) continue;
    const envelope = factEnvelope(field, data[field.id]);
    if (envelope) {
      records.push({ fieldId: field.id, value: envelope });
      continue;
    }
    const existing = valuesByFieldId.get(field.id);
    if (existing) clears.push(existing);
  }
  return { records, clears };
}

export interface UseOnePagerFactsFormResult {
  control: Control<OnePagerFactsFormValues>;
  isDirty: boolean;
  isPending: boolean;
  requiredHint: (field: CustomField) => boolean;
  submit: (event?: React.BaseSyntheticEvent) => Promise<void>;
  cancel: () => void;
}

export function useOnePagerFactsForm(
  fields: CustomField[],
  facts: OnePagerFacts,
  onSaved?: () => void,
): UseOnePagerFactsFormResult {
  const envelopes = useMemo(() => envelopesByField(facts), [facts]);
  const schema = useMemo(() => buildOnePagerFactsSchema(fields, envelopes), [fields, envelopes]);
  const defaults = useMemo(() => factFormDefaults(fields, envelopes), [fields, envelopes]);

  const form = useForm<OnePagerFactsFormValues>({
    resolver: zodResolver(schema) as Resolver<OnePagerFactsFormValues>,
    values: defaults,
    mode: 'onChange',
  });
  const watched = useWatch({ control: form.control });
  const { dirtyFields, isDirty } = form.formState;
  const saveMutation = useSaveOnePagerFacts(facts.subjectType, facts.subjectId);

  const submit = form.handleSubmit((data) => {
    const plan = planOnePagerFactsSubmission(fields, facts, data, dirtyFields);
    if (plan.records.length === 0 && plan.clears.length === 0) return;
    saveMutation.mutate({ facts, ...plan }, { onSuccess: onSaved });
  });

  const cancel = () => form.reset(defaults);

  const requiredHint = (field: CustomField) =>
    field.required && isFactValueEmpty(field.type, (watched[field.id] ?? defaults[field.id]) as FactFormValue);

  return { control: form.control, isDirty, isPending: saveMutation.isPending, requiredHint, submit, cancel };
}
