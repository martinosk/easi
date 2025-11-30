import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useGridPositions } from './useGridPositions';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, LayoutContainer, LayoutContainerId } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getLayout: vi.fn(),
    upsertLayout: vi.fn(),
    upsertElementPosition: vi.fn(),
    batchUpdateElements: vi.fn(),
    updateLayoutPreferences: vi.fn(),
  },
}));

const mockLayout: LayoutContainer = {
  id: 'layout-123' as LayoutContainerId,
  contextType: 'business-domain-grid',
  contextRef: 'bd-finance',
  preferences: {},
  elements: [
    { elementId: 'cap-1', x: 0, y: 0, _links: {} },
    { elementId: 'cap-2', x: 1, y: 0, _links: {} },
  ],
  version: 1,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  _links: {},
};

describe('useGridPositions', () => {
  const domainId = 'bd-finance' as BusinessDomainId;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('layout initialization', () => {
    it('should find existing layout for domain', async () => {
      vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.positions).toEqual({
        'cap-1': { x: 0, y: 0 },
        'cap-2': { x: 1, y: 0 },
      });
      expect(apiClient.getLayout).toHaveBeenCalledWith('business-domain-grid', 'bd-finance');
    });

    it('should create new layout if none exists for domain', async () => {
      const newLayout: LayoutContainer = {
        ...mockLayout,
        elements: [],
      };

      vi.mocked(apiClient.getLayout).mockResolvedValue(null);
      vi.mocked(apiClient.upsertLayout).mockResolvedValue(newLayout);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.upsertLayout).toHaveBeenCalledWith(
        'business-domain-grid',
        'bd-finance',
        {}
      );
      expect(result.current.positions).toEqual({});
    });

    it('should not initialize without domainId', () => {
      const { result } = renderHook(() => useGridPositions(null));

      expect(result.current.isLoading).toBe(false);
      expect(apiClient.getLayout).not.toHaveBeenCalled();
    });
  });

  describe('position updates', () => {
    it('should save position when capability is moved', async () => {
      vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);
      vi.mocked(apiClient.upsertElementPosition).mockResolvedValue({
        elementId: 'cap-1',
        x: 2,
        y: 1,
        _links: {},
      });

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      await act(async () => {
        await result.current.updatePosition('cap-1' as any, 2, 1);
      });

      expect(apiClient.upsertElementPosition).toHaveBeenCalledWith(
        'business-domain-grid',
        'bd-finance',
        'cap-1',
        { x: 2, y: 1 }
      );
      expect(result.current.positions['cap-1']).toEqual({ x: 2, y: 1 });
    });

    it('should add new capability position', async () => {
      vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);
      vi.mocked(apiClient.upsertElementPosition).mockResolvedValue({
        elementId: 'cap-new',
        x: 0,
        y: 0,
        _links: {},
      });

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      await act(async () => {
        await result.current.updatePosition('cap-new' as any, 0, 0);
      });

      expect(apiClient.upsertElementPosition).toHaveBeenCalledWith(
        'business-domain-grid',
        'bd-finance',
        'cap-new',
        { x: 0, y: 0 }
      );
      expect(result.current.positions['cap-new']).toEqual({ x: 0, y: 0 });
    });

    it('should rollback on API error', async () => {
      vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);
      vi.mocked(apiClient.upsertElementPosition).mockRejectedValue(new Error('API Error'));

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      await act(async () => {
        try {
          await result.current.updatePosition('cap-1' as any, 500, 600);
        } catch {
          // Expected to throw
        }
      });

      expect(result.current.positions['cap-1']).toEqual({ x: 0, y: 0 });
    });
  });

  describe('getPositionForCapability', () => {
    it('should return stored position if exists', async () => {
      const layoutWithPosition: LayoutContainer = {
        ...mockLayout,
        elements: [{ elementId: 'cap-1', x: 3, y: 2, _links: {} }],
      };

      vi.mocked(apiClient.getLayout).mockResolvedValue(layoutWithPosition);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.getPositionForCapability('cap-1' as any)).toEqual({ x: 3, y: 2 });
    });

    it('should return null for unknown capability', async () => {
      vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.getPositionForCapability('cap-unknown' as any)).toBeNull();
    });
  });
});
