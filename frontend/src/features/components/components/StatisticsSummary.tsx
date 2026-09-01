import { Group, Stack, Text } from '@mantine/core';
import type React from 'react';
import type { HostingClassification, OwnershipState } from '../../../api/types';
import { useComponentStatistics } from '../hooks/useComponentStatistics';
import { HOSTING_CLASSIFICATION_LABELS } from './ComponentHostingSection';
import { OWNERSHIP_STATE_LABELS } from './ComponentOwnershipSection';

const OWNERSHIP_STATES: OwnershipState[] = ['unknown', 'nominated', 'owned', 'managed'];
const HOSTING_CLASSIFICATIONS: HostingClassification[] = [
  'on-premises',
  'cloud',
  'saas',
  'third-party-hosted',
  'unknown',
];

interface DistributionRowProps {
  testId: string;
  caption: string;
  entries: Array<[string, number]>;
}

const DistributionRow: React.FC<DistributionRowProps> = ({ testId, caption, entries }) => (
  <Group gap="sm" wrap="wrap" data-testid={testId}>
    <Text size="xs" fw={600} c="dimmed">
      {caption}
    </Text>
    {entries.map(([label, count]) => (
      <Group key={label} gap="xs" wrap="nowrap">
        <Text size="xs" c="dimmed">
          {label}
        </Text>
        <Text size="xs" fw={600}>
          {count}
        </Text>
      </Group>
    ))}
  </Group>
);

export const StatisticsSummary: React.FC = () => {
  const { data: statistics } = useComponentStatistics();
  if (!statistics) return null;

  return (
    <Stack gap="xs" px="sm" py="xs" data-testid="statistics-summary">
      <DistributionRow
        testId="ownership-distribution"
        caption="Ownership"
        entries={OWNERSHIP_STATES.map((state) => [OWNERSHIP_STATE_LABELS[state], statistics[state]])}
      />
      <DistributionRow
        testId="hosting-distribution"
        caption="Hosting"
        entries={HOSTING_CLASSIFICATIONS.map((hosting) => [
          HOSTING_CLASSIFICATION_LABELS[hosting],
          statistics.hosting[hosting],
        ])}
      />
    </Stack>
  );
};
