import React from 'react';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useCapabilityRealizations } from './useCapabilityRealizations';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, CapabilityId, CapabilityRealization, CapabilityRealizationsGroup, RealizationId, ComponentId } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getCapabilityRealizationsByDomain: vi.fn(),
  },
}));

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });
}

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const createRealization = (
  id: string,
  capabilityId: string,
  origin: 'Direct' | 'Inherited' = 'Direct',
  sourceCapabilityId?: string,
  componentId = 'comp-1'
): CapabilityRealization => ({
  id: id as RealizationId,
  capabilityId: capabilityId as CapabilityId,
  componentId: componentId as ComponentId,
  componentName: 'Test Component',
  realizationLevel: 'Full',
  origin,
  sourceCapabilityId: sourceCapabilityId as CapabilityId | undefined,
  linkedAt: '2024-01-01T00:00:00Z',
  _links: {},
});

const createGroup = (
  capabilityId: string,
  level: 'L1' | 'L2' | 'L3' | 'L4',
  realizations: CapabilityRealization[]
): CapabilityRealizationsGroup => ({
  capabilityId: capabilityId as CapabilityId,
  capabilityName: `Capability ${capabilityId}`,
  level,
  realizations,
});

const domainId = 'domain-1' as BusinessDomainId;

describe('useCapabilityRealizations', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = createQueryClient();
    vi.clearAllMocks();
  });

  describe('fetching realizations', () => {
    it('should fetch realizations for domain when enabled', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledWith(domainId, 4);
      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.error).toBeNull();
    });

    it('should pass depth parameter to API', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([]);

      renderHook(() => useCapabilityRealizations(true, domainId, 2),
        { wrapper: createWrapper(queryClient) });

      await waitFor(() => {
        expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledWith(domainId, 2);
      });
    });

    it('should not fetch when disabled', async () => {
      const { result } = renderHook(() =>
        useCapabilityRealizations(false, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      expect(result.current.isLoading).toBe(false);
      expect(result.current.realizations).toEqual([]);
      expect(apiClient.getCapabilityRealizationsByDomain).not.toHaveBeenCalled();
    });

    it('should not fetch when domainId is null', async () => {
      const { result } = renderHook(() =>
        useCapabilityRealizations(true, null, 4),
        { wrapper: createWrapper(queryClient) }
      );

      expect(result.current.isLoading).toBe(false);
      expect(result.current.realizations).toEqual([]);
      expect(apiClient.getCapabilityRealizationsByDomain).not.toHaveBeenCalled();
    });

    it('should re-fetch when depth changes', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([]);

      const { rerender } = renderHook(
        ({ depth }) => useCapabilityRealizations(true, domainId, depth),
        { initialProps: { depth: 2 }, wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledWith(domainId, 2);
      });

      rerender({ depth: 4 });

      await waitFor(() => {
        expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledWith(domainId, 4);
      });

      expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledTimes(2);
    });

    it('should handle API errors', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockRejectedValue(
        new Error('API Error')
      );

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.error).toBeInstanceOf(Error);
      expect(result.current.error?.message).toBe('API Error');
    });
  });

  describe('inherited realization visibility', () => {
    it('should show inherited realization when source capability is NOT visible', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-1', 'cap-l1', 'Inherited', 'cap-l3'),
        ]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 1),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].origin).toBe('Inherited');
    });

    it('should show inherited when source capability is not rendered', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2'),
        ]),
        createGroup('cap-l2', 'L2', [
          createRealization('real-direct', 'cap-l2', 'Direct'),
        ]),
      ]);

      const visibleIds = new Set<CapabilityId>(['cap-l1' as CapabilityId]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 2, visibleIds),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-inherited');
      expect(result.current.realizations[0].origin).toBe('Inherited');
    });


    it('should hide inherited realization when source capability IS visible', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2'),
        ]),
        createGroup('cap-l2', 'L2', [
          createRealization('real-direct', 'cap-l2', 'Direct'),
        ]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 2),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-direct');
      expect(result.current.realizations[0].origin).toBe('Direct');
    });

    it('should always show direct realizations', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
        createGroup('cap-2', 'L1', [createRealization('real-2', 'cap-2', 'Direct')]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(2);
      expect(result.current.realizations.every((r) => r.origin === 'Direct')).toBe(true);
    });
  });

  describe('getRealizationsForCapability', () => {
    it('should return realizations for a specific capability', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
        createGroup('cap-2', 'L1', [createRealization('real-2', 'cap-2', 'Direct')]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      const cap1Realizations = result.current.getRealizationsForCapability('cap-1' as CapabilityId);
      expect(cap1Realizations).toHaveLength(1);
      expect(cap1Realizations[0].capabilityId).toBe('cap-1');
    });

    it('should return empty array for capability with no realizations', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      const noRealizations = result.current.getRealizationsForCapability('cap-nonexistent' as CapabilityId);
      expect(noRealizations).toEqual([]);
    });
  });

  describe('depth-based visibility scenarios', () => {
    it('should show inherited on L1 when L2 is not visible (L1 only depth)', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2'),
        ]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 1),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-inherited');
    });

    it('should show direct on L2 and hide inherited on L1 when both visible', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2'),
        ]),
        createGroup('cap-l2', 'L2', [
          createRealization('real-direct', 'cap-l2', 'Direct'),
        ]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 2),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-direct');
      expect(result.current.realizations[0].capabilityId).toBe('cap-l2');
    });

    it('should show inherited only at deepest visible level when source is below visible depth', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-inherited-l1', 'cap-l1', 'Inherited', 'cap-l3'),
        ]),
        createGroup('cap-l2', 'L2', [
          createRealization('real-inherited-l2', 'cap-l2', 'Inherited', 'cap-l3'),
        ]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 2),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-inherited-l2');
      expect(result.current.realizations[0].capabilityId).toBe('cap-l2');
    });

    it('should handle multiple components with inherited realizations at different levels', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-l1', 'L1', [
          createRealization('real-comp1-l1', 'cap-l1', 'Inherited', 'cap-l3', 'comp-1'),
          createRealization('real-comp2-l1', 'cap-l1', 'Inherited', 'cap-l3', 'comp-2'),
        ]),
        createGroup('cap-l2', 'L2', [
          createRealization('real-comp1-l2', 'cap-l2', 'Inherited', 'cap-l3', 'comp-1'),
          createRealization('real-comp2-l2', 'cap-l2', 'Inherited', 'cap-l3', 'comp-2'),
        ]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 2),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(2);
      expect(result.current.realizations.every((r) => r.capabilityId === 'cap-l2')).toBe(true);
    });
  });

  describe('refetch', () => {
    it('should re-fetch realizations when refetch is called', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledTimes(1);

      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
        createGroup('cap-2', 'L1', [createRealization('real-2', 'cap-2', 'Direct')]),
      ]);

      await act(async () => {
        await result.current.refetch();
      });

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(2);
      });

      expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledTimes(2);
    });

    it('should return empty realizations when disabled initially', async () => {
      const { result } = renderHook(() =>
        useCapabilityRealizations(false, domainId, 4),
        { wrapper: createWrapper(queryClient) }
      );

      expect(result.current.isLoading).toBe(false);
      expect(result.current.realizations).toEqual([]);
      expect(apiClient.getCapabilityRealizationsByDomain).not.toHaveBeenCalled();
    });
  });

  describe('cache behavior', () => {
    it('should keep cached data when disabled after fetching', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
      ]);

      const { result, rerender } = renderHook(
        ({ enabled }) => useCapabilityRealizations(enabled, domainId, 4),
        { initialProps: { enabled: true }, wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(1);
      });

      rerender({ enabled: false });

      expect(result.current.realizations).toHaveLength(1);
    });

    it('should return empty when domainId changes to null', async () => {
      vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([
        createGroup('cap-1', 'L1', [createRealization('real-1', 'cap-1', 'Direct')]),
      ]);

      const { result, rerender } = renderHook(
        ({ domainId }) => useCapabilityRealizations(true, domainId, 4),
        { initialProps: { domainId: domainId as BusinessDomainId | null }, wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(1);
      });

      rerender({ domainId: null });

      expect(result.current.realizations).toEqual([]);
    });
  });
});
