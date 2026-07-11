import { Badge, Box, CloseButton, Divider, Group, Paper, Stack, Text, Title } from '@mantine/core';
import { DirectionPanel } from '../../architecture-direction/components/DirectionPanel';
import { OnePagerActionButton, OnePagerFactsSection } from '../../one-pagers';
import { StandardApplicationPanel } from '../../standard-application/components/StandardApplicationPanel';
import { useComposition } from '../hooks/useComposition';
import type { EnterpriseCapability } from '../types';
import classes from './EnterpriseCapabilityDetailPanel.module.css';
import { IncludedCapabilitiesSection } from './IncludedCapabilitiesSection';

interface EnterpriseCapabilityDetailPanelProps {
  capability: EnterpriseCapability;
  onClose: () => void;
}

function StatPair({ label, value }: { label: string; value: number | string }) {
  return (
    <Stack gap={2}>
      <Text size="xl" fw={700} c="blue.7">
        {value}
      </Text>
      <Text size="xs" c="dimmed">
        {label}
      </Text>
    </Stack>
  );
}

export function EnterpriseCapabilityDetailPanel({ capability, onClose }: EnterpriseCapabilityDetailPanelProps) {
  const compositionQuery = useComposition(capability.id);

  const meta = compositionQuery.data?.meta;
  const includedCount = meta?.includedCount ?? capability.includedCapabilityCount;
  const domainCount = meta?.domainCount ?? capability.domainCount;

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
            <StatPair label="Included capabilities" value={includedCount} />
            <StatPair label="Domains" value={domainCount || '—'} />
          </Group>
          <Divider />
        </Box>

        <DirectionPanel enterpriseCapabilityId={capability.id} />

        <StandardApplicationPanel enterpriseCapabilityId={capability.id} />

        <OnePagerActionButton subject={capability} subjectType="enterprise-capability" subjectId={capability.id} />
        <OnePagerFactsSection subjectType="enterprise-capability" subjectId={capability.id} />

        <IncludedCapabilitiesSection
          composition={compositionQuery.data}
          isLoading={compositionQuery.isLoading}
        />
      </Stack>
    </Paper>
  );
}
