import { zodResolver } from '@hookform/resolvers/zod';
import { ActionIcon, Badge, Group, Stack, TextInput } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { type AddSelectionOptionFormData, addSelectionOptionSchema } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import type { CustomField, SelectionOption } from '../types';

interface SelectionOptionsEditorProps {
  field: CustomField;
  onAddOption: (field: CustomField, label: string) => void;
  onRetireOption: (option: SelectionOption) => void;
}

const DEFAULT_VALUES: AddSelectionOptionFormData = { label: '' };

function OptionBadge({
  option,
  onRetireOption,
}: {
  option: SelectionOption;
  onRetireOption: (option: SelectionOption) => void;
}) {
  return (
    <Badge
      key={option.id}
      variant={option.active ? 'light' : 'outline'}
      color={option.active ? 'blue' : 'gray'}
      data-testid={`one-pager-option-${option.id}`}
      rightSection={
        hasLink(option, 'x-retire') ? (
          <ActionIcon
            size="xs"
            variant="transparent"
            color="gray"
            aria-label={`Retire option ${option.label}`}
            onClick={() => onRetireOption(option)}
          >
            ×
          </ActionIcon>
        ) : undefined
      }
    >
      {option.label}
    </Badge>
  );
}

export function SelectionOptionsEditor({ field, onAddOption, onRetireOption }: SelectionOptionsEditorProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<AddSelectionOptionFormData>({
    resolver: zodResolver(addSelectionOptionSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  const submit = handleSubmit((data) => {
    onAddOption(field, data.label);
    reset(DEFAULT_VALUES);
  });

  const canAddOption = hasLink(field, 'x-add-option');

  return (
    <Stack gap="xs" pl="md">
      <Group gap="xs" wrap="wrap">
        {(field.options ?? []).map((option) => (
          <OptionBadge key={option.id} option={option} onRetireOption={onRetireOption} />
        ))}
      </Group>
      {canAddOption && (
        <form onSubmit={submit}>
          <Group gap="xs">
            <TextInput
              size="xs"
              placeholder="New option"
              {...register('label')}
              error={errors.label?.message}
              data-testid={`one-pager-add-option-input-${field.id}`}
            />
            <ActionIcon
              type="submit"
              size="sm"
              variant="light"
              disabled={!isValid}
              aria-label={`Add option to ${field.name}`}
              data-testid={`one-pager-add-option-submit-${field.id}`}
            >
              +
            </ActionIcon>
          </Group>
        </form>
      )}
    </Stack>
  );
}
