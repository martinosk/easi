import { zodResolver } from '@hookform/resolvers/zod';
import { Button, Group, NumberInput } from '@mantine/core';
import { Controller, useForm } from 'react-hook-form';
import { type NumberFieldBoundsFormData, numberFieldBoundsSchema } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import type { CustomField } from '../types';

interface NumberFieldBoundsEditorProps {
  field: CustomField;
  onSave: (field: CustomField, min: number | undefined, max: number | undefined) => void;
}

export function NumberFieldBoundsEditor({ field, onSave }: NumberFieldBoundsEditorProps) {
  const {
    control,
    handleSubmit,
    formState: { errors, isValid, isDirty },
  } = useForm<NumberFieldBoundsFormData>({
    resolver: zodResolver(numberFieldBoundsSchema),
    values: { min: field.min ?? '', max: field.max ?? '' },
    mode: 'onChange',
  });

  if (!hasLink(field, 'x-set-bounds')) return null;

  const submit = handleSubmit((data) => {
    onSave(field, data.min === '' ? undefined : data.min, data.max === '' ? undefined : data.max);
  });

  return (
    <form onSubmit={submit} data-testid={`one-pager-bounds-form-${field.id}`}>
      <Group gap="xs" align="flex-end" pl="md">
        <Controller
          name="min"
          control={control}
          render={({ field: rhf }) => (
            <NumberInput
              label="Minimum"
              size="xs"
              value={rhf.value}
              onChange={rhf.onChange}
              data-testid={`one-pager-bounds-min-${field.id}`}
            />
          )}
        />
        <Controller
          name="max"
          control={control}
          render={({ field: rhf }) => (
            <NumberInput
              label="Maximum"
              size="xs"
              value={rhf.value}
              onChange={rhf.onChange}
              error={errors.max?.message}
              data-testid={`one-pager-bounds-max-${field.id}`}
            />
          )}
        />
        <Button
          type="submit"
          size="xs"
          variant="light"
          disabled={!isValid || !isDirty}
          data-testid={`one-pager-bounds-save-${field.id}`}
        >
          Save
        </Button>
      </Group>
    </form>
  );
}
