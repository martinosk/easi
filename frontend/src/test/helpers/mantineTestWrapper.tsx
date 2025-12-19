import { ReactNode } from 'react';
import { MantineProvider } from '@mantine/core';
import { theme } from '../../theme/mantine';

export function MantineTestWrapper({ children }: { children: ReactNode }) {
  return <MantineProvider theme={theme}>{children}</MantineProvider>;
}
