import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCreateDynamicView } from './useCreateDynamicView';

const createdView = { id: 'view-99' as const };
const mockCreateViewMutation = {
  mutateAsync: vi.fn().mockResolvedValue(createdView),
};

vi.mock('../../views/hooks/useViews', () => ({
  useCreateView: () => mockCreateViewMutation,
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function Wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('useCreateDynamicView', () => {
  afterEach(() => {
    act(() => {
      useAppStore.setState({
        dynamicEnabled: false,
        dynamicEntities: [],
        dynamicPositions: {},
        dynamicOriginal: null,
        dynamicViewId: null,
      });
    });
    mockCreateViewMutation.mutateAsync.mockClear();
  });

  it('seeds dynamic mode with the source entity and tags it with the new view id', async () => {
    const { result } = renderHook(() => useCreateDynamicView(), { wrapper: Wrapper });

    await act(async () => {
      await result.current.create({ id: 'comp-42', type: 'component' }, 'My System');
    });

    const state = useAppStore.getState();
    expect(state.dynamicEnabled).toBe(true);
    expect(state.dynamicViewId).toBe('view-99');
    expect(state.dynamicEntities).toEqual([{ id: 'comp-42', type: 'component' }]);
    expect(state.dynamicOriginal?.entities).toEqual([{ id: 'comp-42', type: 'component' }]);
  });
});
