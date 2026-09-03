import { Group, Select, Text } from '@mantine/core';
import React, { useState } from 'react';
import { DetailField } from './DetailField';
import { InlineEditButton, InlineEmptyPrompt, InlineFieldActions, useInlineCommit } from './InlineFieldChrome';
import classes from './InlineTextField.module.css';

export interface InlineSelectOption {
  value: string;
  label: string;
}

export interface InlineSelectFieldProps {
  label: string;
  value: string;
  options: InlineSelectOption[];
  canEdit: boolean;
  onSave: (value: string) => Promise<unknown>;
  editLabel: string;
  testId: string;
  emptyPrompt?: string;
  renderValue?: (value: string, label: string) => React.ReactNode;
  searchable?: boolean;
}

interface EditorProps {
  initialValue: string;
  options: InlineSelectOption[];
  searchable: boolean;
  testId: string;
  onSave: (value: string) => Promise<unknown>;
  onDone: () => void;
}

function parseSelection(draft: string): { ok: true; value: string } | { ok: false; message: string } {
  return draft ? { ok: true, value: draft } : { ok: false, message: 'A value is required' };
}

function Editor({ initialValue, options, searchable, testId, onSave, onDone }: EditorProps) {
  const [draft, setDraft] = useState(initialValue);
  const { error, saving, commit } = useInlineCommit({ initialValue, parse: parseSelection, onSave, onDone });

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      onDone();
    }
  };

  return (
    <Group gap="xs" align="flex-start" wrap="nowrap">
      <div className={classes.grow}>
        <Select
          value={draft || null}
          onChange={(value) => setDraft(value ?? '')}
          onKeyDown={handleKeyDown}
          data={options}
          searchable={searchable}
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

export function InlineSelectField({
  label,
  value,
  options,
  canEdit,
  onSave,
  editLabel,
  testId,
  emptyPrompt,
  renderValue,
  searchable = false,
}: InlineSelectFieldProps) {
  const [editing, setEditing] = useState(false);

  if (!value && !canEdit) return null;

  const displayLabel = options.find((option) => option.value === value)?.label ?? value;
  const startEditing = () => setEditing(true);

  const body = editing ? (
    <Editor
      initialValue={value}
      options={options}
      searchable={searchable}
      testId={testId}
      onSave={onSave}
      onDone={() => setEditing(false)}
    />
  ) : !value ? (
    <InlineEmptyPrompt testId={testId} prompt={emptyPrompt ?? editLabel} onEdit={startEditing} />
  ) : (
    <Group gap="xs" align="flex-start" wrap="nowrap" justify="space-between">
      <div data-testid={`${testId}-value`}>
        {renderValue ? renderValue(value, displayLabel) : <Text size="sm">{displayLabel}</Text>}
      </div>
      {canEdit && <InlineEditButton testId={testId} editLabel={editLabel} onEdit={startEditing} />}
    </Group>
  );

  return <DetailField label={label}>{body}</DetailField>;
}
