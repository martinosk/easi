import { Group, Text } from '@mantine/core';
import type React from 'react';
import type { OwnershipState } from '../../../api/types';
import { useOwnershipStatistics } from '../hooks/useComponentOwnership';
import { OWNERSHIP_STATE_LABELS } from './ComponentOwnershipSection';

const SUMMARY_STATES: OwnershipState[] = ['unknown', 'nominated', 'owned', 'managed'];

export const OwnershipSummary: React.FC = () => {
  const { data: statistics } = useOwnershipStatistics();
  if (!statistics) return null;

  return (
    <Group gap="sm" px="sm" py="xs" wrap="wrap" data-testid="ownership-summary">
      {SUMMARY_STATES.map((state) => (
        <Group key={state} gap="xs" wrap="nowrap">
          <Text size="xs" c="dimmed">
            {OWNERSHIP_STATE_LABELS[state]}
          </Text>
          <Text size="xs" fw={600}>
            {statistics[state]}
          </Text>
        </Group>
      ))}
    </Group>
  );
};
