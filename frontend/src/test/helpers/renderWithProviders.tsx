/* eslint-disable react-refresh/only-export-components */
import type { ReactNode } from 'react';
import { render } from '@testing-library/react';
import type { RenderOptions, RenderResult } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { MemoryRouter } from 'react-router-dom';
import type { MemoryRouterProps } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { theme } from '../../theme/mantine';

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

export interface RenderWithProvidersOptions extends Omit<RenderOptions, 'wrapper'> {
  routerProps?: MemoryRouterProps;
  withRouter?: boolean;
  queryClient?: QueryClient;
}

interface TestProvidersProps {
  children: ReactNode;
  routerProps?: MemoryRouterProps;
  withRouter?: boolean;
  queryClient?: QueryClient;
}

function TestProviders({
  children,
  routerProps,
  withRouter = true,
  queryClient,
}: TestProvidersProps) {
  const testQueryClient = queryClient ?? createTestQueryClient();

  const content = (
    <QueryClientProvider client={testQueryClient}>
      <MantineProvider theme={theme}>
        {children}
      </MantineProvider>
    </QueryClientProvider>
  );

  if (!withRouter) {
    return content;
  }

  return (
    <MemoryRouter {...routerProps}>
      {content}
    </MemoryRouter>
  );
}

export function renderWithProviders(
  ui: React.ReactElement,
  options: RenderWithProvidersOptions = {}
): RenderResult & { queryClient: QueryClient } {
  const { routerProps, withRouter = true, queryClient, ...renderOptions } = options;
  const testQueryClient = queryClient ?? createTestQueryClient();

  const result = render(ui, {
    wrapper: ({ children }) => (
      <TestProviders
        routerProps={routerProps}
        withRouter={withRouter}
        queryClient={testQueryClient}
      >
        {children}
      </TestProviders>
    ),
    ...renderOptions,
  });

  return {
    ...result,
    queryClient: testQueryClient,
  };
}

export { TestProviders, createTestQueryClient };
