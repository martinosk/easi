import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('react-hot-toast', () => ({ default: { success: vi.fn(), error: vi.fn() } }));
vi.mock('../api/stewardshipApi', () => ({
  stewardshipApi: {
    getDomainStewardships: vi.fn(),
    assign: vi.fn(),
    release: vi.fn(),
  },
}));

import toast from 'react-hot-toast';
import { stewardshipApi } from '../api/stewardshipApi';
import { stewardshipQueryKeys } from '../queryKeys';
import { useAssignSteward, useDomainStewardships, useReleaseSteward } from './useStewardships';

const assignLink = { href: '/api/v1/stewardships/d-1/assessment', method: 'PUT' as const };
const releaseLink = { href: '/api/v1/stewardships/d-1/assessment', method: 'DELETE' as const };

function setup() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const invalidate = vi.spyOn(queryClient, 'invalidateQueries');
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { wrapper, invalidate };
}

describe('useDomainStewardships', () => {
  beforeEach(() => vi.clearAllMocks());

  it('follows the domain x-stewardships href', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue({
      domainId: 'd-1',
      domainName: 'Sales',
      fallback: null,
      data: [],
      _links: {},
    });
    const { wrapper } = setup();

    const { result } = renderHook(() => useDomainStewardships('d-1', '/api/v1/stewardships?domainId=d-1'), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(stewardshipApi.getDomainStewardships).toHaveBeenCalledWith('/api/v1/stewardships?domainId=d-1');
  });

  it('does not fetch without an href', () => {
    const { wrapper } = setup();
    renderHook(() => useDomainStewardships('d-1', undefined), { wrapper });
    expect(stewardshipApi.getDomainStewardships).not.toHaveBeenCalled();
  });
});

describe('steward mutations', () => {
  beforeEach(() => vi.clearAllMocks());

  it('assigning follows x-assign and invalidates the domain stewardships', async () => {
    vi.mocked(stewardshipApi.assign).mockResolvedValue({} as never);
    const { wrapper, invalidate } = setup();
    const { result } = renderHook(() => useAssignSteward('d-1'), { wrapper });

    await act(() => result.current.mutateAsync({ link: assignLink, stewardId: 'mette' }));

    expect(stewardshipApi.assign).toHaveBeenCalledWith(assignLink, 'mette');
    expect(invalidate).toHaveBeenCalledWith({ queryKey: stewardshipQueryKeys.domain('d-1') });
    expect(toast.success).toHaveBeenCalled();
  });

  it('releasing follows x-release and invalidates the domain stewardships', async () => {
    vi.mocked(stewardshipApi.release).mockResolvedValue();
    const { wrapper, invalidate } = setup();
    const { result } = renderHook(() => useReleaseSteward('d-1'), { wrapper });

    await act(() => result.current.mutateAsync(releaseLink));

    expect(stewardshipApi.release).toHaveBeenCalledWith(releaseLink);
    expect(invalidate).toHaveBeenCalledWith({ queryKey: stewardshipQueryKeys.domain('d-1') });
  });

  it('reports a failed assignment', async () => {
    vi.mocked(stewardshipApi.assign).mockRejectedValue(new Error('Steward must be an active user'));
    const { wrapper } = setup();
    const { result } = renderHook(() => useAssignSteward('d-1'), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ link: assignLink, stewardId: 'gone' }).catch(() => undefined);
    });

    expect(toast.error).toHaveBeenCalledWith('Steward must be an active user');
  });
});
