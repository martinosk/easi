import { Badge, Box, Button, Center, Divider, Group, Paper, Stack, Text, Title } from '@mantine/core';
import { useMemo, useState } from 'react';
import type { Control } from 'react-hook-form';
import { useParams } from 'react-router-dom';
import { ApiError } from '../../../api/types';
import { ShareIcon } from '../../../components/shared/ContextMenu/icons';
import { DetailField } from '../../../components/shared/DetailField';
import { LoadingFallback } from '../../../components/shared/LoadingFallback';
import type { OnePagerFactsFormValues } from '../../../lib/schemas/onePagerFacts';
import { hasLink } from '../../../utils/hateoas';
import { copyToClipboard, generateOnePagerShareUrl } from '../../../utils/clipboard';
import { BuiltInValueDisplay } from '../components/BuiltInValueDisplay';
import { FactFieldInput } from '../components/FactFieldInput';
import { FactValueDisplay } from '../components/FactValueDisplay';
import { activeCustomFieldsInOrder, customFieldViewDisplayProps } from '../factFields';
import { useOnePager } from '../hooks/useOnePager';
import { useOnePagerConfiguration } from '../hooks/useOnePagerConfiguration';
import { useOnePagerFacts } from '../hooks/useOnePagerFacts';
import { useOnePagerFactsForm } from '../hooks/useOnePagerFactsForm';
import { subjectTypeLabel } from '../subjectTypes';
import {
  ONE_PAGER_SUBJECT_TYPES,
  type CustomField,
  type OnePagerCompleteness,
  type OnePagerFacts,
  type OnePagerSubjectType,
  type OnePagerView,
  type OnePagerViewField,
} from '../types';

const NOT_FOUND_MESSAGE = 'One-pager not found';

type OnePagerMode = 'read' | 'edit';

function isOnePagerSubjectType(value: string | undefined): value is OnePagerSubjectType {
  return !!value && (ONE_PAGER_SUBJECT_TYPES as readonly string[]).includes(value);
}

function OnePagerStatus({ title, message }: { title: string; message?: string }) {
  return (
    <Center mih="60vh" p="xl">
      <Stack align="center" gap="xs">
        <Title order={3}>{title}</Title>
        {message && (
          <Text c="dimmed" size="sm">
            {message}
          </Text>
        )}
      </Stack>
    </Center>
  );
}

function fieldKey(field: OnePagerViewField): string {
  return field.kind === 'builtIn' ? `builtIn-${field.builtIn.id}` : `custom-${field.custom.fieldId}`;
}

function MissingRequiredValue({ fieldId }: { fieldId: string }) {
  return (
    <Text size="sm" c="orange.7" data-testid={`one-pager-missing-required-${fieldId}`}>
      missing — required
    </Text>
  );
}

interface OnePagerFieldRowProps {
  field: OnePagerViewField;
  missingFieldIds: Set<string>;
  editableFieldsById: Map<string, CustomField>;
  control: Control<OnePagerFactsFormValues>;
  requiredHint: (field: CustomField) => boolean;
}

function OnePagerFieldRow({ field, missingFieldIds, editableFieldsById, control, requiredHint }: OnePagerFieldRowProps) {
  if (field.kind === 'builtIn') {
    return (
      <DetailField label={field.builtIn.label}>
        {missingFieldIds.has(field.builtIn.id) ? (
          <MissingRequiredValue fieldId={field.builtIn.id} />
        ) : (
          <BuiltInValueDisplay value={field.builtIn.value} />
        )}
      </DetailField>
    );
  }

  const editableField = editableFieldsById.get(field.custom.fieldId);
  if (editableField) {
    return <FactFieldInput field={editableField} control={control} showRequiredHint={requiredHint(editableField)} />;
  }

  const { field: customField, fieldValue } = customFieldViewDisplayProps(field.custom);
  if (missingFieldIds.has(field.custom.fieldId)) {
    return (
      <DetailField label={field.custom.name}>
        <MissingRequiredValue fieldId={field.custom.fieldId} />
      </DetailField>
    );
  }
  return (
    <DetailField label={field.custom.name}>
      <FactValueDisplay field={customField} fieldValue={fieldValue} />
    </DetailField>
  );
}

function CompletenessSummary({ completeness }: { completeness: OnePagerCompleteness }) {
  if (completeness.requiredCount === 0) return null;
  const complete = completeness.filledCount === completeness.requiredCount;
  return (
    <Badge
      data-testid="one-pager-completeness-summary"
      variant="light"
      color={complete ? 'teal' : 'orange'}
      radius="sm"
    >
      {`${completeness.filledCount} of ${completeness.requiredCount} required fields filled`}
    </Badge>
  );
}

interface OnePagerSheetActionsProps {
  mode: OnePagerMode;
  canEdit: boolean;
  isDirty: boolean;
  isPending: boolean;
  onShare: () => void;
  onEdit: () => void;
  onSave: () => void;
  onCancel: () => void;
}

function OnePagerSheetActions({ mode, canEdit, isDirty, isPending, onShare, onEdit, onSave, onCancel }: OnePagerSheetActionsProps) {
  if (mode === 'edit') {
    return (
      <Group gap="sm">
        <Button variant="default" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button size="sm" onClick={onSave} disabled={!isDirty} loading={isPending}>
          Save
        </Button>
      </Group>
    );
  }
  return (
    <Group gap="sm">
      {canEdit && (
        <Button variant="default" size="sm" onClick={onEdit}>
          Edit
        </Button>
      )}
      <Button variant="light" leftSection={<ShareIcon />} onClick={onShare}>
        Share (copy URL)
      </Button>
    </Group>
  );
}

const EMPTY_EDITABLE_FIELDS = new Map<string, CustomField>();

function useEditableFacts(view: OnePagerView, editing: boolean) {
  const configurationQuery = useOnePagerConfiguration(view.subjectType, editing);
  const factsQuery = useOnePagerFacts(view.subjectType, view.subjectId, editing);

  const editableFields = useMemo(
    () => (configurationQuery.data ? activeCustomFieldsInOrder(configurationQuery.data) : []),
    [configurationQuery.data],
  );
  const editableFieldsById = useMemo(
    () => (editing ? new Map(editableFields.map((field) => [field.id, field])) : EMPTY_EDITABLE_FIELDS),
    [editing, editableFields],
  );
  const facts = useMemo<OnePagerFacts>(
    () => factsQuery.data ?? { subjectType: view.subjectType, subjectId: view.subjectId, values: [], _links: {} },
    [factsQuery.data, view.subjectType, view.subjectId],
  );

  return { editableFields, editableFieldsById, facts };
}

function OnePagerSheet({ view }: { view: OnePagerView }) {
  const [mode, setMode] = useState<OnePagerMode>('read');
  const editing = mode === 'edit';
  const { editableFields, editableFieldsById, facts } = useEditableFacts(view, editing);
  const editForm = useOnePagerFactsForm(editableFields, facts, () => setMode('read'));

  const missingFieldIds = useMemo(
    () => new Set(view.completeness.missingFields.map((field) => field.fieldId)),
    [view.completeness.missingFields],
  );

  const handleShare = () => {
    copyToClipboard(generateOnePagerShareUrl(view.subjectType, view.subjectId));
  };

  const handleCancel = () => {
    editForm.cancel();
    setMode('read');
  };

  return (
    <Box p="xl">
      <Paper shadow="sm" radius="lg" p="xl">
        <Stack gap="lg">
          <Group justify="space-between" align="flex-start" wrap="nowrap">
            <Stack gap="xs">
              <Title order={2}>{view.subjectName}</Title>
              <Group gap="xs">
                <Badge variant="light" color="gray" radius="sm">
                  {subjectTypeLabel(view.subjectType)}
                </Badge>
                <CompletenessSummary completeness={view.completeness} />
              </Group>
            </Stack>
            <OnePagerSheetActions
              mode={mode}
              canEdit={hasLink(view, 'x-record')}
              isDirty={editForm.isDirty}
              isPending={editForm.isPending}
              onShare={handleShare}
              onEdit={() => setMode('edit')}
              onSave={editForm.submit}
              onCancel={handleCancel}
            />
          </Group>

          <Divider />

          <Stack gap="sm">
            {view.fields.map((field) => (
              <OnePagerFieldRow
                key={fieldKey(field)}
                field={field}
                missingFieldIds={missingFieldIds}
                editableFieldsById={editableFieldsById}
                control={editForm.control}
                requiredHint={editForm.requiredHint}
              />
            ))}
          </Stack>
        </Stack>
      </Paper>
    </Box>
  );
}

export function OnePagerPage() {
  const { subjectType, subjectId } = useParams<{ subjectType: string; subjectId: string }>();
  const validSubjectType = isOnePagerSubjectType(subjectType) ? subjectType : undefined;
  const onePagerQuery = useOnePager(validSubjectType, subjectId);

  if (!validSubjectType || !subjectId) {
    return <OnePagerStatus title={NOT_FOUND_MESSAGE} />;
  }

  if (onePagerQuery.isLoading) {
    return <LoadingFallback message="Loading one-pager..." />;
  }

  if (onePagerQuery.error) {
    if (onePagerQuery.error instanceof ApiError && onePagerQuery.error.statusCode === 404) {
      return <OnePagerStatus title={NOT_FOUND_MESSAGE} />;
    }
    return <OnePagerStatus title="Failed to load one-pager" message={onePagerQuery.error.message} />;
  }

  if (!onePagerQuery.data) {
    return <OnePagerStatus title={NOT_FOUND_MESSAGE} />;
  }

  return <OnePagerSheet view={onePagerQuery.data} />;
}
