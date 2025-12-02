import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useCapabilityRealizations, VisibleCapability } from './useCapabilityRealizations';
import { apiClient } from '../../../api/client';
import type { CapabilityId, CapabilityLevel, CapabilityRealization, RealizationId, ComponentId } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getSystemsByCapability: vi.fn(),
  },
}));

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

const cap = (id: string, level: CapabilityLevel): VisibleCapability => ({
  id: id as CapabilityId,
  level,
});

describe('useCapabilityRealizations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('fetching realizations', () => {
    it('should fetch realizations for each visible capability when enabled', async () => {
      vi.mocked(apiClient.getSystemsByCapability).mockResolvedValue([
        createRealization('real-1', 'cap-1', 'Direct'),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [cap('cap-1', 'L1')])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.getSystemsByCapability).toHaveBeenCalledWith('cap-1');
      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.error).toBeNull();
    });

    it('should fetch for multiple capabilities in parallel', async () => {
      vi.mocked(apiClient.getSystemsByCapability).mockImplementation((capabilityId) => {
        if (capabilityId === 'cap-1') {
          return Promise.resolve([createRealization('real-1', 'cap-1', 'Direct')]);
        }
        return Promise.resolve([createRealization('real-2', 'cap-2', 'Direct')]);
      });

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [cap('cap-1', 'L1'), cap('cap-2', 'L1')])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.getSystemsByCapability).toHaveBeenCalledWith('cap-1');
      expect(apiClient.getSystemsByCapability).toHaveBeenCalledWith('cap-2');
      expect(result.current.realizations).toHaveLength(2);
    });

    it('should not fetch when disabled', async () => {
      const { result } = renderHook(() =>
        useCapabilityRealizations(false, [cap('cap-1', 'L1')])
      );

      expect(result.current.isLoading).toBe(false);
      expect(result.current.realizations).toEqual([]);
      expect(apiClient.getSystemsByCapability).not.toHaveBeenCalled();
    });

    it('should not fetch when no capabilities are visible', async () => {
      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [])
      );

      expect(result.current.isLoading).toBe(false);
      expect(result.current.realizations).toEqual([]);
      expect(apiClient.getSystemsByCapability).not.toHaveBeenCalled();
    });

    it('should cache fetched capabilities and not re-fetch', async () => {
      vi.mocked(apiClient.getSystemsByCapability).mockResolvedValue([
        createRealization('real-1', 'cap-1', 'Direct'),
      ]);

      const { result, rerender } = renderHook(
        ({ caps }) => useCapabilityRealizations(true, caps),
        { initialProps: { caps: [cap('cap-1', 'L1')] } }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.getSystemsByCapability).toHaveBeenCalledTimes(1);

      rerender({ caps: [cap('cap-1', 'L1')] });

      expect(apiClient.getSystemsByCapability).toHaveBeenCalledTimes(1);
    });

    it('should only fetch new capabilities when list expands', async () => {
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([createRealization('real-1', 'cap-1', 'Direct')])
        .mockResolvedValueOnce([createRealization('real-2', 'cap-2', 'Direct')]);

      const { result, rerender } = renderHook(
        ({ caps }) => useCapabilityRealizations(true, caps),
        { initialProps: { caps: [cap('cap-1', 'L1')] } }
      );

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(1);
      });

      rerender({ caps: [cap('cap-1', 'L1'), cap('cap-2', 'L1')] });

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(2);
      });

      expect(apiClient.getSystemsByCapability).toHaveBeenCalledTimes(2);
      expect(apiClient.getSystemsByCapability).toHaveBeenNthCalledWith(1, 'cap-1');
      expect(apiClient.getSystemsByCapability).toHaveBeenNthCalledWith(2, 'cap-2');
    });

    it('should handle API errors', async () => {
      vi.mocked(apiClient.getSystemsByCapability).mockRejectedValue(
        new Error('API Error')
      );

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [cap('cap-1', 'L1')])
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
      vi.mocked(apiClient.getSystemsByCapability).mockResolvedValue([
        createRealization('real-1', 'cap-parent', 'Inherited', 'cap-child'),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [cap('cap-parent', 'L1')])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].origin).toBe('Inherited');
    });

    it('should hide inherited realization when source capability IS visible', async () => {
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([createRealization('real-inherited', 'cap-parent', 'Inherited', 'cap-child')])
        .mockResolvedValueOnce([createRealization('real-direct', 'cap-child', 'Direct')]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [
          cap('cap-parent', 'L1'),
          cap('cap-child', 'L2'),
        ])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-direct');
      expect(result.current.realizations[0].origin).toBe('Direct');
    });

    it('should always show direct realizations', async () => {
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([createRealization('real-1', 'cap-1', 'Direct')])
        .mockResolvedValueOnce([createRealization('real-2', 'cap-2', 'Direct')]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [
          cap('cap-1', 'L1'),
          cap('cap-2', 'L1'),
        ])
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
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([createRealization('real-1', 'cap-1', 'Direct')])
        .mockResolvedValueOnce([createRealization('real-2', 'cap-2', 'Direct')]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [
          cap('cap-1', 'L1'),
          cap('cap-2', 'L1'),
        ])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      const cap1Realizations = result.current.getRealizationsForCapability('cap-1' as CapabilityId);
      expect(cap1Realizations).toHaveLength(1);
      expect(cap1Realizations[0].capabilityId).toBe('cap-1');
    });

    it('should return empty array for capability with no realizations', async () => {
      vi.mocked(apiClient.getSystemsByCapability).mockResolvedValue([]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [cap('cap-1', 'L1')])
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
      vi.mocked(apiClient.getSystemsByCapability).mockResolvedValue([
        createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2'),
      ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [cap('cap-l1', 'L1')])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-inherited');
    });

    it('should show direct on L2 and hide inherited on L1 when both visible (L1-L2 depth)', async () => {
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2')])
        .mockResolvedValueOnce([createRealization('real-direct', 'cap-l2', 'Direct')]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [
          cap('cap-l1', 'L1'),
          cap('cap-l2', 'L2'),
        ])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-direct');
      expect(result.current.realizations[0].capabilityId).toBe('cap-l2');
    });

    it('should show inherited only at deepest visible level when source is below visible depth', async () => {
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([createRealization('real-inherited-l1', 'cap-l1', 'Inherited', 'cap-l3')])
        .mockResolvedValueOnce([createRealization('real-inherited-l2', 'cap-l2', 'Inherited', 'cap-l3')]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [
          cap('cap-l1', 'L1'),
          cap('cap-l2', 'L2'),
        ])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-inherited-l2');
      expect(result.current.realizations[0].capabilityId).toBe('cap-l2');
    });

    it('should handle multiple components with inherited realizations at different levels', async () => {
      vi.mocked(apiClient.getSystemsByCapability)
        .mockResolvedValueOnce([
          createRealization('real-comp1-l1', 'cap-l1', 'Inherited', 'cap-l3', 'comp-1'),
          createRealization('real-comp2-l1', 'cap-l1', 'Inherited', 'cap-l3', 'comp-2'),
        ])
        .mockResolvedValueOnce([
          createRealization('real-comp1-l2', 'cap-l2', 'Inherited', 'cap-l3', 'comp-1'),
          createRealization('real-comp2-l2', 'cap-l2', 'Inherited', 'cap-l3', 'comp-2'),
        ]);

      const { result } = renderHook(() =>
        useCapabilityRealizations(true, [
          cap('cap-l1', 'L1'),
          cap('cap-l2', 'L2'),
        ])
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.realizations).toHaveLength(2);
      expect(result.current.realizations.every((r) => r.capabilityId === 'cap-l2')).toBe(true);
    });
  });

  describe('clearing state', () => {
    it('should clear realizations when disabled', async () => {
      vi.mocked(apiClient.getSystemsByCapability).mockResolvedValue([
        createRealization('real-1', 'cap-1', 'Direct'),
      ]);

      const { result, rerender } = renderHook(
        ({ enabled }) => useCapabilityRealizations(enabled, [cap('cap-1', 'L1')]),
        { initialProps: { enabled: true } }
      );

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(1);
      });

      rerender({ enabled: false });

      expect(result.current.realizations).toEqual([]);
    });
  });

  describe('depth change race conditions', () => {
    it('should keep inherited on L1 visible until L2 realizations are fetched', async () => {
      let resolveL2Fetch: (value: CapabilityRealization[]) => void;
      const l2FetchPromise = new Promise<CapabilityRealization[]>((resolve) => {
        resolveL2Fetch = resolve;
      });

      vi.mocked(apiClient.getSystemsByCapability).mockImplementation((capabilityId) => {
        if (capabilityId === 'cap-l1') {
          return Promise.resolve([createRealization('real-inherited', 'cap-l1', 'Inherited', 'cap-l2')]);
        }
        return l2FetchPromise;
      });

      const { result, rerender } = renderHook(
        ({ caps }) => useCapabilityRealizations(true, caps),
        { initialProps: { caps: [cap('cap-l1', 'L1')] } }
      );

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(1);
      });

      expect(result.current.realizations[0].id).toBe('real-inherited');
      expect(result.current.realizations[0].capabilityId).toBe('cap-l1');

      rerender({ caps: [cap('cap-l1', 'L1'), cap('cap-l2', 'L2')] });

      expect(result.current.realizations).toHaveLength(1);
      expect(result.current.realizations[0].id).toBe('real-inherited');

      resolveL2Fetch!([createRealization('real-direct', 'cap-l2', 'Direct')]);

      await waitFor(() => {
        expect(result.current.realizations).toHaveLength(1);
        expect(result.current.realizations[0].id).toBe('real-direct');
      });
    });
  });
});
