import { Group, Text } from '@mantine/core';
import type { CapabilityRealization, ComponentId } from '../../../api/types';
import { ApplicationChip } from './ApplicationChip';

export interface ApplicationChipListProps {
  realizations: CapabilityRealization[];
  onApplicationClick: (componentId: ComponentId) => void;
}

const MAX_VISIBLE_CHIPS = 5;

export function ApplicationChipList({ realizations, onApplicationClick }: ApplicationChipListProps) {
  const visibleRealizations = realizations.slice(0, MAX_VISIBLE_CHIPS);
  const overflowCount = realizations.length - MAX_VISIBLE_CHIPS;

  if (realizations.length === 0) {
    return null;
  }

  return (
    <Group gap={6} align="center" wrap="wrap">
      {visibleRealizations.map((realization) => (
        <ApplicationChip key={realization.id} realization={realization} onClick={onApplicationClick} />
      ))}
      {overflowCount > 0 && (
        <Text size="xs" c="dimmed" fw={500}>
          +{overflowCount} more
        </Text>
      )}
    </Group>
  );
}
