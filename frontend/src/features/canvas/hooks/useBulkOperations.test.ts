import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useBulkOperations } from './useBulkOperations';
import type { NodeContextMenu } from './useNodeContextMenu';

const mockRemoveComponent = vi.fn().mockResolvedValue(undefined);
const mockRemoveCapability = vi.fn().mockResolvedValue(undefined);
const mockRemoveOriginEntity = vi.fn().mockResolvedValue(undefined);
const mockDeleteComponent = vi.fn().mockResolvedValue(undefined);
const mockDeleteCapability = vi.fn().mockResolvedValue(undefined);
const mockDeleteAcquiredEntity = vi.fn().mockResolvedValue(undefined);
const mockDeleteVendor = vi.fn().mockResolvedValue(undefined);
const mockDeleteInternalTeam = vi.fn().mockResolvedValue(undefined);

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentViewId: 'view-1' }),
}));

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: () => ({
    data: [{ id: 'comp-1', name: 'Component 1' }, { id: 'comp-2', name: 'Component 2' }],
  }),
  useDeleteComponent: () => ({ mutateAsync: mockDeleteComponent }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({
    data: [{ id: 'cap-1', name: 'Capability 1' }],
  }),
  useDeleteCapability: () => ({ mutateAsync: mockDeleteCapability }),
}));

vi.mock('../../views/hooks/useViews', () => ({
  useRemoveComponentFromView: () => ({ mutateAsync: mockRemoveComponent }),
  useRemoveCapabilityFromView: () => ({ mutateAsync: mockRemoveCapability }),
  useRemoveOriginEntityFromView: () => ({ mutateAsync: mockRemoveOriginEntity }),
}));

vi.mock('../../origin-entities/hooks/useAcquiredEntities', () => ({
  useAcquiredEntitiesQuery: () => ({ data: [] }),
  useDeleteAcquiredEntity: () => ({ mutateAsync: mockDeleteAcquiredEntity }),
}));

vi.mock('../../origin-entities/hooks/useVendors', () => ({
  useVendorsQuery: () => ({ data: [] }),
  useDeleteVendor: () => ({ mutateAsync: mockDeleteVendor }),
}));

vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: () => ({ data: [] }),
  useDeleteInternalTeam: () => ({ mutateAsync: mockDeleteInternalTeam }),
}));

vi.mock('../utils/nodeFactory', () => ({
  extractOriginEntityId: (id: string) => id.replace(/^(ae-|vendor-|team-)/, ''),
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
};

function makeNode(overrides: Partial<NodeContextMenu> = {}): NodeContextMenu {
  const nodeId = overrides.nodeId ?? 'comp-1';
  return {
    x: 100,
    y: 200,
    nodeId,
    viewElementId: nodeId,
    nodeName: 'Component 1',
    nodeType: 'component',
    ...overrides,
  };
}

describe('useBulkOperations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('starts with no bulk operation', () => {
    const { result } = renderHook(() => useBulkOperations(), { wrapper: createWrapper() });
    expect(result.current.bulkOperation).toBeNull();
    expect(result.current.isExecuting).toBe(false);
    expect(result.current.result).toBeNull();
  });

  it('sets bulk operation via setBulkOperation', () => {
    const { result } = renderHook(() => useBulkOperations(), { wrapper: createWrapper() });
    const nodes = [makeNode({ nodeId: 'comp-1' }), makeNode({ nodeId: 'comp-2' })];

    act(() => {
      result.current.setBulkOperation({ type: 'removeFromView', nodes });
    });

    expect(result.current.bulkOperation).not.toBeNull();
    expect(result.current.bulkOperation!.type).toBe('removeFromView');
    expect(result.current.bulkOperation!.nodes).toHaveLength(2);
  });

  it('clears bulk operation on cancel', () => {
    const { result } = renderHook(() => useBulkOperations(), { wrapper: createWrapper() });
    const nodes = [makeNode({ nodeId: 'comp-1' }), makeNode({ nodeId: 'comp-2' })];

    act(() => {
      result.current.setBulkOperation({ type: 'removeFromView', nodes });
    });

    act(() => {
      result.current.handleBulkCancel();
    });

    expect(result.current.bulkOperation).toBeNull();
  });

  it('executes bulk remove from view in parallel', async () => {
    const { result } = renderHook(() => useBulkOperations(), { wrapper: createWrapper() });
    const nodes = [
      makeNode({ nodeId: 'comp-1', nodeName: 'Component 1', nodeType: 'component' }),
      makeNode({ nodeId: 'comp-2', nodeName: 'Component 2', nodeType: 'component' }),
    ];

    act(() => {
      result.current.setBulkOperation({ type: 'removeFromView', nodes });
    });

    await act(async () => {
      await result.current.handleBulkConfirm();
    });

    expect(mockRemoveComponent).toHaveBeenCalledTimes(2);
    expect(result.current.bulkOperation).toBeNull();
    expect(result.current.isExecuting).toBe(false);
  });

  it('executes bulk delete from model sequentially', async () => {
    const { result } = renderHook(() => useBulkOperations(), { wrapper: createWrapper() });
    const nodes = [
      makeNode({ nodeId: 'comp-1', nodeName: 'Component 1', nodeType: 'component' }),
      makeNode({ nodeId: 'comp-2', nodeName: 'Component 2', nodeType: 'component' }),
    ];

    act(() => {
      result.current.setBulkOperation({ type: 'deleteFromModel', nodes });
    });

    await act(async () => {
      await result.current.handleBulkConfirm();
    });

    expect(mockDeleteComponent).toHaveBeenCalledTimes(2);
    expect(result.current.bulkOperation).toBeNull();
  });

  it('stops sequential delete on first failure and reports partial result', async () => {
    mockDeleteComponent
      .mockResolvedValueOnce(undefined)
      .mockRejectedValueOnce(new Error('Delete failed'));

    const { result } = renderHook(() => useBulkOperations(), { wrapper: createWrapper() });
    const nodes = [
      makeNode({ nodeId: 'comp-1', nodeName: 'Component 1', nodeType: 'component' }),
      makeNode({ nodeId: 'comp-2', nodeName: 'Component 2', nodeType: 'component' }),
    ];

    act(() => {
      result.current.setBulkOperation({ type: 'deleteFromModel', nodes });
    });

    await act(async () => {
      await result.current.handleBulkConfirm();
    });

    expect(result.current.result).not.toBeNull();
    expect(result.current.result!.succeeded).toEqual(['Component 1']);
    expect(result.current.result!.failed).toHaveLength(1);
    expect(result.current.result!.failed[0].name).toBe('Component 2');
    expect(result.current.bulkOperation).not.toBeNull();
  });
});
