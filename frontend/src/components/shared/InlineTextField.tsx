import { Group, Text, Textarea, TextInput, Title } from '@mantine/core';
import React, { useState } from 'react';
import type { ZodType } from 'zod';
import { DetailField } from './DetailField';
import { InlineEditButton, InlineEmptyPrompt, InlineFieldActions, useInlineCommit } from './InlineFieldChrome';
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
  const parse = (value: string): { ok: true; value: string } | { ok: false; message: string } => {
    const parsed = schema.safeParse(value);
    return parsed.success
      ? { ok: true, value: parsed.data }
      : { ok: false, message: parsed.error.issues[0]?.message ?? 'Invalid value' };
  };
  const { error, saving, commit } = useInlineCommit({ initialValue, parse, onSave, onDone });

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      onDone();
    } else if (isConfirmKey(event, multiline)) {
      event.preventDefault();
      void commit(draft);
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
      <InlineFieldActions testId={testId} saving={saving} onConfirm={() => void commit(draft)} onCancel={onDone} />
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
      {canEdit && <InlineEditButton testId={testId} editLabel={editLabel} onEdit={onEdit} />}
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
    <InlineEmptyPrompt testId={testId} prompt={emptyPrompt ?? editLabel} onEdit={() => setEditing(true)} />
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
