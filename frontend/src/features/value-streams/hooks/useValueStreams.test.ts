import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useValueStreams, useValueStreamsQuery, useValueStream } from './useValueStreams';
import { valueStreamsQueryKeys } from '../queryKeys';
import type { ValueStream, ValueStreamId, ValueStreamDetail, ValueStreamsResponse } from '../../../api/types';

vi.mock('../api', () => ({
  valueStreamsApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { valueStreamsApi } from '../api';

const createValueStream = (id: string, name: string, description = ''): ValueStream => ({
  id: id as ValueStreamId,
  name,
  description,
  stageCount: 0,
  createdAt: '2024-01-01T00:00:00Z',
  _links: {
    self: { href: `/api/v1/value-streams/${id}`, method: 'GET' },
    edit: { href: `/api/v1/value-streams/${id}`, method: 'PUT' },
    delete: { href: `/api/v1/value-streams/${id}`, method: 'DELETE' },
    collection: { href: '/api/v1/value-streams', method: 'GET' },
  },
});

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useValueStreams', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useValueStreamsQuery', () => {
    it('should fetch value streams', async () => {
      const response: ValueStreamsResponse = {
        data: [createValueStream('vs-1', 'Customer Onboarding')],
        _links: {
          self: { href: '/api/v1/value-streams', method: 'GET' },
          'x-create': { href: '/api/v1/value-streams', method: 'POST' },
        },
      };
      vi.mocked(valueStreamsApi.getAll).mockResolvedValue(response);

      const { result } = renderHook(() => useValueStreamsQuery(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(valueStreamsApi.getAll).toHaveBeenCalled();
      expect(result.current.data).toEqual(response);
    });
  });

  describe('useValueStream', () => {
    it('should fetch a single value stream by id', async () => {
      const vs: ValueStreamDetail = { ...createValueStream('vs-1', 'Customer Onboarding'), stages: [], stageCapabilities: [] };
      vi.mocked(valueStreamsApi.getById).mockResolvedValue(vs);

      const { result } = renderHook(() => useValueStream('vs-1' as ValueStreamId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(valueStreamsApi.getById).toHaveBeenCalledWith('vs-1');
      expect(result.current.data).toEqual(vs);
    });

    it('should not fetch when id is undefined', async () => {
      const { result } = renderHook(() => useValueStream(undefined), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(valueStreamsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe('useValueStreams composite hook', () => {
    it('should return value streams from query', async () => {
      const streams = [createValueStream('vs-1', 'Customer Onboarding')];
      const response: ValueStreamsResponse = {
        data: streams,
        _links: { self: { href: '/api/v1/value-streams', method: 'GET' } },
      };
      vi.mocked(valueStreamsApi.getAll).mockResolvedValue(response);

      const { result } = renderHook(() => useValueStreams(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.valueStreams).toEqual(streams);
      expect(result.current.collectionLinks).toEqual(response._links);
    });

    it('should return empty array when no data', async () => {
      const response: ValueStreamsResponse = {
        data: [],
        _links: { self: { href: '/api/v1/value-streams', method: 'GET' } },
      };
      vi.mocked(valueStreamsApi.getAll).mockResolvedValue(response);

      const { result } = renderHook(() => useValueStreams(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.valueStreams).toEqual([]);
    });

    it.each([
      {
        scenario: 'create',
        seedData: [] as ValueStream[],
        setupMock: () => {
          vi.mocked(valueStreamsApi.create).mockResolvedValue(createValueStream('vs-new', 'New Stream'));
        },
        performAction: async (hook: ReturnType<typeof useValueStreams>) => {
          await hook.createValueStream('New Stream', 'Description');
        },
        assertApi: () => {
          expect(valueStreamsApi.create).toHaveBeenCalledWith({ name: 'New Stream', description: 'Description' });
        },
      },
      {
        scenario: 'delete',
        seedData: [createValueStream('vs-1', 'To Delete')],
        setupMock: () => {
          vi.mocked(valueStreamsApi.delete).mockResolvedValue(undefined);
        },
        performAction: async (hook: ReturnType<typeof useValueStreams>) => {
          await hook.deleteValueStream(createValueStream('vs-1', 'To Delete'));
        },
        assertApi: () => {
          expect(valueStreamsApi.delete).toHaveBeenCalledWith(createValueStream('vs-1', 'To Delete'));
        },
      },
    ])('should $scenario a value stream and invalidate queries', async ({ seedData, setupMock, performAction, assertApi }) => {
      const response: ValueStreamsResponse = {
        data: seedData,
        _links: { self: { href: '/api/v1/value-streams', method: 'GET' } },
      };
      vi.mocked(valueStreamsApi.getAll).mockResolvedValue(response);
      setupMock();

      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useValueStreams(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      await act(async () => {
        await performAction(result.current);
      });

      assertApi();
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: valueStreamsQueryKeys.lists(),
      });
    });
  });
});
