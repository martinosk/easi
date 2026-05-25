import { Alert, Badge, Box, Button, Group, Loader, Modal, Stack, Text, Title } from '@mantine/core';
import { useState } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { useStandardApplicationForEnterpriseCapability } from '../hooks/useStandardApplication';
import type { ECStandardApplicationResponse, StandardApplication } from '../types';
import { SetStandardApplicationForm } from './SetStandardApplicationForm';
import { StandardApplicationHistoryDialog } from './StandardApplicationHistoryDialog';

interface StandardApplicationPanelProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

export function StandardApplicationPanel({ enterpriseCapabilityId }: StandardApplicationPanelProps) {
  const { data, isLoading, error } = useStandardApplicationForEnterpriseCapability(enterpriseCapabilityId);
  const [isEditing, setIsEditing] = useState(false);
  const [isViewingHistory, setIsViewingHistory] = useState(false);

  return (
    <>
      <PanelShell aria-busy={isLoading || undefined}>
        <PanelContent data={data} isLoading={isLoading} error={error} onEdit={() => setIsEditing(true)} onViewHistory={() => setIsViewingHistory(true)} />
      </PanelShell>
      <SetStandardModal
        enterpriseCapabilityId={enterpriseCapabilityId}
        standard={data?.standard ?? null}
        opened={isEditing}
        onClose={() => setIsEditing(false)}
      />
      <StandardApplicationHistoryDialog
        enterpriseCapabilityId={enterpriseCapabilityId}
        opened={isViewingHistory}
        onClose={() => setIsViewingHistory(false)}
      />
    </>
  );
}

function PanelHeader() {
  return (
    <Group justify="space-between" align="center" mb="sm">
      <Title order={4}>Standard application</Title>
    </Group>
  );
}

function PanelLoading() {
  return (
    <>
      <PanelHeader />
      <Loader size="sm" />
    </>
  );
}

function PanelError() {
  return (
    <>
      <PanelHeader />
      <Alert color="red">Failed to load standard application.</Alert>
    </>
  );
}

function PanelReady({ data, onEdit, onViewHistory }: { data: ECStandardApplicationResponse; onEdit: () => void; onViewHistory: () => void }) {
  const standard = data.standard;
  if (standard) {
    return (
      <StandardDetail
        standard={standard}
        canEdit={!!data._links?.edit}
        hasHistory={!!data._links?.['x-history']}
        onEdit={onEdit}
        onViewHistory={onViewHistory}
      />
    );
  }
  return <NoStandardView canSet={!!data._links?.['x-set-standard']} onSet={onEdit} />;
}

function PanelContent({
  data,
  isLoading,
  error,
  onEdit,
  onViewHistory,
}: {
  data: ECStandardApplicationResponse | undefined;
  isLoading: boolean;
  error: Error | null;
  onEdit: () => void;
  onViewHistory: () => void;
}) {
  if (isLoading) return <PanelLoading />;
  if (error) return <PanelError />;
  if (!data) return null;
  return <PanelReady data={data} onEdit={onEdit} onViewHistory={onViewHistory} />;
}

function SetStandardModal({
  enterpriseCapabilityId,
  standard,
  opened,
  onClose,
}: {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  standard: StandardApplication | null;
  opened: boolean;
  onClose: () => void;
}) {
  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={standard ? 'Change standard application' : 'Set standard application'}
      size="lg"
      centered
      data-testid="standard-application-modal"
    >
      <SetStandardApplicationForm
        enterpriseCapabilityId={enterpriseCapabilityId}
        initialApplicationId={standard ? String(standard.applicationId) : undefined}
        initialNarrative={standard?.narrative}
        onSubmitted={onClose}
        onCancel={onClose}
      />
    </Modal>
  );
}

function PanelShell({ children, ...rest }: { children: React.ReactNode } & Record<string, unknown>) {
  return (
    <Box data-testid="standard-application-panel" component="section" {...rest}>
      {children}
    </Box>
  );
}

function NoStandardView({ canSet, onSet }: { canSet: boolean; onSet: () => void }) {
  return (
    <Stack gap="sm">
      <Group justify="space-between" align="center">
        <Title order={4}>Standard application</Title>
        <Badge variant="light" color="gray" data-testid="standard-application-empty-state">
          No standard yet
        </Badge>
      </Group>
      <Text c="dimmed">The architecture group has not set a standard application for this enterprise capability.</Text>
      {canSet && (
        <Group justify="flex-start">
          <Button onClick={onSet} data-testid="set-standard-application-button">
            Set standard
          </Button>
        </Group>
      )}
    </Stack>
  );
}

interface StandardDetailProps {
  standard: StandardApplication;
  canEdit: boolean;
  hasHistory: boolean;
  onEdit: () => void;
  onViewHistory: () => void;
}

function StandardDetail({ standard, canEdit, hasHistory, onEdit, onViewHistory }: StandardDetailProps) {
  return (
    <Stack gap="sm">
      <Group justify="space-between" align="center">
        <Title order={4}>Standard application</Title>
        {standard.applicationStale && (
          <Badge color="yellow" variant="light" data-testid="standard-application-stale-indicator">
            Application reference stale
          </Badge>
        )}
      </Group>
      <Group gap="sm" align="baseline">
        <Text fw={500} data-testid="standard-application-name">
          {standard.applicationName ?? '—'}
        </Text>
        <Text size="xs" c="dimmed">
          Set {new Date(standard.setAt).toLocaleDateString()}
        </Text>
      </Group>
      <Text data-testid="standard-application-narrative">{standard.narrative}</Text>
      <Group justify="flex-start" gap="xs">
        {canEdit && (
          <Button variant="default" onClick={onEdit} data-testid="change-standard-application-button">
            Change standard
          </Button>
        )}
        {hasHistory && (
          <Button variant="subtle" onClick={onViewHistory} data-testid="view-standard-application-history-button">
            View history
          </Button>
        )}
      </Group>
    </Stack>
  );
}
