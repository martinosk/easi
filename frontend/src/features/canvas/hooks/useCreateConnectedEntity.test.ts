import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCreateConnectedEntity } from './useCreateConnectedEntity';

const mockPost = vi.fn();
const mockPut = vi.fn();
const mockAddComponentToView = vi.fn().mockResolvedValue(undefined);
const mockInvalidateFor = vi.fn();

vi.mock('../../../api/core/httpClient', () => ({
  httpClient: {
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

vi.mock('../../views/hooks/useViewOperations', () => ({
  useViewOperations: () => ({ addComponentToView: mockAddComponentToView }),
}));

vi.mock('../../../lib/invalidateFor', () => ({
  invalidateFor: (...args: unknown[]) => mockInvalidateFor(...args),
}));

vi.mock('../../components/mutationEffects', () => ({
  componentsMutationEffects: { create: () => ['components'] },
}));

vi.mock('../../relations/mutationEffects', () => ({
  relationsMutationEffects: { create: () => ['relations'] },
}));

vi.mock('../../origin-entities/queryKeys', () => ({
  originRelationshipsQueryKeys: { lists: () => 'origin-lists' },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function Wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return QueryClientProvider({ client: qc, children });
}

describe('useCreateConnectedEntity', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPost.mockReset();
    mockPut.mockReset();
  });

  it('calls create component API then create relation API for x-add-relation', async () => {
    mockPost
      .mockResolvedValueOnce({ data: { id: 'new-comp-1', name: 'Test Component' } })
      .mockResolvedValueOnce({ data: {} });

    const { result } = renderHook(
      () => useCreateConnectedEntity('source-1', { x: 100, y: 200 }, 'right'),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createConnectedEntity({
        name: 'Test Component',
        description: 'A test',
        actionType: 'x-add-relation',
        relationType: 'Triggers',
        actionLink: { href: '/api/v1/relations', method: 'POST' },
      });
    });

    expect(mockPost).toHaveBeenCalledTimes(2);
    expect(mockPost).toHaveBeenNthCalledWith(1, '/api/v1/components', {
      name: 'Test Component',
      description: 'A test',
    });
    expect(mockPost).toHaveBeenNthCalledWith(2, '/api/v1/relations', {
      sourceComponentId: 'source-1',
      targetComponentId: 'new-comp-1',
      relationType: 'Triggers',
    });
  });

  it('calls origin PUT endpoint when action is x-set-origin-built-by', async () => {
    mockPost.mockResolvedValueOnce({ data: { id: 'new-comp-2', name: 'Team Component' } });
    mockPut.mockResolvedValueOnce({ data: {} });

    const { result } = renderHook(
      () => useCreateConnectedEntity('source-1', { x: 100, y: 200 }, 'right'),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createConnectedEntity({
        name: 'Team Component',
        description: '',
        actionType: 'x-set-origin-built-by',
        actionLink: { href: '/api/v1/components/source-1/origins/built-by', method: 'POST' },
      });
    });

    expect(mockPost).toHaveBeenCalledTimes(1);
    expect(mockPut).toHaveBeenCalledTimes(1);
    expect(mockPut).toHaveBeenCalledWith('/api/v1/components/source-1/origins/built-by', {
      componentId: 'source-1',
      teamId: 'new-comp-2',
    });
  });

  it('uses the HATEOAS link href for the relation call', async () => {
    const customHref = '/api/v1/custom/relations/endpoint';
    mockPost
      .mockResolvedValueOnce({ data: { id: 'new-comp-3', name: 'Custom' } })
      .mockResolvedValueOnce({ data: {} });

    const { result } = renderHook(
      () => useCreateConnectedEntity('source-1', { x: 100, y: 200 }, 'right'),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createConnectedEntity({
        name: 'Custom',
        description: '',
        actionType: 'x-add-relation',
        relationType: 'Serves',
        actionLink: { href: customHref, method: 'POST' },
      });
    });

    expect(mockPost).toHaveBeenNthCalledWith(2, customHref, expect.any(Object));
  });

  it('adds the new component to the view at the computed position', async () => {
    mockPost
      .mockResolvedValueOnce({ data: { id: 'new-comp-4', name: 'Positioned' } })
      .mockResolvedValueOnce({ data: {} });

    const { result } = renderHook(
      () => useCreateConnectedEntity('source-1', { x: 100, y: 200 }, 'right'),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createConnectedEntity({
        name: 'Positioned',
        description: '',
        actionType: 'x-add-relation',
        relationType: 'Triggers',
        actionLink: { href: '/api/v1/relations', method: 'POST' },
      });
    });

    expect(mockAddComponentToView).toHaveBeenCalledWith('new-comp-4', 350, 200);
  });

  it('invalidates caches after successful creation', async () => {
    mockPost
      .mockResolvedValueOnce({ data: { id: 'new-comp-5', name: 'Cached' } })
      .mockResolvedValueOnce({ data: {} });

    const { result } = renderHook(
      () => useCreateConnectedEntity('source-1', { x: 100, y: 200 }, 'right'),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createConnectedEntity({
        name: 'Cached',
        description: '',
        actionType: 'x-add-relation',
        relationType: 'Triggers',
        actionLink: { href: '/api/v1/relations', method: 'POST' },
      });
    });

    expect(mockInvalidateFor).toHaveBeenCalled();
  });
});
