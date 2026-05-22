import { Stack, Text } from '@mantine/core';

interface EditGrantsEmptyStateProps {
  statusFilter: string;
}

export function EditGrantsEmptyState({ statusFilter }: EditGrantsEmptyStateProps) {
  const message =
    statusFilter === 'all' ? 'No edit grants found' : `No ${statusFilter} edit grants`;

  return (
    <Stack align="center" gap="md" py="xl">
      <Text size="lg" c="dimmed">
        {message}
      </Text>
    </Stack>
  );
}
