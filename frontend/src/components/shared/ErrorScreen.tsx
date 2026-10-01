import React from 'react';
import { Button, Center, Stack, Text, Title } from '@mantine/core';
import { type ScreenFill, screenFillProps } from './screenFill';

interface ErrorScreenProps {
  error: string;
  onRetry: () => void;
  retryLabel?: string;
  title?: string;
  fill?: ScreenFill;
}

export const ErrorScreen: React.FC<ErrorScreenProps> = ({
  error,
  onRetry,
  retryLabel = 'Retry',
  title = 'Error Loading Data',
  fill = 'viewport',
}) => {
  return (
    <Center {...screenFillProps(fill)} p="lg" data-testid="error-screen">
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
