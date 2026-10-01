import { Paper, Stack, Text } from '@mantine/core';
import type { ReactNode } from 'react';

interface TileProps {
  label: string;
  testId: string;
  children: ReactNode;
}

export function Tile({ label, testId, children }: TileProps) {
  return (
    <Paper withBorder radius="md" p="md" data-testid={testId}>
      <Stack gap="xs">
        <Text size="xs" tt="uppercase" fw={600} c="dimmed">
          {label}
        </Text>
        {children}
      </Stack>
    </Paper>
  );
}

export function TileTotal({ value }: { value: number }) {
  return (
    <Text fz="xl" fw={700}>
      {value}
    </Text>
  );
}
