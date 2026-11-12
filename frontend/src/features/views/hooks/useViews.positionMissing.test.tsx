import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import React from 'react';
import toast from 'react-hot-toast';
import { describe, expect, it, vi } from 'vitest';
import type { CapabilityId, ComponentId, ViewId } from '../../../api/types';
import { server } from '../../../test/helpers';
import { useUpdateCapabilityPosition, useUpdateComponentPosition, useUpdateOriginEntityPosition } from './useViews';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

const API_BASE = 'http://localhost:8080';
const viewId = 'view-missing-position' as ViewId;
const position = { x: 3, y: 4 };

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function respondWith(path: string, status: number) {
  server.use(
    http.patch(`${API_BASE}/api/v1/views/${viewId}/${path}/position`, () =>
      status === 404 ? HttpResponse.json({ message: 'not in view' }, { status }) : new HttpResponse(null, { status }),
    ),
  );
}

const hooks = [
  {
    name: 'useUpdateComponentPosition',
    path: 'components/comp-1',
    useHook: useUpdateComponentPosition,
    variables: { viewId, componentId: 'comp-1' as ComponentId, request: position },
  },
  {
    name: 'useUpdateCapabilityPosition',
    path: 'capabilities/cap-1',
    useHook: useUpdateCapabilityPosition,
    variables: { viewId, capabilityId: 'cap-1' as CapabilityId, position },
  },
  {
    name: 'useUpdateOriginEntityPosition',
    path: 'origin-entities/org-1',
    useHook: useUpdateOriginEntityPosition,
    variables: { viewId, originEntityId: 'org-1', position },
  },
];

describe.each(hooks)('$name when the element is no longer in the view', ({ path, useHook, variables }) => {
  it('succeeds without an error toast on a 404', async () => {
    respondWith(path, 404);
    const { result } = renderHook(() => useHook(), { wrapper });

    result.current.mutate(variables as never);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(toast.error).not.toHaveBeenCalled();
  });

  it('reports a 500 as an error', async () => {
    respondWith(path, 500);
    const { result } = renderHook(() => useHook(), { wrapper });

    result.current.mutate(variables as never);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
