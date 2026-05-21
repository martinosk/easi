import { Badge, Box, CloseButton, Divider, Group, Loader, Paper, Stack, Text, Title } from '@mantine/core';
import { useMemo } from 'react';
import { DirectionPanel } from '../../architecture-direction/components/DirectionPanel';
import { useEnterpriseCapabilityLinks, useUnlinkDomainCapability } from '../hooks/useEnterpriseCapabilities';
import type { EnterpriseCapability, EnterpriseCapabilityLink } from '../types';
import { LinkedCapabilitiesSection } from './LinkedCapabilitiesSection';
import classes from './EnterpriseCapabilityDetailPanel.module.css';

interface EnterpriseCapabilityDetailPanelProps {
  capability: EnterpriseCapability;
  onClose: () => void;
}

interface GroupedDomain {
  domainName: string;
  links: EnterpriseCapabilityLink[];
}

function groupLinksByDomain(links: EnterpriseCapabilityLink[]): GroupedDomain[] {
  const grouped = new Map<string, EnterpriseCapabilityLink[]>();
  for (const link of links) {
    const domainName = link.businessDomainName || 'Unassigned';
    const existing = grouped.get(domainName);
    if (existing) existing.push(link);
    else grouped.set(domainName, [link]);
  }
  return Array.from(grouped.entries())
    .map(([domainName, domainLinks]) => ({ domainName, links: domainLinks }))
    .sort((a, b) => {
      if (a.domainName === 'Unassigned') return 1;
      if (b.domainName === 'Unassigned') return -1;
      return a.domainName.localeCompare(b.domainName);
    });
}

function StatPair({ label, value }: { label: string; value: number }) {
  return (
    <Group gap="xs" wrap="nowrap">
      <Text size="sm" c="dimmed">
        {label}:
      </Text>
      <Text size="sm" fw={600} c="blue.6">
        {value}
      </Text>
    </Group>
  );
}

export function EnterpriseCapabilityDetailPanel({ capability, onClose }: EnterpriseCapabilityDetailPanelProps) {
  const { data: links, isLoading } = useEnterpriseCapabilityLinks(capability.id);
  const unlinkMutation = useUnlinkDomainCapability();

  const groupedLinks = useMemo(() => (links ? groupLinksByDomain(links) : []), [links]);

  const handleUnlink = async (link: EnterpriseCapabilityLink) => {
    await unlinkMutation.mutateAsync({
      enterpriseCapabilityId: capability.id,
      linkId: link.id,
    });
  };

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
          <Group gap="xl" py="md">
            <StatPair label="Links" value={capability.linkCount} />
            <StatPair label="Domains" value={capability.domainCount} />
          </Group>
          <Divider />
        </Box>

        <DirectionPanel enterpriseCapabilityId={capability.id} />

        <Stack gap="sm">
          <Title order={4}>Linked Capabilities</Title>
          {isLoading ? (
            <Group gap="xs">
              <Loader size="sm" />
              <Text size="sm" c="dimmed">
                Loading linked capabilities...
              </Text>
            </Group>
          ) : (
            <LinkedCapabilitiesSection
              groups={groupedLinks}
              onUnlink={handleUnlink}
              isUnlinking={unlinkMutation.isPending}
            />
          )}
        </Stack>
      </Stack>
    </Paper>
  );
}
