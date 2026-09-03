import { Badge, Group, Stack, Text } from '@mantine/core';
import { useState } from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineEditButton, InlineFieldActions, useInlineCommit } from '../../../components/shared/InlineFieldChrome';
import { MaturitySlider } from '../../../components/shared/MaturitySlider';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { getDefaultSections, getSectionForValue } from '../../../utils/maturity';

const TEST_ID = 'capability-maturity';

type MaturityColor = 'red' | 'orange' | 'green' | 'blue' | 'gray';

const SECTION_COLORS: Record<string, MaturityColor> = {
  genesis: 'red',
  'custom build': 'orange',
  'custom built': 'orange',
  product: 'green',
  commodity: 'blue',
};

function maturityBadgeColor(sectionName: string): MaturityColor {
  return SECTION_COLORS[sectionName.toLowerCase()] ?? 'gray';
}

export interface InlineMaturityFieldProps {
  value: number;
  canEdit: boolean;
  onSave: (value: number) => Promise<unknown>;
}

interface EditorProps {
  initialValue: number;
  onSave: (value: number) => Promise<unknown>;
  onDone: () => void;
}

function acceptNumber(draft: number): { ok: true; value: number } {
  return { ok: true, value: draft };
}

function Editor({ initialValue, onSave, onDone }: EditorProps) {
  const [draft, setDraft] = useState(initialValue);
  const { error, saving, commit } = useInlineCommit({ initialValue, parse: acceptNumber, onSave, onDone });

  return (
    <Stack gap="xs">
      <MaturitySlider value={draft} onChange={setDraft} disabled={saving} />
      {error && (
        <Text size="xs" c="red">
          {error}
        </Text>
      )}
      <Group gap="xs" justify="flex-end">
        <InlineFieldActions testId={TEST_ID} saving={saving} onConfirm={() => void commit(draft)} onCancel={onDone} />
      </Group>
    </Stack>
  );
}

export function InlineMaturityField({ value, canEdit, onSave }: InlineMaturityFieldProps) {
  const [editing, setEditing] = useState(false);
  const { data: maturityScale } = useMaturityScale();
  const sections = maturityScale?.sections?.length ? maturityScale.sections : getDefaultSections();
  const sectionName = getSectionForValue(value, sections)?.name ?? 'Unknown';

  return (
    <DetailField label="Maturity">
      {editing ? (
        <Editor initialValue={value} onSave={onSave} onDone={() => setEditing(false)} />
      ) : (
        <Group gap="xs" align="flex-start" wrap="nowrap" justify="space-between">
          <Badge color={maturityBadgeColor(sectionName)} variant="filled" size="md" data-testid={`${TEST_ID}-value`}>
            {sectionName} ({value})
          </Badge>
          {canEdit && <InlineEditButton testId={TEST_ID} editLabel="Edit maturity" onEdit={() => setEditing(true)} />}
        </Group>
      )}
    </DetailField>
  );
}
