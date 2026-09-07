import { Group, Text, TextInput } from '@mantine/core';
import type React from 'react';
import { useState } from 'react';
import { DetailField } from './DetailField';
import { InlineEditButton, InlineEmptyPrompt, InlineFieldActions, useInlineCommit } from './InlineFieldChrome';
import classes from './InlineTextField.module.css';

export interface InlineDateFieldProps {
  label: string;
  value: string;
  canEdit: boolean;
  onSave: (value: string) => Promise<unknown>;
  editLabel: string;
  emptyPrompt: string;
  testId: string;
}

interface EditorProps {
  initialValue: string;
  testId: string;
  onSave: (value: string) => Promise<unknown>;
  onDone: () => void;
}

function acceptDate(draft: string): { ok: true; value: string } {
  return { ok: true, value: draft };
}

function Editor({ initialValue, testId, onSave, onDone }: EditorProps) {
  const [draft, setDraft] = useState(initialValue);
  const { error, saving, commit } = useInlineCommit({ initialValue, parse: acceptDate, onSave, onDone });

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      onDone();
    } else if (event.key === 'Enter') {
      event.preventDefault();
      void commit(draft);
    }
  };

  return (
    <Group gap="xs" align="flex-start" wrap="nowrap">
      <div className={classes.grow}>
        <TextInput
          type="date"
          value={draft}
          onChange={(event) => setDraft(event.currentTarget.value)}
          onKeyDown={handleKeyDown}
          error={error}
          disabled={saving}
          size="sm"
          autoFocus
          data-testid={`${testId}-input`}
        />
      </div>
      <InlineFieldActions testId={testId} saving={saving} onConfirm={() => void commit(draft)} onCancel={onDone} />
    </Group>
  );
}

function formatDate(value: string): string {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleDateString();
}

export function InlineDateField({
  label,
  value,
  canEdit,
  onSave,
  editLabel,
  emptyPrompt,
  testId,
}: InlineDateFieldProps) {
  const [editing, setEditing] = useState(false);

  if (!value && !canEdit) return null;

  const startEditing = () => setEditing(true);

  const body = editing ? (
    <Editor initialValue={value} testId={testId} onSave={onSave} onDone={() => setEditing(false)} />
  ) : !value ? (
    <InlineEmptyPrompt testId={testId} prompt={emptyPrompt} onEdit={startEditing} />
  ) : (
    <Group gap="xs" align="flex-start" wrap="nowrap" justify="space-between">
      <Text size="sm" data-testid={`${testId}-value`}>
        {formatDate(value)}
      </Text>
      {canEdit && <InlineEditButton testId={testId} editLabel={editLabel} onEdit={startEditing} />}
    </Group>
  );

  return <DetailField label={label}>{body}</DetailField>;
}
