import { Alert, Anchor, Badge, Box, Button, Group, List, Loader, Modal, Stack, Text, Title } from '@mantine/core';
import { useMemo, useState } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
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
      <Text c="dimmed">
        The architecture group has not captured a direction on this enterprise capability.
      </Text>
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

function usePlacementDomainNames() {
  const { data: domainsResponse } = useBusinessDomainsQuery();
  return useMemo(() => {
    const lookup = new Map<string, string>();
    for (const d of domainsResponse?.data ?? []) {
      lookup.set(d.id, d.name);
    }
    return (id: string) => lookup.get(id);
  }, [domainsResponse]);
}

function DirectionDetail({ direction, enterpriseCapabilityId }: DirectionDetailProps) {
  const resolvePlacementDomain = usePlacementDomainNames();
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

      <Box>
        <Text size="sm" fw={600}>
          Sources
        </Text>
        <SourceList sources={direction.sourceCapabilities} />
      </Box>

      {direction.placements.length > 0 && (
        <Box>
          <Text size="sm" fw={600}>
            Placements
          </Text>
          <PlacementList placements={direction.placements} resolveDomainName={resolvePlacementDomain} />
        </Box>
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

function SourceList({ sources }: { sources: Direction['sourceCapabilities'] }) {
  return (
    <List size="sm" listStyleType="disc" data-testid="direction-sources">
      {sources.map((source) => (
        <List.Item key={source.id}>
          <Anchor component="span" inherit>
            {source.name ?? '—'}
          </Anchor>
          {source.businessDomainName && (
            <Text component="span" size="xs" c="dimmed">
              {' · '}
              {source.businessDomainName}
            </Text>
          )}
          {source.stale && (
            <Text component="span" size="xs" c="red" data-testid="stale-reference" ml="xs">
              (stale — source capability deleted)
            </Text>
          )}
        </List.Item>
      ))}
    </List>
  );
}

interface PlacementListProps {
  placements: Direction['placements'];
  resolveDomainName: (id: string) => string | undefined;
}

function PlacementList({ placements, resolveDomainName }: PlacementListProps) {
  return (
    <List size="sm" listStyleType="disc">
      {placements.map((placement, i) => {
        const domain = resolveDomainName(placement.targetBusinessDomainId);
        const name = placement.resultingName;
        return (
          // biome-ignore lint/suspicious/noArrayIndexKey: placements may share a target domain ID; index is a stable composite key here
          <List.Item key={`${placement.targetBusinessDomainId}-${i}`}>
            {name ? (
              <Anchor component="span" inherit>
                {name}
              </Anchor>
            ) : (
              <Text component="code" size="xs">
                {placement.targetBusinessDomainId}
              </Text>
            )}
            {domain && (
              <Text component="span" size="xs" c="dimmed">
                {' · '}
                {domain}
              </Text>
            )}
          </List.Item>
        );
      })}
    </List>
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
