import { Badge, Box, Group, NumberInput, Select, Stack, Text, TextInput } from '@mantine/core';
import { Controller, type Control, type FieldError } from 'react-hook-form';
import {
  emptyFactValue,
  type ContactPersonFactValue,
  type LinkFactValue,
  type OnePagerFactsFormValues,
} from '../../../lib/schemas/onePagerFacts';
import { selectionItems } from '../factFields';
import type { CustomField, OnePagerFieldType } from '../types';

interface ControllerRenderField {
  value: unknown;
  onChange: (value: unknown) => void;
  onBlur: () => void;
}

interface FactInputProps {
  field: CustomField;
  rhf: ControllerRenderField;
  error?: FieldError;
}

function nestedError(error: FieldError | undefined, key: string): string | undefined {
  if (!error) return undefined;
  const nested = (error as unknown as Record<string, { message?: string } | undefined>)[key];
  return nested?.message;
}

function TextFactInput({ field, rhf, error, type }: FactInputProps & { type?: string }) {
  return (
    <TextInput
      type={type}
      aria-label={field.name}
      value={rhf.value as string}
      onChange={(event) => rhf.onChange(event.currentTarget.value)}
      onBlur={rhf.onBlur}
      error={error?.message}
    />
  );
}

function DateFactInput(props: FactInputProps) {
  return <TextFactInput {...props} type="date" />;
}

function NumberFactInput({ field, rhf, error }: FactInputProps) {
  return (
    <NumberInput
      aria-label={field.name}
      value={rhf.value as number | ''}
      onChange={rhf.onChange}
      onBlur={rhf.onBlur}
      min={field.min}
      max={field.max}
      error={error?.message}
    />
  );
}

function SelectionFactInput({ field, rhf, error }: FactInputProps) {
  return (
    <Select
      aria-label={field.name}
      data={selectionItems(field, rhf.value as string)}
      value={(rhf.value as string) || null}
      onChange={(value) => rhf.onChange(value ?? '')}
      onBlur={rhf.onBlur}
      clearable
      error={error?.message}
    />
  );
}

function LinkFactInput({ field, rhf, error }: FactInputProps) {
  const link = rhf.value as LinkFactValue;
  return (
    <Group gap="sm" align="flex-start" wrap="nowrap">
      <Box flex={1}>
        <TextInput
          aria-label={`${field.name} label`}
          placeholder="Label"
          value={link.label}
          onChange={(event) => rhf.onChange({ ...link, label: event.currentTarget.value })}
          onBlur={rhf.onBlur}
          error={nestedError(error, 'label')}
        />
      </Box>
      <Box flex={2}>
        <TextInput
          aria-label={`${field.name} URL`}
          placeholder="https://…"
          value={link.url}
          onChange={(event) => rhf.onChange({ ...link, url: event.currentTarget.value })}
          onBlur={rhf.onBlur}
          error={nestedError(error, 'url')}
        />
      </Box>
    </Group>
  );
}

function ContactPersonFactInput({ field, rhf, error }: FactInputProps) {
  const contact = rhf.value as ContactPersonFactValue;
  const partInput = (part: keyof ContactPersonFactValue, placeholder: string) => (
    <TextInput
      aria-label={`${field.name} ${part}`}
      placeholder={placeholder}
      value={contact[part]}
      onChange={(event) => rhf.onChange({ ...contact, [part]: event.currentTarget.value })}
      onBlur={rhf.onBlur}
      error={nestedError(error, part)}
    />
  );
  return (
    <Stack gap="xs">
      {partInput('name', 'Name')}
      {partInput('email', 'Email')}
      {partInput('company', 'Company')}
    </Stack>
  );
}

const factInputsByType: Record<OnePagerFieldType, (props: FactInputProps) => React.ReactElement> = {
  text: TextFactInput,
  number: NumberFactInput,
  date: DateFactInput,
  link: LinkFactInput,
  selection: SelectionFactInput,
  'contact-person': ContactPersonFactInput,
};

interface FactFieldInputProps {
  field: CustomField;
  control: Control<OnePagerFactsFormValues>;
  showRequiredHint: boolean;
}

export function FactFieldInput({ field, control, showRequiredHint }: FactFieldInputProps) {
  const InputControl = factInputsByType[field.type];
  return (
    <Stack gap={4}>
      <Group gap="xs">
        <Text size="sm" fw={500}>
          {field.name}
        </Text>
        {showRequiredHint && (
          <Badge size="xs" color="yellow" variant="light">
            Required
          </Badge>
        )}
      </Group>
      {field.helpText && (
        <Text size="xs" c="dimmed">
          {field.helpText}
        </Text>
      )}
      <Controller
        name={field.id}
        control={control}
        defaultValue={emptyFactValue(field.type)}
        render={({ field: rhf, fieldState }) => <InputControl field={field} rhf={rhf} error={fieldState.error} />}
      />
    </Stack>
  );
}
