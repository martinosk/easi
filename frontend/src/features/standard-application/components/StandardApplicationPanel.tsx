import { Alert, Badge, Box, Button, Group, Loader, Modal, Stack, Text, Title } from '@mantine/core';
import { useMemo, useState } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { useComponents } from '../../components/hooks/useComponents';
import { useStandardApplicationForEnterpriseCapability } from '../hooks/useStandardApplication';
import type { StandardApplication } from '../types';
import { SetStandardApplicationForm } from './SetStandardApplicationForm';
import { StandardApplicationHistoryDialog } from './StandardApplicationHistoryDialog';

interface StandardApplicationPanelProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

export function StandardApplicationPanel({ enterpriseCapabilityId }: StandardApplicationPanelProps) {
  const { data, isLoading, error } = useStandardApplicationForEnterpriseCapability(enterpriseCapabilityId);
  const [isEditing, setIsEditing] = useState(false);
  const [isViewingHistory, setIsViewingHistory] = useState(false);

  if (isLoading) {
    return (
      <PanelShell aria-busy="true">
        <Group justify="space-between" align="center" mb="sm">
          <Title order={4}>Standard application</Title>
        </Group>
        <Loader size="sm" />
      </PanelShell>
    );
  }

  if (error) {
    return (
      <PanelShell>
        <Group justify="space-between" align="center" mb="sm">
          <Title order={4}>Standard application</Title>
        </Group>
        <Alert color="red">Failed to load standard application.</Alert>
      </PanelShell>
    );
  }

  const standard = data?.standard ?? null;
  const canSet = !!data?._links?.['x-set-standard'];
  const canEdit = !!data?._links?.edit;
  const hasHistory = !!data?._links?.['x-history'];

  return (
    <>
      <PanelShell>
        {standard ? (
          <StandardDetail
            standard={standard}
            canEdit={canEdit}
            hasHistory={hasHistory}
            onEdit={() => setIsEditing(true)}
            onViewHistory={() => setIsViewingHistory(true)}
          />
        ) : (
          <NoStandardView canSet={canSet} onSet={() => setIsEditing(true)} />
        )}
      </PanelShell>
      <Modal
        opened={isEditing}
        onClose={() => setIsEditing(false)}
        title={standard ? 'Change standard application' : 'Set standard application'}
        size="lg"
        centered
        data-testid="standard-application-modal"
      >
        <SetStandardApplicationForm
          enterpriseCapabilityId={enterpriseCapabilityId}
          initialApplicationId={standard ? String(standard.applicationId) : undefined}
          initialNarrative={standard?.narrative}
          onSubmitted={() => setIsEditing(false)}
          onCancel={() => setIsEditing(false)}
        />
      </Modal>
      <StandardApplicationHistoryDialog
        enterpriseCapabilityId={enterpriseCapabilityId}
        opened={isViewingHistory}
        onClose={() => setIsViewingHistory(false)}
      />
    </>
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
  const { data: components } = useComponents();
  const applicationName = useMemo(() => {
    const match = (components ?? []).find((c) => String(c.id) === String(standard.applicationId));
    return match?.name ?? String(standard.applicationId);
  }, [components, standard.applicationId]);

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
          {applicationName}
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
