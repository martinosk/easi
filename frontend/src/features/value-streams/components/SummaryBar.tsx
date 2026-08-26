import { Group, Text } from '@mantine/core';

interface SummaryBarProps {
  stageCount: number;
  capabilityCount: number;
}

interface SummaryItemProps {
  count: number;
  singular: string;
  plural: string;
}

function SummaryItem({ count, singular, plural }: SummaryItemProps) {
  return (
    <Group gap="xs" align="baseline">
      <Text size="xl" fw={700}>
        {count}
      </Text>
      <Text size="sm" c="dimmed">
        {count === 1 ? singular : plural}
      </Text>
    </Group>
  );
}

export function SummaryBar({ stageCount, capabilityCount }: SummaryBarProps) {
  return (
    <Group gap="xl" data-testid="summary-bar">
      <SummaryItem count={stageCount} singular="Stage" plural="Stages" />
      <SummaryItem count={capabilityCount} singular="Capability" plural="Capabilities" />
    </Group>
  );
}
