import { ActionIcon, Button, Group, Text, Textarea, TextInput, Title } from '@mantine/core';
import { IconCheck, IconPencil, IconX } from '@tabler/icons-react';
import React, { useState } from 'react';
import type { ZodType } from 'zod';
import { DetailField } from './DetailField';
import classes from './InlineTextField.module.css';

export interface InlineTextFieldProps {
  label?: string;
  value: string;
  canEdit: boolean;
  schema: ZodType<string, string>;
  onSave: (value: string) => Promise<unknown>;
  editLabel: string;
  testId: string;
  emptyPrompt?: string;
  multiline?: boolean;
}

interface EditorProps {
  initialValue: string;
  multiline: boolean;
  schema: ZodType<string, string>;
  testId: string;
  onSave: (value: string) => Promise<unknown>;
  onDone: () => void;
}

function isConfirmKey(event: React.KeyboardEvent, multiline: boolean): boolean {
  if (event.key !== 'Enter') return false;
  return multiline ? event.ctrlKey || event.metaKey : true;
}

function Editor({ initialValue, multiline, schema, testId, onSave, onDone }: EditorProps) {
  const [draft, setDraft] = useState(initialValue);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const confirm = async () => {
    const parsed = schema.safeParse(draft);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'Invalid value');
      return;
    }
    if (parsed.data === initialValue) {
      onDone();
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await onSave(parsed.data);
      onDone();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
      setSaving(false);
    }
  };

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      onDone();
    } else if (isConfirmKey(event, multiline)) {
      event.preventDefault();
      void confirm();
    }
  };

  const inputProps = {
    value: draft,
    onChange: (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => setDraft(event.currentTarget.value),
    onKeyDown: handleKeyDown,
    error,
    disabled: saving,
    size: 'sm' as const,
    autoFocus: true,
    'data-testid': `${testId}-input`,
  };

  return (
    <Group gap="xs" align="flex-start" wrap="nowrap">
      <div className={classes.grow}>
        {multiline ? <Textarea {...inputProps} autosize minRows={2} /> : <TextInput {...inputProps} />}
      </div>
      <ActionIcon
        variant="filled"
        aria-label="Save"
        loading={saving}
        onClick={() => void confirm()}
        data-testid={`${testId}-save`}
      >
        <IconCheck size={16} stroke={1.75} />
      </ActionIcon>
      <ActionIcon variant="default" aria-label="Cancel" disabled={saving} onClick={onDone} data-testid={`${testId}-cancel`}>
        <IconX size={16} stroke={1.75} />
      </ActionIcon>
    </Group>
  );
}

interface ReadViewProps {
  value: string;
  heading: boolean;
  canEdit: boolean;
  editLabel: string;
  testId: string;
  onEdit: () => void;
}

function ReadView({ value, heading, canEdit, editLabel, testId, onEdit }: ReadViewProps) {
  return (
    <Group gap="xs" align="flex-start" wrap="nowrap" justify="space-between">
      {heading ? (
        <Title order={4} data-testid={`${testId}-value`}>
          {value}
        </Title>
      ) : (
        <Text size="sm" className={classes.preWrap} data-testid={`${testId}-value`}>
          {value}
        </Text>
      )}
      {canEdit && (
        <ActionIcon variant="subtle" aria-label={editLabel} onClick={onEdit} data-testid={`${testId}-edit`}>
          <IconPencil size={16} stroke={1.75} />
        </ActionIcon>
      )}
    </Group>
  );
}

export function InlineTextField({
  label,
  value,
  canEdit,
  schema,
  onSave,
  editLabel,
  testId,
  emptyPrompt,
  multiline = false,
}: InlineTextFieldProps) {
  const [editing, setEditing] = useState(false);

  if (!value && !canEdit) return null;

  const body = editing ? (
    <Editor
      initialValue={value}
      multiline={multiline}
      schema={schema}
      testId={testId}
      onSave={onSave}
      onDone={() => setEditing(false)}
    />
  ) : !value ? (
    <Button variant="subtle" size="compact-sm" onClick={() => setEditing(true)} data-testid={`${testId}-edit`}>
      {emptyPrompt ?? editLabel}
    </Button>
  ) : (
    <ReadView
      value={value}
      heading={label === undefined}
      canEdit={canEdit}
      editLabel={editLabel}
      testId={testId}
      onEdit={() => setEditing(true)}
    />
  );

  return label === undefined ? body : <DetailField label={label}>{body}</DetailField>;
}
