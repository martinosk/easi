import React from 'react';
import { Button, Center, Stack, Text, Title } from '@mantine/core';

interface ErrorScreenProps {
  error: string;
  onRetry: () => void;
  retryLabel?: string;
  title?: string;
}

export const ErrorScreen: React.FC<ErrorScreenProps> = ({
  error,
  onRetry,
  retryLabel = 'Retry',
  title = 'Error Loading Data',
}) => {
  return (
    <Center mih="100vh" p="lg">
      <Stack align="center" gap="lg">
        <Title order={2} c="red">
          {title}
        </Title>
        <Text c="dimmed">{error}</Text>
        <Button onClick={onRetry}>{retryLabel}</Button>
      </Stack>
    </Center>
  );
};
