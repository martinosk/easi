import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import type { ComponentId, ViewId } from '../../../api/types';
import { server } from '../../../test/helpers';
import { viewsApi } from '../api';
import { useAddComponentToView } from './useViews';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

const API_BASE = 'http://localhost:8080';
const viewId = 'view-add-retry' as ViewId;
const componentId = 'comp-retry' as ComponentId;

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function recordAddThenPosition(log: string[]) {
  let addAttempts = 0;
  server.use(
    http.post(`${API_BASE}/api/v1/views/${viewId}/components`, () => {
      addAttempts++;
      log.push(`add-${addAttempts}`);
      return addAttempts === 1 ? new HttpResponse(null, { status: 412 }) : new HttpResponse(null, { status: 201 });
    }),
    http.patch(`${API_BASE}/api/v1/views/${viewId}/components/${componentId}/position`, () => {
      log.push('position');
      return new HttpResponse(null, { status: 204 });
    }),
  );
}

describe('useAddComponentToView retried after a concurrency conflict', () => {
  it('holds a position save until the retried add succeeds', async () => {
    const log: string[] = [];
    recordAddThenPosition(log);
    const { result } = renderHook(() => useAddComponentToView(), { wrapper });

    result.current.mutate({ viewId, request: { componentId, x: 1, y: 2 } });
    await waitFor(() => expect(log).toContain('add-1'));
    await viewsApi.updateComponentPosition(viewId, componentId, { x: 50, y: 60 });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(log).toEqual(['add-1', 'add-2', 'position']);
  });
});
