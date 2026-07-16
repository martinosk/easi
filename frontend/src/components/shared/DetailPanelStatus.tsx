import { Center, Group, Loader, Stack, Text } from '@mantine/core';

export function DetailPanelLoading() {
  return (
    <Stack gap="sm" p="md">
      <Center py="xl">
        <Group gap="xs">
          <Loader size="sm" />
          <Text c="dimmed">Loading...</Text>
        </Group>
      </Center>
    </Stack>
  );
}

export function DetailPanelFailure({ message }: { message: string }) {
  return (
    <Stack gap="sm" p="md">
      <Center py="xl">
        <Text c="red">{message}</Text>
      </Center>
    </Stack>
  );
}
