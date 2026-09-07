import { ActionIcon, Button } from '@mantine/core';
import { IconCheck, IconPencil, IconX } from '@tabler/icons-react';
import { useState } from 'react';

interface UseInlineCommitOptions<T> {
  initialValue: T;
  parse: (draft: T) => { ok: true; value: T } | { ok: false; message: string };
  onSave: (value: T) => Promise<unknown>;
  onDone: () => void;
}

export function useInlineCommit<T>({ initialValue, parse, onSave, onDone }: UseInlineCommitOptions<T>) {
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const commit = async (draft: T) => {
    const parsed = parse(draft);
    if (!parsed.ok) {
      setError(parsed.message);
      return;
    }
    if (parsed.value === initialValue) {
      onDone();
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await onSave(parsed.value);
      onDone();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
      setSaving(false);
    }
  };

  return { error, saving, commit };
}

interface InlineFieldActionsProps {
  testId: string;
  saving: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function InlineFieldActions({ testId, saving, onConfirm, onCancel }: InlineFieldActionsProps) {
  return (
    <>
      <ActionIcon
        variant="filled"
        aria-label="Save"
        loading={saving}
        onClick={onConfirm}
        data-testid={`${testId}-save`}
      >
        <IconCheck size={16} stroke={1.75} />
      </ActionIcon>
      <ActionIcon
        variant="default"
        aria-label="Cancel"
        disabled={saving}
        onClick={onCancel}
        data-testid={`${testId}-cancel`}
      >
        <IconX size={16} stroke={1.75} />
      </ActionIcon>
    </>
  );
}

interface InlineEditButtonProps {
  testId: string;
  editLabel: string;
  onEdit: () => void;
}

export function InlineEditButton({ testId, editLabel, onEdit }: InlineEditButtonProps) {
  return (
    <ActionIcon variant="subtle" aria-label={editLabel} onClick={onEdit} data-testid={`${testId}-edit`}>
      <IconPencil size={16} stroke={1.75} />
    </ActionIcon>
  );
}

interface InlineEmptyPromptProps {
  testId: string;
  prompt: string;
  onEdit: () => void;
}

export function InlineEmptyPrompt({ testId, prompt, onEdit }: InlineEmptyPromptProps) {
  return (
    <Button variant="subtle" size="compact-sm" onClick={onEdit} data-testid={`${testId}-edit`}>
      {prompt}
    </Button>
  );
}
