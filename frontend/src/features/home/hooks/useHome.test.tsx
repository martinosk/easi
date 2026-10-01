import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { buildHomeFixture } from '../../../test/mocks/home';
import { homeQueryKeys } from '../queryKeys';
import { useHome } from './useHome';

function renderUseHome() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: 5 * 60 * 1000, refetchOnMount: true, refetchOnWindowFocus: false },
    },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, ...renderHook(() => useHome(), { wrapper }) };
}

describe('useHome', () => {
  it('loads the caller home from the API', async () => {
    const { result } = renderUseHome();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(buildHomeFixture());
  });

  it('always refetches on mount and on window focus', async () => {
    const { result, queryClient } = renderUseHome();
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const options = queryClient.getQueryCache().find({ queryKey: homeQueryKeys.detail() })?.options as Record<
      string,
      unknown
    >;

    expect(options.staleTime).toBe(0);
    expect(options.refetchOnMount).toBe('always');
    expect(options.refetchOnWindowFocus).toBe(true);
  });
});
