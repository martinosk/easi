/* eslint-disable react-refresh/only-export-components */

import { MantineProvider } from '@mantine/core';
import { QueryClientProvider } from '@tanstack/react-query';
import { type ReactNode, useState } from 'react';
import { theme } from '../../theme/mantine';
import { createTestQueryClient } from './renderWithProviders';

export function MantineTestWrapper({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createTestQueryClient);
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme}>{children}</MantineProvider>
    </QueryClientProvider>
  );
}
