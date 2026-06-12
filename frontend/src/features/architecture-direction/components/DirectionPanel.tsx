import { Alert, Badge, Box, Button, Group, Loader, Modal, Stack, Text, Title } from '@mantine/core';
import { useState } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import {
  useAgreeDirection,
  useDirectionForEnterpriseCapability,
  useProposeDirection,
  useRejectDirection,
} from '../hooks/useDirection';
import type { Direction } from '../types';
import { CaptureDirectionForm } from './CaptureDirectionForm';
import { DirectionStatusBadge } from './DirectionStatusBadge';

interface DirectionPanelProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

const TYPE_LABELS: Record<Direction['type'], string> = {
  consolidate: 'Consolidate',
  decompose: 'Decompose',
  stay: 'Stay',
};

const HORIZON_LABELS: Record<Direction['horizon'], string> = {
  now: 'Now',
  next: 'Next',
  later: 'Later',
};

export function DirectionPanel({ enterpriseCapabilityId }: DirectionPanelProps) {
  const { data, isLoading, error } = useDirectionForEnterpriseCapability(enterpriseCapabilityId);
  const [isCapturing, setIsCapturing] = useState(false);

  if (isLoading) {
    return (
      <PanelShell aria-busy="true">
        <Group justify="space-between" align="center" mb="sm">
          <Title order={4}>Direction</Title>
        </Group>
        <Loader size="sm" />
      </PanelShell>
    );
  }

  if (error) {
    return (
      <PanelShell>
        <Group justify="space-between" align="center" mb="sm">
          <Title order={4}>Direction</Title>
        </Group>
        <Alert color="red">Failed to load direction.</Alert>
      </PanelShell>
    );
  }

  const direction = data?.direction ?? null;
  const canCapture = !!data?._links?.['x-capture-direction'];

  return (
    <>
      <PanelShell>
        {direction ? (
          <DirectionDetail direction={direction} enterpriseCapabilityId={enterpriseCapabilityId} />
        ) : (
          <NoDirectionView canCapture={canCapture} onCapture={() => setIsCapturing(true)} />
        )}
      </PanelShell>
      <Modal
        opened={isCapturing}
        onClose={() => setIsCapturing(false)}
        title="Capture a direction"
        size="lg"
        centered
        data-testid="capture-direction-modal"
      >
        <CaptureDirectionForm
          enterpriseCapabilityId={enterpriseCapabilityId}
          onCaptured={() => setIsCapturing(false)}
          onCancel={() => setIsCapturing(false)}
        />
      </Modal>
    </>
  );
}

function PanelShell({ children, ...rest }: { children: React.ReactNode } & Record<string, unknown>) {
  return (
    <Box data-testid="direction-panel" component="section" {...rest}>
      {children}
    </Box>
  );
}

function NoDirectionView({ canCapture, onCapture }: { canCapture: boolean; onCapture: () => void }) {
  return (
    <Stack gap="sm">
      <Group justify="space-between" align="center">
        <Title order={4}>Direction</Title>
        <Badge variant="light" color="gray" data-testid="direction-empty-state">
          No direction set
        </Badge>
      </Group>
      <Text c="dimmed">The architecture group has not captured a direction on this enterprise capability.</Text>
      {canCapture && (
        <Group justify="flex-start">
          <Button onClick={onCapture} data-testid="capture-direction-button">
            Capture direction
          </Button>
        </Group>
      )}
    </Stack>
  );
}

interface DirectionDetailProps {
  direction: Direction;
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

function DirectionDetail({ direction, enterpriseCapabilityId }: DirectionDetailProps) {
  return (
    <Stack gap="sm">
      <Group justify="space-between" align="center">
        <Group gap="sm">
          <Title order={4}>Direction</Title>
          <Text c="dimmed" data-testid="direction-type">
            {TYPE_LABELS[direction.type]}
          </Text>
        </Group>
        <DirectionStatusBadge status={direction.status} />
      </Group>

      <DirectionNarrative narrative={direction.narrative} />

      <Box>
        <Text size="sm" fw={600}>
          Horizon
        </Text>
        <Text size="sm">{HORIZON_LABELS[direction.horizon]}</Text>
      </Box>

      {(direction.status === 'agreed' || direction.status === 'proposed') && (
        <Alert color="gray" variant="light" data-testid="direction-frozen-callout">
          This direction is {direction.status} and its composition is frozen. To recompose, reject this direction and
          capture a new one.
        </Alert>
      )}

      <DirectionActions direction={direction} enterpriseCapabilityId={enterpriseCapabilityId} />
    </Stack>
  );
}

function DirectionNarrative({ narrative }: { narrative: Direction['narrative'] }) {
  if (narrative) {
    return (
      <Text data-testid="direction-narrative" size="sm">
        {narrative}
      </Text>
    );
  }
  return (
    <Text c="dimmed" size="sm" fs="italic">
      No narrative yet. Add one before advancing this direction to proposed.
    </Text>
  );
}

function DirectionActions({ direction, enterpriseCapabilityId }: DirectionDetailProps) {
  const proposeMutation = useProposeDirection();
  const agreeMutation = useAgreeDirection();
  const rejectMutation = useRejectDirection();
  const links = direction._links ?? {};
  const anyPending = proposeMutation.isPending || agreeMutation.isPending || rejectMutation.isPending;

  return (
    <Group gap="sm" data-testid="direction-actions">
      {links['x-propose'] && (
        <Button
          data-testid="advance-to-proposed"
          disabled={anyPending}
          loading={proposeMutation.isPending}
          onClick={() => proposeMutation.mutate({ enterpriseCapabilityId })}
        >
          Advance to proposed
        </Button>
      )}
      {links['x-agree'] && (
        <Button
          data-testid="advance-to-agreed"
          disabled={anyPending}
          loading={agreeMutation.isPending}
          onClick={() => agreeMutation.mutate({ enterpriseCapabilityId })}
        >
          Advance to agreed
        </Button>
      )}
      {links['x-reject'] && (
        <Button
          variant="default"
          data-testid="reject-direction"
          disabled={anyPending}
          loading={rejectMutation.isPending}
          onClick={() => rejectMutation.mutate({ enterpriseCapabilityId })}
        >
          Reject
        </Button>
      )}
    </Group>
  );
}
