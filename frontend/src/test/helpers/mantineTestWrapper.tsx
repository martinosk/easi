/* eslint-disable react-refresh/only-export-components */
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router-dom";
import { MantineProvider } from "@mantine/core";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { theme } from "../../theme/mantine";

function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

const testQueryClient = createTestQueryClient();

export function MantineTestWrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={testQueryClient}>
      <MantineProvider theme={theme}>
        <MemoryRouter>{children}</MemoryRouter>
      </MantineProvider>
    </QueryClientProvider>
  );
}

export function createMantineTestWrapper() {
  const queryClient = createTestQueryClient();

  const Wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme}>
        <MemoryRouter>{children}</MemoryRouter>
      </MantineProvider>
    </QueryClientProvider>
  );

  return { Wrapper, queryClient };
}
