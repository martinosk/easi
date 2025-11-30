import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useGridPositions } from './useGridPositions';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, ViewId } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getViews: vi.fn(),
    createView: vi.fn(),
    getViewById: vi.fn(),
    addCapabilityToView: vi.fn(),
    updateCapabilityPositionInView: vi.fn(),
  },
}));

describe('useGridPositions', () => {
  const domainId = 'bd-finance' as BusinessDomainId;
  const viewId = 'view-123' as ViewId;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('view initialization', () => {
    it('should find existing view for domain', async () => {
      const existingView = {
        id: viewId,
        name: 'bd-finance Domain Layout',
        viewType: 'businessDomain',
        capabilities: [
          { capabilityId: 'cap-1', x: 0, y: 0 },
          { capabilityId: 'cap-2', x: 1, y: 0 },
        ],
        isDefault: false,
        components: [],
        createdAt: '2024-01-01',
        _links: { self: '/api/v1/views/view-123' },
      };

      vi.mocked(apiClient.getViews).mockResolvedValue([existingView]);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.viewId).toBe(viewId);
      expect(result.current.positions).toEqual({
        'cap-1': { x: 0, y: 0 },
        'cap-2': { x: 1, y: 0 },
      });
    });

    it('should create new view if none exists for domain', async () => {
      const newView = {
        id: 'view-new' as ViewId,
        name: 'bd-finance Domain Layout',
        viewType: 'businessDomain',
        capabilities: [],
        isDefault: false,
        components: [],
        createdAt: '2024-01-01',
        _links: { self: '/api/v1/views/view-new' },
      };

      vi.mocked(apiClient.getViews).mockResolvedValue([]);
      vi.mocked(apiClient.createView).mockResolvedValue(newView);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.createView).toHaveBeenCalledWith({
        name: 'bd-finance Domain Layout',
        description: 'Grid layout for business domain bd-finance',
      });
      expect(result.current.viewId).toBe('view-new');
    });

    it('should not initialize without domainId', () => {
      const { result } = renderHook(() => useGridPositions(null));

      expect(result.current.viewId).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe('position updates', () => {
    it('should save position when capability is moved', async () => {
      const existingView = {
        id: viewId,
        name: 'bd-finance Domain Layout',
        capabilities: [{ capabilityId: 'cap-1', x: 0, y: 0 }],
        isDefault: false,
        components: [],
        createdAt: '2024-01-01',
        _links: { self: '/api/v1/views/view-123' },
      };

      vi.mocked(apiClient.getViews).mockResolvedValue([existingView]);
      vi.mocked(apiClient.updateCapabilityPositionInView).mockResolvedValue();

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.viewId).toBe(viewId);
      });

      await act(async () => {
        await result.current.updatePosition('cap-1' as any, 2, 1);
      });

      expect(apiClient.updateCapabilityPositionInView).toHaveBeenCalledWith(
        viewId,
        'cap-1',
        { x: 2, y: 1 }
      );
    });

    it('should add capability to view if not present', async () => {
      const existingView = {
        id: viewId,
        name: 'bd-finance Domain Layout',
        capabilities: [],
        isDefault: false,
        components: [],
        createdAt: '2024-01-01',
        _links: { self: '/api/v1/views/view-123' },
      };

      vi.mocked(apiClient.getViews).mockResolvedValue([existingView]);
      vi.mocked(apiClient.addCapabilityToView).mockResolvedValue();

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.viewId).toBe(viewId);
      });

      await act(async () => {
        await result.current.updatePosition('cap-new' as any, 0, 0);
      });

      expect(apiClient.addCapabilityToView).toHaveBeenCalledWith(viewId, {
        capabilityId: 'cap-new',
        x: 0,
        y: 0,
      });
    });
  });

  describe('getPositionForCapability', () => {
    it('should return stored position if exists', async () => {
      const existingView = {
        id: viewId,
        name: 'bd-finance Domain Layout',
        capabilities: [{ capabilityId: 'cap-1', x: 3, y: 2 }],
        isDefault: false,
        components: [],
        createdAt: '2024-01-01',
        _links: { self: '/api/v1/views/view-123' },
      };

      vi.mocked(apiClient.getViews).mockResolvedValue([existingView]);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.viewId).toBe(viewId);
      });

      expect(result.current.getPositionForCapability('cap-1' as any)).toEqual({ x: 3, y: 2 });
    });

    it('should return null for unknown capability', async () => {
      const existingView = {
        id: viewId,
        name: 'bd-finance Domain Layout',
        capabilities: [],
        isDefault: false,
        components: [],
        createdAt: '2024-01-01',
        _links: { self: '/api/v1/views/view-123' },
      };

      vi.mocked(apiClient.getViews).mockResolvedValue([existingView]);

      const { result } = renderHook(() => useGridPositions(domainId));

      await waitFor(() => {
        expect(result.current.viewId).toBe(viewId);
      });

      expect(result.current.getPositionForCapability('cap-unknown' as any)).toBeNull();
    });
  });
});
