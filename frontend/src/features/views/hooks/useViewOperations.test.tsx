import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { toViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useViewOperations } from './useViewOperations';

vi.mock('./useCurrentView', () => ({
  useCurrentView: () => ({
    currentView: null,
    currentViewId: useAppStore.getState().currentViewId,
    isLoading: false,
    error: null,
  }),
}));

vi.mock('./useViews', () => ({
  useAddComponentToView: () => ({ mutateAsync: vi.fn() }),
  useRemoveComponentFromView: () => ({ mutateAsync: vi.fn() }),
}));

function Wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

const v1 = toViewId('view-1');
const v2 = toViewId('view-2');

describe('useViewOperations.switchView', () => {
  beforeEach(() => {
    useAppStore.setState({ currentViewId: null, openViewIds: [] });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('opens the view as a tab and sets it as current', () => {
    const { result } = renderHook(() => useViewOperations(), { wrapper: Wrapper });

    act(() => result.current.switchView(v1));

    expect(useAppStore.getState().openViewIds).toEqual([v1]);
    expect(useAppStore.getState().currentViewId).toBe(v1);
  });

  it('is idempotent — calling switchView twice does not duplicate', () => {
    const { result } = renderHook(() => useViewOperations(), { wrapper: Wrapper });

    act(() => result.current.switchView(v1));
    act(() => result.current.switchView(v1));

    expect(useAppStore.getState().openViewIds).toEqual([v1]);
  });

  it('appends additional views to openViewIds', () => {
    const { result } = renderHook(() => useViewOperations(), { wrapper: Wrapper });

    act(() => result.current.switchView(v1));
    act(() => result.current.switchView(v2));

    expect(useAppStore.getState().openViewIds).toEqual([v1, v2]);
    expect(useAppStore.getState().currentViewId).toBe(v2);
  });
});
