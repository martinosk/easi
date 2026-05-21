import { Button, Group, Paper, Stack, Text } from '@mantine/core';
import type { EnterpriseCapabilityLink } from '../types';

interface GroupedDomain {
  domainName: string;
  links: EnterpriseCapabilityLink[];
}

interface LinkedCapabilitiesSectionProps {
  groups: GroupedDomain[];
  onUnlink: (link: EnterpriseCapabilityLink) => void;
  isUnlinking: boolean;
}

export function LinkedCapabilitiesSection({ groups, onUnlink, isUnlinking }: LinkedCapabilitiesSectionProps) {
  if (groups.length === 0) {
    return (
      <Text size="sm" c="dimmed" fs="italic">
        No capabilities linked yet. Use the Manage Links page to link domain capabilities.
      </Text>
    );
  }

  return (
    <Stack gap="md">
      {groups.map((group) => (
        <DomainGroup
          key={group.domainName}
          domainName={group.domainName}
          links={group.links}
          onUnlink={onUnlink}
          isUnlinking={isUnlinking}
        />
      ))}
    </Stack>
  );
}

interface DomainGroupProps {
  domainName: string;
  links: EnterpriseCapabilityLink[];
  onUnlink: (link: EnterpriseCapabilityLink) => void;
  isUnlinking: boolean;
}

function DomainGroup({ domainName, links, onUnlink, isUnlinking }: DomainGroupProps) {
  return (
    <Stack gap={4}>
      <Paper bg="gray.1" px="md" py="xs" radius="sm">
        <Text size="sm" fw={600} c="gray.7">
          {domainName}
        </Text>
      </Paper>
      <Stack gap={0}>
        {links.map((link) => (
          <Group key={link.id} justify="space-between" px="md" py="xs" gap="sm" wrap="nowrap">
            <Text size="sm">{link.domainCapabilityName || link.domainCapabilityId}</Text>
            {link._links?.delete && (
              <Button
                size="compact-xs"
                variant="subtle"
                color="red"
                onClick={() => onUnlink(link)}
                disabled={isUnlinking}
                title="Unlink capability"
              >
                Unlink
              </Button>
            )}
          </Group>
        ))}
      </Stack>
    </Stack>
  );
}
