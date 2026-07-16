import { Button, Group, Loader, Modal, Stack, Text } from '@mantine/core';
import { useImpactPreview } from '../hooks/useImpactPreview';
import { pluralSubjectTypeLabel } from '../subjectTypes';
import type { ImpactPreviewFieldKind, OnePagerConfiguration } from '../types';

interface ImpactPreviewDialogProps {
  configuration: OnePagerConfiguration;
  fieldName: string;
  fieldId?: string;
  fieldKind?: ImpactPreviewFieldKind;
  isConfirming: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

function previewMessage(fieldName: string, pluralLabel: string, count: number | undefined, isError: boolean): string {
  if (isError || count === undefined) {
    return `Making ${fieldName} required — the number of affected ${pluralLabel} could not be determined.`;
  }
  return `Making ${fieldName} required will mark ${count} ${pluralLabel} incomplete`;
}

export function ImpactPreviewDialog({
  configuration,
  fieldName,
  fieldId,
  fieldKind = 'custom',
  isConfirming,
  onConfirm,
  onCancel,
}: ImpactPreviewDialogProps) {
  const { data, isLoading, isError } = useImpactPreview(configuration, fieldId, true, fieldKind);
  const pluralLabel = pluralSubjectTypeLabel(configuration.subjectType);

  return (
    <Modal
      opened
      onClose={onCancel}
      title="Confirm required field"
      centered
      data-testid="one-pager-impact-preview-dialog"
    >
      <Stack gap="md">
        {isLoading ? (
          <Group gap="xs" data-testid="one-pager-impact-preview-loading">
            <Loader size="xs" />
            <Text size="sm" c="dimmed">
              Checking impact…
            </Text>
          </Group>
        ) : (
          <Text data-testid="one-pager-impact-preview-message">
            {previewMessage(fieldName, pluralLabel, data?.affectedSubjectCount, isError)}
          </Text>
        )}
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={isConfirming}>
            Cancel
          </Button>
          <Button onClick={onConfirm} loading={isConfirming} data-testid="one-pager-impact-preview-confirm">
            Confirm
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
