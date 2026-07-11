import { zodResolver } from '@hookform/resolvers/zod';
import { ActionIcon, Badge, Button, Checkbox, Group, Select, Stack, Text, Textarea, TextInput } from '@mantine/core';
import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { type DefineCustomFieldFormData, defineCustomFieldSchema } from '../../../lib/schemas/onePagerConfiguration';
import { ONE_PAGER_FIELD_TYPE_OPTIONS } from '../fieldTypes';

interface AddCustomFieldFormProps {
  isSaving: boolean;
  onSubmit: (data: DefineCustomFieldFormData) => void;
}

const DEFAULT_VALUES: DefineCustomFieldFormData = {
  name: '',
  fieldType: 'text',
  required: false,
  helpText: '',
  options: [],
};

interface OptionEntry {
  id: string;
  label: string;
}

function OptionsEditor({ onChange, error }: { onChange: (labels: string[]) => void; error?: string }) {
  const [entries, setEntries] = useState<OptionEntry[]>([]);
  const [optionInput, setOptionInput] = useState('');

  const add = () => {
    const trimmed = optionInput.trim();
    if (!trimmed) return;
    const next = [...entries, { id: crypto.randomUUID(), label: trimmed }];
    setEntries(next);
    onChange(next.map((entry) => entry.label));
    setOptionInput('');
  };

  const remove = (id: string) => {
    const next = entries.filter((entry) => entry.id !== id);
    setEntries(next);
    onChange(next.map((entry) => entry.label));
  };

  return (
    <Stack gap="xs">
      <Group gap="xs">
        <TextInput
          placeholder="Option label"
          value={optionInput}
          onChange={(e) => setOptionInput(e.currentTarget.value)}
          data-testid="one-pager-new-option-input"
        />
        <Button variant="default" onClick={add} data-testid="one-pager-new-option-add">
          Add option
        </Button>
      </Group>
      <Group gap="xs" wrap="wrap">
        {entries.map((entry) => (
          <Badge
            key={entry.id}
            variant="light"
            rightSection={
              <ActionIcon
                size="xs"
                variant="transparent"
                color="gray"
                aria-label={`Remove ${entry.label}`}
                onClick={() => remove(entry.id)}
              >
                ×
              </ActionIcon>
            }
          >
            {entry.label}
          </Badge>
        ))}
      </Group>
      {error && (
        <Text c="red" size="xs">
          {error}
        </Text>
      )}
    </Stack>
  );
}

export function AddCustomFieldForm({ isSaving, onSubmit }: AddCustomFieldFormProps) {
  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isValid },
  } = useForm<DefineCustomFieldFormData>({
    resolver: zodResolver(defineCustomFieldSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  const fieldType = watch('fieldType');

  useEffect(() => {
    if (fieldType !== 'selection') setValue('options', [], { shouldValidate: true });
  }, [fieldType, setValue]);

  const submit = handleSubmit(onSubmit);

  return (
    <form onSubmit={submit} data-testid="one-pager-add-field-form">
      <Stack gap="md">
        <TextInput
          label="Field name"
          {...register('name')}
          required
          withAsterisk
          error={errors.name?.message}
          data-testid="one-pager-new-field-name"
        />
        <Controller
          name="fieldType"
          control={control}
          render={({ field }) => (
            <Select
              label="Field type"
              data={ONE_PAGER_FIELD_TYPE_OPTIONS}
              data-testid="one-pager-new-field-type"
              {...field}
            />
          )}
        />
        <Textarea
          label="Help text"
          {...register('helpText')}
          rows={2}
          error={errors.helpText?.message}
          data-testid="one-pager-new-field-help"
        />
        <Controller
          name="required"
          control={control}
          render={({ field }) => (
            <Checkbox
              label="Required"
              checked={field.value}
              onChange={(e) => field.onChange(e.currentTarget.checked)}
              data-testid="one-pager-new-field-required"
            />
          )}
        />
        {fieldType === 'selection' && (
          <OptionsEditor
            onChange={(labels) => setValue('options', labels, { shouldValidate: true })}
            error={errors.options?.message}
          />
        )}
        <Group justify="flex-end">
          <Button type="submit" loading={isSaving} disabled={!isValid} data-testid="one-pager-new-field-submit">
            Add field
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
