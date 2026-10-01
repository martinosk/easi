import React from 'react';
import { Center, Loader, Stack, Text } from '@mantine/core';
import { type ScreenFill, screenFillProps } from './screenFill';

interface LoadingScreenProps {
  fill?: ScreenFill;
}

export const LoadingScreen: React.FC<LoadingScreenProps> = ({ fill = 'viewport' }) => {
  return (
    <Center {...screenFillProps(fill)} data-testid="loading-screen">
      <Stack align="center" gap="lg">
        <Loader size="lg" />
        <Text>Loading component modeler...</Text>
      </Stack>
    </Center>
  );
};
