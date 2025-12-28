import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useLayout } from './useLayout';
import { apiClient } from '../api/client';
import type { LayoutContainer, LayoutContainerId } from '../api/types';

vi.mock('../api/client', () => ({
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
  contextRef: 'domain-finance',
  preferences: { colorScheme: 'default' },
  elements: [
    {
      elementId: 'cap-1',
      x: 100,
      y: 200,
      _links: {
        self: { href: '/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-1' },
      },
    },
  ],
  version: 1,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  _links: {
    self: { href: '/api/v1/layouts/business-domain-grid/domain-finance' },
  },
};

describe('useLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should load existing layout', async () => {
    vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);

    const { result } = renderHook(() =>
      useLayout('business-domain-grid', 'domain-finance')
    );

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.layout).toEqual(mockLayout);
    expect(result.current.positions).toEqual({ 'cap-1': { x: 100, y: 200 } });
    expect(result.current.error).toBeNull();
  });

  it('should create layout if not found', async () => {
    vi.mocked(apiClient.getLayout).mockResolvedValue(null);
    vi.mocked(apiClient.upsertLayout).mockResolvedValue(mockLayout);

    const { result } = renderHook(() =>
      useLayout('business-domain-grid', 'domain-finance')
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(apiClient.upsertLayout).toHaveBeenCalledWith(
      'business-domain-grid',
      'domain-finance',
      {}
    );
    expect(result.current.layout).toEqual(mockLayout);
  });

  it('should update element position optimistically', async () => {
    vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);
    vi.mocked(apiClient.upsertElementPosition).mockResolvedValue({
      elementId: 'cap-1',
      x: 300,
      y: 400,
      _links: { self: { href: '/test' } },
    });

    const { result } = renderHook(() =>
      useLayout('business-domain-grid', 'domain-finance')
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    await act(async () => {
      await result.current.updateElementPosition('cap-1', 300, 400);
    });

    expect(result.current.positions['cap-1']).toEqual({ x: 300, y: 400 });
    expect(apiClient.upsertElementPosition).toHaveBeenCalledWith(
      'business-domain-grid',
      'domain-finance',
      'cap-1',
      { x: 300, y: 400 }
    );
  });

  it('should rollback on API error', async () => {
    vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);
    vi.mocked(apiClient.upsertElementPosition).mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() =>
      useLayout('business-domain-grid', 'domain-finance')
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    await act(async () => {
      try {
        await result.current.updateElementPosition('cap-1', 500, 600);
      } catch {
        // Expected to throw
      }
    });

    expect(result.current.positions['cap-1']).toEqual({ x: 100, y: 200 });
  });

  it('should batch update positions', async () => {
    vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);
    vi.mocked(apiClient.batchUpdateElements).mockResolvedValue({
      updated: 2,
      elements: [],
      _links: {},
    });

    const { result } = renderHook(() =>
      useLayout('business-domain-grid', 'domain-finance')
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    await act(async () => {
      await result.current.batchUpdatePositions([
        { elementId: 'cap-1', x: 150, y: 250 },
        { elementId: 'cap-2', x: 350, y: 450 },
      ]);
    });

    expect(apiClient.batchUpdateElements).toHaveBeenCalledWith(
      'business-domain-grid',
      'domain-finance',
      [
        { elementId: 'cap-1', x: 150, y: 250 },
        { elementId: 'cap-2', x: 350, y: 450 },
      ]
    );
  });

  it('should handle null contextRef', async () => {
    const { result } = renderHook(() =>
      useLayout('business-domain-grid', null)
    );

    expect(result.current.isLoading).toBe(false);
    expect(result.current.layout).toBeNull();
    expect(apiClient.getLayout).not.toHaveBeenCalled();
  });

  it('should refetch layout', async () => {
    vi.mocked(apiClient.getLayout).mockResolvedValue(mockLayout);

    const { result } = renderHook(() =>
      useLayout('business-domain-grid', 'domain-finance')
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    vi.mocked(apiClient.getLayout).mockClear();

    await act(async () => {
      await result.current.refetch();
    });

    expect(apiClient.getLayout).toHaveBeenCalledWith(
      'business-domain-grid',
      'domain-finance'
    );
  });
});
