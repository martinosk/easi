import { Box, Stack, Text } from '@mantine/core';
import React from 'react';

interface DetailFieldProps {
  label: string;
  children: React.ReactNode;
}

export const DetailField: React.FC<DetailFieldProps> = ({ label, children }) => (
  <Stack gap="xs">
    <Text component="label" size="xs" fw={600} c="dimmed" tt="uppercase" lts="0.05em">
      {label}
    </Text>
    <Box>{children}</Box>
  </Stack>
);
