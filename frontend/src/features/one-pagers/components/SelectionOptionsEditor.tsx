import { zodResolver } from '@hookform/resolvers/zod';
import { ActionIcon, Badge, Group, Stack, TextInput } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { type AddSelectionOptionFormData, addSelectionOptionSchema } from '../../../lib/schemas/onePagerConfiguration';
import type { SubjectAttribute, SubjectAttributeOption } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';

interface SelectionOptionsEditorProps {
  attribute: SubjectAttribute;
  onAddOption: (attribute: SubjectAttribute, label: string) => void;
  onRetireOption: (option: SubjectAttributeOption) => void;
}

const DEFAULT_VALUES: AddSelectionOptionFormData = { label: '' };

function OptionBadge({
  option,
  onRetireOption,
}: {
  option: SubjectAttributeOption;
  onRetireOption: (option: SubjectAttributeOption) => void;
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

export function SelectionOptionsEditor({ attribute, onAddOption, onRetireOption }: SelectionOptionsEditorProps) {
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
    onAddOption(attribute, data.label);
    reset(DEFAULT_VALUES);
  });

  const canAddOption = hasLink(attribute, 'x-add-option');

  return (
    <Stack gap="xs" pl="md">
      <Group gap="xs" wrap="wrap">
        {(attribute.options ?? []).map((option) => (
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
              data-testid={`one-pager-add-option-input-${attribute.id}`}
            />
            <ActionIcon
              type="submit"
              size="sm"
              variant="light"
              disabled={!isValid}
              aria-label={`Add option to ${attribute.name}`}
              data-testid={`one-pager-add-option-submit-${attribute.id}`}
            >
              +
            </ActionIcon>
          </Group>
        </form>
      )}
    </Stack>
  );
}
