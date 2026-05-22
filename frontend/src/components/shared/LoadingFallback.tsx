import { Center, Group, Loader, Text } from '@mantine/core';

interface LoadingFallbackProps {
  message?: string;
}

export function LoadingFallback({ message = 'Loading...' }: LoadingFallbackProps) {
  return (
    <Center mih="200px" p="xl">
      <Group gap="md">
        <Loader size="sm" />
        <Text size="sm" c="dimmed">
          {message}
        </Text>
      </Group>
    </Center>
  );
}
