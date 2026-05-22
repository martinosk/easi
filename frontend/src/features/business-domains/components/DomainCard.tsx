import { Group, Paper, Stack, Text, UnstyledButton } from '@mantine/core';
import type { BusinessDomain } from '../../../api/types';

interface DomainCardProps {
  domain: BusinessDomain;
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
  isSelected?: boolean;
}

export function DomainCard({ domain, onVisualize, onContextMenu, isSelected }: DomainCardProps) {
  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onContextMenu(e, domain);
  };

  return (
    <UnstyledButton
      onClick={() => onVisualize(domain)}
      onContextMenu={handleContextMenu}
      data-testid={`domain-card-${domain.id}`}
      w="100%"
    >
      <Paper
        withBorder
        p="sm"
        radius="md"
        bg={isSelected ? 'blue.0' : 'white'}
        style={isSelected ? { borderColor: 'var(--mantine-color-blue-6)', borderWidth: 2 } : undefined}
      >
        <Stack gap={4}>
          <Group justify="space-between" align="flex-start" wrap="nowrap">
            <Text fw={600} size="md">
              {domain.name}
            </Text>
            <Text size="xs" c="dimmed" style={{ whiteSpace: 'nowrap' }}>
              {domain.capabilityCount} {domain.capabilityCount === 1 ? 'capability' : 'capabilities'}
            </Text>
          </Group>
          <Text size="sm" c="dimmed">
            {domain.description || 'No description'}
          </Text>
          <Text size="xs" c="dimmed">
            Created: {new Date(domain.createdAt).toLocaleDateString()}
          </Text>
        </Stack>
      </Paper>
    </UnstyledButton>
  );
}
