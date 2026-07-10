import { Button, Group, Stack } from '@mantine/core';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMemo } from 'react';
import { useForm, useWatch, type Resolver } from 'react-hook-form';
import {
  buildOnePagerFactsSchema,
  factEnvelope,
  factFormDefaults,
  isFactValueEmpty,
  type OnePagerFactsFormValues,
} from '../../../lib/schemas/onePagerFacts';
import { envelopesByField } from '../factFields';
import { useSaveOnePagerFacts, type FieldValueRecord } from '../hooks/useSaveOnePagerFacts';
import type { CustomField, FieldValue, OnePagerFacts } from '../types';
import { FactFieldInput } from './FactFieldInput';

interface OnePagerFactsFormProps {
  fields: CustomField[];
  facts: OnePagerFacts;
}

interface SubmissionPlan {
  records: FieldValueRecord[];
  clears: FieldValue[];
}

function planSubmission(
  fields: CustomField[],
  facts: OnePagerFacts,
  data: OnePagerFactsFormValues,
  dirtyFields: Record<string, unknown>,
): SubmissionPlan {
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

export function OnePagerFactsForm({ fields, facts }: OnePagerFactsFormProps) {
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

  const onSubmit = (data: OnePagerFactsFormValues) => {
    const plan = planSubmission(fields, facts, data, dirtyFields);
    if (plan.records.length === 0 && plan.clears.length === 0) return;
    saveMutation.mutate({ facts, ...plan });
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      <Stack gap="sm">
        {fields.map((field) => (
          <FactFieldInput
            key={field.id}
            field={field}
            control={form.control}
            showRequiredHint={field.required && isFactValueEmpty(field.type, watched[field.id] ?? defaults[field.id])}
          />
        ))}
        <Group justify="flex-end">
          <Button type="submit" size="xs" disabled={!isDirty} loading={saveMutation.isPending}>
            Save one-pager
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
