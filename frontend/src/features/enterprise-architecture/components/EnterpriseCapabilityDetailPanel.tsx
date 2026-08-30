import { Badge, Box, CloseButton, Divider, Group, Paper, Stack, Text, Title } from '@mantine/core';
import { DirectionPanel } from '../../architecture-direction/components/DirectionPanel';
import { OnePagerActionButton } from '../../one-pagers/components/OnePagerActionButton';
import { StandardApplicationPanel } from '../../standard-application/components/StandardApplicationPanel';
import { useComposition } from '../hooks/useComposition';
import { useCompositionSummaries } from '../hooks/useCompositionSummaries';
import type { EnterpriseCapability } from '../types';
import classes from './EnterpriseCapabilityDetailPanel.module.css';
import { IncludedCapabilitiesSection } from './IncludedCapabilitiesSection';

const EMPTY_COUNT = '—';

interface EnterpriseCapabilityDetailPanelProps {
  capability: EnterpriseCapability;
  onClose: () => void;
}

function StatPair({ label, value, testId }: { label: string; value: number | string; testId: string }) {
  return (
    <Stack gap={2}>
      <Text size="xl" fw={700} c="blue.7" data-testid={testId}>
        {value}
      </Text>
      <Text size="xs" c="dimmed">
        {label}
      </Text>
    </Stack>
  );
}

export function EnterpriseCapabilityDetailPanel({ capability, onClose }: EnterpriseCapabilityDetailPanelProps) {
  const summariesQuery = useCompositionSummaries();
  const summary = summariesQuery.data?.get(capability.id);
  const compositionHref = summary?._links?.['x-composition']?.href;
  const directionHref = summary?._links?.['x-direction']?.href;
  const compositionQuery = useComposition(capability.id, compositionHref);

  const meta = compositionQuery.data?.meta;

  return (
    <Paper shadow="sm" radius="lg" p="xl" className={classes.panel}>
      <Stack gap="lg">
        <Group justify="space-between" align="flex-start" wrap="nowrap">
          <Stack gap="xs">
            <Title order={2}>{capability.name}</Title>
            {capability.category && (
              <Badge variant="light" color="gray" radius="sm">
                {capability.category}
              </Badge>
            )}
          </Stack>
          <CloseButton onClick={onClose} aria-label="Close detail panel" />
        </Group>

        {capability.description && (
          <Text size="sm" c="dimmed">
            {capability.description}
          </Text>
        )}

        <Box>
          <Divider />
          <Group gap={48} py="md">
            <StatPair
              label="Included capabilities"
              value={meta?.includedCount ?? EMPTY_COUNT}
              testId="stat-included-capabilities"
            />
            <StatPair label="Domains" value={meta?.domainCount ?? EMPTY_COUNT} testId="stat-domains" />
          </Group>
          <Divider />
        </Box>

        <DirectionPanel enterpriseCapabilityId={capability.id} directionHref={directionHref} />

        <StandardApplicationPanel enterpriseCapabilityId={capability.id} />

        <OnePagerActionButton subject={capability} subjectType="enterprise-capability" subjectId={capability.id} />

        <IncludedCapabilitiesSection composition={compositionQuery.data} isLoading={compositionQuery.isLoading} />
      </Stack>
    </Paper>
  );
}
