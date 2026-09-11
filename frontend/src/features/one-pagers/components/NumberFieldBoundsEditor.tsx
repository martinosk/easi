import { zodResolver } from '@hookform/resolvers/zod';
import { Button, Group, NumberInput } from '@mantine/core';
import { Controller, useForm } from 'react-hook-form';
import { type NumberFieldBoundsFormData, numberFieldBoundsSchema } from '../../../lib/schemas/onePagerConfiguration';
import type { SubjectAttribute } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';

interface NumberFieldBoundsEditorProps {
  attribute: SubjectAttribute;
  onSave: (attribute: SubjectAttribute, min: number | undefined, max: number | undefined) => void;
}

export function NumberFieldBoundsEditor({ attribute, onSave }: NumberFieldBoundsEditorProps) {
  const {
    control,
    handleSubmit,
    formState: { errors, isValid, isDirty },
  } = useForm<NumberFieldBoundsFormData>({
    resolver: zodResolver(numberFieldBoundsSchema),
    values: { min: attribute.min ?? '', max: attribute.max ?? '' },
    mode: 'onChange',
  });

  if (!hasLink(attribute, 'x-set-bounds')) return null;

  const submit = handleSubmit((data) => {
    onSave(attribute, data.min === '' ? undefined : data.min, data.max === '' ? undefined : data.max);
  });

  return (
    <form onSubmit={submit} data-testid={`one-pager-bounds-form-${attribute.id}`}>
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
              data-testid={`one-pager-bounds-min-${attribute.id}`}
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
              data-testid={`one-pager-bounds-max-${attribute.id}`}
            />
          )}
        />
        <Button
          type="submit"
          size="xs"
          variant="light"
          disabled={!isValid || !isDirty}
          data-testid={`one-pager-bounds-save-${attribute.id}`}
        >
          Save
        </Button>
      </Group>
    </form>
  );
}
