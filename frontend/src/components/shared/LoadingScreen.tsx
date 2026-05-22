import React from 'react';
import { Center, Loader, Stack, Text } from '@mantine/core';

export const LoadingScreen: React.FC = () => {
  return (
    <Center mih="100vh">
      <Stack align="center" gap="lg">
        <Loader size="lg" />
        <Text>Loading component modeler...</Text>
      </Stack>
    </Center>
  );
};
