import { Group, Progress, Stack, Text } from '@mantine/core';
import type { ImportProgress } from '../types';

interface ImportProgressStepProps {
  progress: ImportProgress;
}

function formatPhase(phase: string): string {
  return phase.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export function ImportProgressStep({ progress }: ImportProgressStepProps) {
  const { phase, totalItems, completedItems } = progress;
  const percentage = totalItems > 0 ? Math.round((completedItems / totalItems) * 100) : 0;

  return (
    <Stack gap="md">
      <Text c="dimmed" size="sm">
        Please wait while the import is in progress. Do not close this dialog.
      </Text>

      <Stack gap="sm">
        <Group justify="space-between">
          <Text fw={600} data-testid="progress-phase">
            {formatPhase(phase)}
          </Text>
          <Text c="dimmed" size="sm" data-testid="progress-stats">
            {completedItems} / {totalItems} items
          </Text>
        </Group>

        <Progress value={percentage} data-testid="progress-bar" aria-label="import progress" />

        <Text ta="right" size="sm" data-testid="progress-percentage">
          {percentage}%
        </Text>
      </Stack>
    </Stack>
  );
}
