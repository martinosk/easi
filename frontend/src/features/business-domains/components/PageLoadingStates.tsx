import { Alert, Center, Group, Loader, Text } from '@mantine/core';

interface PageLoadingStatesProps {
  isLoading: boolean;
  hasData: boolean;
  error: Error | null;
  children: React.ReactNode;
}

export function PageLoadingStates({ isLoading, hasData, error, children }: PageLoadingStatesProps) {
  if (isLoading && !hasData) {
    return (
      <Center p="xl">
        <Group gap="sm">
          <Loader size="sm" />
          <Text c="dimmed">Loading business domains...</Text>
        </Group>
      </Center>
    );
  }

  if (error && !hasData) {
    return (
      <Center p="xl">
        <Alert color="red" data-testid="domains-error">
          {error.message}
        </Alert>
      </Center>
    );
  }

  return <>{children}</>;
}
