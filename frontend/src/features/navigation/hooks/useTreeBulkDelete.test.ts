import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTreeBulkDelete } from './useTreeBulkDelete';
import type { TreeSelectedItem } from './useTreeMultiSelect';
import type { HATEOASLinks, Component, ComponentId, Capability, CapabilityId } from '../../../api/types';


const deleteLinks: HATEOASLinks = {
  self: { href: '/test', method: 'GET' },
  delete: { href: '/test', method: 'DELETE' },
};

function makeComponent(id: string, name: string): Component {
  return {
    id: id as ComponentId,
    name,
    createdAt: '2024-01-01',
    _links: deleteLinks,
  };
}

function makeCapability(id: string, name: string): Capability {
  return {
    id: id as CapabilityId,
    name,
    level: 'L1',
    maturityLevel: 'Initial',
    createdAt: '2024-01-01',
    _links: deleteLinks,
  };
}

function makeSelectedItem(id: string, type: TreeSelectedItem['type'], name: string): TreeSelectedItem {
  return { id, name, type, links: deleteLinks };
}

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: vi.fn(() => ({ data: [makeComponent('comp-1', 'App A'), makeComponent('comp-2', 'App B')] })),
  useDeleteComponent: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue(undefined),
  })),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: vi.fn(() => ({ data: [makeCapability('cap-1', 'Cap A')] })),
  useDeleteCapability: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue(undefined),
  })),
}));

vi.mock('../../origin-entities/hooks/useAcquiredEntities', () => ({
  useAcquiredEntitiesQuery: vi.fn(() => ({ data: [] })),
  useDeleteAcquiredEntity: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue(undefined),
  })),
}));

vi.mock('../../origin-entities/hooks/useVendors', () => ({
  useVendorsQuery: vi.fn(() => ({ data: [] })),
  useDeleteVendor: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue(undefined),
  })),
}));

vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: vi.fn(() => ({ data: [] })),
  useDeleteInternalTeam: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue(undefined),
  })),
}));

describe('useTreeBulkDelete', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('requestBulkDelete sets bulk items', () => {
    const { result } = renderHook(() => useTreeBulkDelete());

    const items = [
      makeSelectedItem('comp-1', 'component', 'App A'),
      makeSelectedItem('comp-2', 'component', 'App B'),
    ];

    act(() => {
      result.current.requestBulkDelete(items);
    });

    expect(result.current.bulkItems).toEqual(items);
    expect(result.current.itemNames).toEqual(['App A', 'App B']);
  });

  it('handleCancel clears bulk items', () => {
    const { result } = renderHook(() => useTreeBulkDelete());

    act(() => {
      result.current.requestBulkDelete([makeSelectedItem('comp-1', 'component', 'App A')]);
    });

    act(() => {
      result.current.handleCancel();
    });

    expect(result.current.bulkItems).toBeNull();
  });

  it('successful bulk delete clears bulk items', async () => {
    const { result } = renderHook(() => useTreeBulkDelete());

    const items = [
      makeSelectedItem('comp-1', 'component', 'App A'),
      makeSelectedItem('comp-2', 'component', 'App B'),
    ];

    act(() => {
      result.current.requestBulkDelete(items);
    });

    await act(async () => {
      await result.current.handleConfirm();
    });

    expect(result.current.bulkItems).toBeNull();
    expect(result.current.isExecuting).toBe(false);
  });

  it('partial failure stops execution and reports results', async () => {
    const { useDeleteComponent } = await import('../../components/hooks/useComponents');
    let callCount = 0;
    (useDeleteComponent as ReturnType<typeof vi.fn>).mockReturnValue({
      mutateAsync: vi.fn().mockImplementation(() => {
        callCount++;
        if (callCount === 2) {
          return Promise.reject(new Error('Delete failed'));
        }
        return Promise.resolve();
      }),
    });

    const { result } = renderHook(() => useTreeBulkDelete());

    const items = [
      makeSelectedItem('comp-1', 'component', 'App A'),
      makeSelectedItem('comp-2', 'component', 'App B'),
      makeSelectedItem('comp-3', 'component', 'App C'),
    ];

    act(() => {
      result.current.requestBulkDelete(items);
    });

    await act(async () => {
      await result.current.handleConfirm();
    });

    expect(result.current.result).not.toBeNull();
    expect(result.current.result?.succeeded).toContain('App A');
    expect(result.current.result?.failed).toHaveLength(1);
    expect(result.current.result?.failed[0].name).toBe('App B');
    expect(result.current.bulkItems).not.toBeNull();
  });
});
