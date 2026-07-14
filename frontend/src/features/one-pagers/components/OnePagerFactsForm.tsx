import { Button, Group, Stack } from '@mantine/core';
import { useOnePagerFactsForm } from '../hooks/useOnePagerFactsForm';
import type { CustomField, OnePagerFacts } from '../types';
import { FactFieldInput } from './FactFieldInput';

interface OnePagerFactsFormProps {
  fields: CustomField[];
  facts: OnePagerFacts;
}

export function OnePagerFactsForm({ fields, facts }: OnePagerFactsFormProps) {
  const { control, isDirty, isPending, requiredHint, submit } = useOnePagerFactsForm(fields, facts);

  return (
    <form onSubmit={submit}>
      <Stack gap="sm">
        {fields.map((field) => (
          <FactFieldInput key={field.id} field={field} control={control} showRequiredHint={requiredHint(field)} />
        ))}
        <Group justify="flex-end">
          <Button type="submit" size="xs" disabled={!isDirty} loading={isPending}>
            Save one-pager
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
