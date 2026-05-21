import { Badge, Group, Paper, Stack, Text, Title } from '@mantine/core';
import { useState } from 'react';
import toast from 'react-hot-toast';
import type { Capability } from '../../../api/types';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { EnterpriseCapability } from '../types';
import classes from './EnterpriseCapabilityCard.module.css';

export interface EnterpriseCapabilityCardProps {
  capability: EnterpriseCapability;
  onDrop: (capability: Capability) => void;
}

export function EnterpriseCapabilityCard({ capability, onDrop }: EnterpriseCapabilityCardProps) {
  const [isDragOver, setIsDragOver] = useState(false);
  const canAcceptLink = capability._links?.['x-create-link'] !== undefined;

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    if (!canAcceptLink) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setIsDragOver(true);
  };

  const handleDragLeave = () => setIsDragOver(false);

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    if (!canAcceptLink) return;
    e.preventDefault();
    setIsDragOver(false);

    try {
      const data = e.dataTransfer.getData('application/json');
      const domainCapability = JSON.parse(data) as Capability;
      onDrop(domainCapability);
    } catch {
      toast.error('Failed to link capability. Invalid data format.');
    }
  };

  return (
    <Paper
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      withBorder
      radius="md"
      p="md"
      mb="md"
      bg={isDragOver ? 'blue.0' : 'white'}
      data-drag-over={isDragOver || undefined}
      className={classes.card}
    >
      <Stack gap="xs">
        <Stack gap={4} align="flex-start">
          <Title order={4}>{capability.name}</Title>
          {capability.category && (
            <Badge variant="light" color="gray" radius="sm">
              {capability.category}
            </Badge>
          )}
        </Stack>

        {capability.description && (
          <Text size="sm" c="dimmed">
            {capability.description}
          </Text>
        )}

        <Group gap="lg" mt={4}>
          <Group gap={4}>
            <Text size="sm" fw={500} c="gray.7">
              Links:
            </Text>
            <Text size="sm" c="blue.6" fw={600}>
              {capability.linkCount}
            </Text>
            <HelpTooltip content="Number of domain capabilities linked to this enterprise capability" iconOnly />
          </Group>
          <Group gap={4}>
            <Text size="sm" fw={500} c="gray.7">
              Domains:
            </Text>
            <Text size="sm" c="blue.6" fw={600}>
              {capability.domainCount}
            </Text>
            <HelpTooltip content="Number of business domains containing linked capabilities" iconOnly />
          </Group>
        </Group>

        {capability.linkCount === 0 && (
          <Text size="xs" c="dimmed" fs="italic">
            Drag domain capabilities here to link them
          </Text>
        )}
      </Stack>
    </Paper>
  );
}
